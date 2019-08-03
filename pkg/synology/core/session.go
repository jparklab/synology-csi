/*
 * Copyright 2018 Ji-Young Park(jiyoung.park.dev@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"errors"
	"fmt"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/golang/glog"

	"github.com/google/go-querystring/query"
)

func errorToDesc(code int) string {
	errorToDesc := map[int]string{
		100: "Unknown error",
		101: "Invalid parameter",
		102: "The requested API does not exist",
		103: "The requested method does not exist",
		104: "The requested version does not support the functionality",
		105: "The logged in session does not have permission",
		106: "Session timeout",
		107: "Session interrupted by duplicate login",
	}

	return errorToDesc[code]
}

/************************************************************
 * Session
 ************************************************************/

// Session provides session level functions
type Session interface {
	GetSid() string
	Login(account string, password string) (string, error)
	Logout() error
	Get(path string, params url.Values) (*http.Response, error)
	Post(path string, data url.Values) (*http.Response, error)
}

type authOption struct {
	API     string `url:"api"`
	Version int    `url:"version"`
	Method  string `url:"method"`
	Account string `url:"account"`
	Passwd  string `url:"passwd"`
	Session string `url:"session"`
	Format  string `url:"format"` // "cookie" or "sid"
}

type securityData struct {
	Timeout int `json:"timeout"`
}

type responseData struct {
	Data  map[string]*json.RawMessage `json:"data"`
	Error struct {
		Code int `json:"code"`
	} `json:"error"`

	Success bool `json:"success"`
}

type session struct {
	sid         string
	baseURL     string
	sessionName string

	account  string
	password string

	timeoutMinute int
	lastLoginTime *time.Time
}

// NewSession creates a new Session object
func NewSession(baseURL string, sessionName string) Session {
	return &session{
		baseURL:     baseURL,
		sessionName: sessionName,
	}
}

func (s *session) GetSid() string {
	return s.sid
}

func (s *session) login() (string, error) {
	params := authOption{
		API:     "SYNO.API.Auth",
		Version: 2,
		Method:  "login",
		Account: s.account,
		Passwd:  s.password,
		Session: s.sessionName,
		Format:  "sid",
	}

	v, _ := query.Values(params)
	uri := fmt.Sprintf(
		"%s/%s?%s",
		s.baseURL,
		"auth.cgi",
		v.Encode(),
	)

	glog.V(5).Infof("Logging in via %s", uri)

	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	authResp := responseData{}
	if err = json.Unmarshal(body, &authResp); err != nil {
		glog.Errorf("Failed to parse login response: %s(%v)", body, err)
		return "", err
	}

	if !authResp.Success {
		code := authResp.Error.Code
		msg := fmt.Sprintf("Failed to login: %s(%d)", errorToDesc(code), code)
		return "", errors.New(msg)
	}

	json.Unmarshal(*authResp.Data["sid"], &s.sid)

	// get login timeout
	securityParams := url.Values{
		"_sid":    {s.sid},
		"api":     {"SYNO.Core.Security.DSM"},
		"version": {"1"},
		"method":  {"get"},
	}

	urlObj, _ := url.Parse(fmt.Sprintf("%s/auth.cgi", s.baseURL))
	urlObj.RawQuery = securityParams.Encode()

	securityResp, err := http.Get(urlObj.String())
	if err != nil {
		return "", errors.New("Failed to get security config")
	}

	body, err = ioutil.ReadAll(securityResp.Body)
	defer securityResp.Body.Close()

	securityConf := securityData{}
	if err = json.Unmarshal(body, &securityConf); err != nil {
		glog.Errorf("Failed to parse auth response: %s(%v)", body, err)
		return "", err
	}

	s.timeoutMinute = securityConf.Timeout

	now := time.Now()
	s.lastLoginTime = &now

	return s.sid, nil
}

func (s *session) ensureLoggedIn() error {
	if s.lastLoginTime == nil {
		return errors.New("Session has not been logged in yet")
	}

	minuteSinceLastLogin := time.Since(*s.lastLoginTime)
	if int(minuteSinceLastLogin.Minutes()) < s.timeoutMinute-1 {
		return nil
	}

	// re-login
	_, err := s.login()
	return err
}

func (s *session) Login(account string, password string) (string, error) {
	s.account = account
	s.password = password

	return s.login()
}

func (s *session) Logout() error {

	params := url.Values{
		"_sid":    {s.sid},
		"session": {s.sessionName},
	}

	urlObj, _ := url.Parse(fmt.Sprintf("%s/auth.cgi", s.baseURL))
	urlObj.RawQuery = params.Encode()

	_, err := http.Get(urlObj.String())

	return err
}

func (s *session) Get(path string, params url.Values) (*http.Response, error) {
	if err := s.ensureLoggedIn(); err != nil {
		return nil, err
	}

	urlObj, _ := url.Parse(fmt.Sprintf("%s/%s", s.baseURL, path))
	urlObj.RawQuery = params.Encode()

	glog.V(8).Infof("Querying %s\n", urlObj.String())

	return http.Get(urlObj.String())
}

func (s *session) Post(path string, data url.Values) (*http.Response, error) {
	if err := s.ensureLoggedIn(); err != nil {
		return nil, err
	}

	targetURL := fmt.Sprintf("%s/%s", s.baseURL, path)

	glog.V(8).Infof("Postting %s: %#v\n", targetURL, data)
	return http.PostForm(targetURL, data)
}

/************************************************************
 * API entry
 ************************************************************/

// APIEntry provides functions for an endpoint
type APIEntry interface {
	Get(method string, params url.Values) (map[string]*json.RawMessage, error)
	Post(method string, data url.Values) (map[string]*json.RawMessage, error)
}

type apiEntry struct {
	session Session
	path    string
	api     string
	version string
}

// NewAPIEntry creates an APIEntry object
func NewAPIEntry(s Session, path string, api string, version string) APIEntry {
	return &apiEntry{
		session: s,
		path:    path,
		api:     api,
		version: version,
	}
}

// Get sends 'GET' request to the endpoint for the method with the parameters
// It returns value of 'data' field when the request succeeds
func (e *apiEntry) Get(method string, params url.Values) (map[string]*json.RawMessage, error) {
	params.Add("api", e.api)
	params.Add("version", e.version)
	params.Add("method", method)
	params.Add("_sid", e.session.GetSid())

	resp, err := e.session.Get(e.path, params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var data responseData
	if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
		glog.V(3).Infof("Failed to parse response: %s", body)
		return nil, jsonErr
	}

	if !data.Success {
		code := data.Error.Code
		msg := fmt.Sprintf("Failed to %s: %s(%d)", method, errorToDesc(code), code)
		return nil, errors.New(msg)
	}

	return data.Data, nil
}

// Get sends 'POST' request to the endpoint for the method with the parameters
// It returns value of 'data' field when the request succeeds, or nil if
// the request fails or response does not contain data
func (e *apiEntry) Post(method string, params url.Values) (map[string]*json.RawMessage, error) {
	params.Add("api", e.api)
	params.Add("version", e.version)
	params.Add("method", method)
	params.Add("_sid", e.session.GetSid())

	resp, err := e.session.Post(e.path, params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var data responseData
	if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
		glog.V(3).Infof("Failed to parse: %s", body)
		return nil, jsonErr
	}

	if !data.Success {
		code := data.Error.Code
		msg := fmt.Sprintf("Failed to %s: %s(%d)", method, errorToDesc(code), code)
		return nil, errors.New(msg)
	}

	return data.Data, nil
}
