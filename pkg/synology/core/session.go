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
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/golang/glog"
	"github.com/google/go-querystring/query"
	"github.com/jparklab/synology-csi/pkg/synology/options"
)

func errorToDesc(code int) string {
	codeList := map[int]string{
		100: "Unknown error",
		101: "Invalid parameter",
		102: "The requested API does not exist",
		103: "The requested method does not exist",
		104: "The requested version does not support the functionality",
		105: "The logged in session does not have permission",
		106: "Session timeout",
		107: "Session interrupted by duplicate login",
	}

	return codeList[code]
}

func loginErrorToDesc(code int) string {
	codeList := map[int]string{
		400: "No such account or incorrect password",
		401: "Account disabled402Permission denied",
		403: "2-step verification code required",
		404: "Failed to authenticate 2-step verification code",
		405: "App portal incorrect.",
		406: "OTP code enforced.",
		407: "Max Tries (if auto blocking is set to true).",
		408: "Password Expired Can not Change.",
		409: "Password Expired.",
		410: "Password must change (when first time use or after reset password by admin).",
		411: "Account Locked (when account max try exceed).",
	}

	result, ok := codeList[code]
	if ok {
		return result
	} else {
		return errorToDesc(code)
	}
}

/************************************************************
 * Session
 ************************************************************/

// Session provides session level functions
type Session interface {
	GetSid() string
	Login(synoOption *options.SynologyOptions) (string, error)
	Logout() error
	Get(path string, params url.Values) (*http.Response, error)
	Post(path string, data url.Values) (*http.Response, error)
}

type securityData struct {
	Timeout int `json:"timeout"`
}

/*

RESPONSE

+----------------+-----------+--------------+-------------------------------------------------------------------------+
| Name           | Value     | Availability | Description                                                             |
+----------------+-----------+--------------+-------------------------------------------------------------------------+
| id             | <string>  | 2 and onward | Session ID, pass this value by HTTP argument "_sid" or Cookie argument. |
| did            | <string>  | 6 and onward | Device id, use to skip OTP checking.                                    |
| synotoken      | <string>  | 3 and onward | If CSRF enabled in DSM, pass this value by HTTP argument "SynoToken"    |
| is_portal_port | <boolean> | 4 and onward | Login through app portal                                                |
+----------------+-----------+--------------+-------------------------------------------------------------------------+

*/
type responseData struct {
	Data  map[string]*json.RawMessage `json:"data"`
	Error struct {
		Code int `json:"code"`
	} `json:"error"`

	Success bool `json:"success"`
}

func (r *responseData) String() string {
	if b, err := json.Marshal(r); err != nil {
		return string(b)
	} else {
		return fmt.Sprintf("%v", err)
	}
}

type session struct {
	sid         string
	baseURL     string
	sessionName string

	options       *options.SynologyOptions
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

func (s *session) prepareArguments() (url.Values, error) {
	return query.Values(s.options)

}

func (s *session) login() (string, error) {
	v, err := s.prepareArguments()
	if err != nil {
		glog.Errorf("Failed parsing URL parameters: %v", err)
		return "", err
	}

	var uri string
	var requestBody []byte
	var method string

	if s.options.LoginApiVersion >= 6 {
		uri = fmt.Sprintf(
			"%s/%s",
			s.baseURL,
			"auth.cgi",
		)
		method = "POST"
		requestBody = []byte(v.Encode())
	} else {
		uri = fmt.Sprintf(
			"%s/%s?%s",
			s.baseURL,
			"auth.cgi",
			v.Encode(),
		)
		method = "GET"
		requestBody = nil
	}

	i := 10
	authResp := responseData{}
	var body []byte

	client := &http.Client{
		CheckRedirect: nil,
	}

	for i > 0 {
		glog.Infof("Logging in via %s", uri)

		if i < 10 {
			time.Sleep(2000 * time.Millisecond)
		}
		i = i - 1

		var reader io.Reader
		if requestBody != nil {
			reader = bytes.NewReader(requestBody)
		}
		req, err := http.NewRequest(method, uri, reader)
		if err != nil {
			glog.Errorf("Failed making a %s request: %v", method, err)
			if i > 0 {
				continue
			} else {
				return "", err
			}
		}

		req.Header.Set("Accept", "*/*")
		if method == "POST" {
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(requestBody)))
		}

		resp, err := client.Do(req)
		if err != nil {
			glog.Errorf("Failed logging in: %v", err)
			if i > 0 {
				continue
			} else {
				return "", err
			}
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				glog.Errorf("Failed closing the body: %v", err)
			}
		}()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Errorf("Failed parsing response data : %v", err)
			if i > 0 {
				continue
			} else {
				return "", err
			}
		}

		authResp = responseData{}
		if err = json.Unmarshal(body, &authResp); err != nil {
			glog.Errorf("Failed to parse login response: %s(%v)", body, err)
			if i > 0 {
				continue
			} else {
				return "", err
			}
		}

		if !authResp.Success {
			code := authResp.Error.Code
			msg := fmt.Sprintf("Failed to login: %d %s: %s", code, loginErrorToDesc(code), body)
			glog.Errorf(msg)
			if i > 0 {
				continue
			} else {
				return "", errors.New(msg)
			}
		}

		break
	}

	if !authResp.Success {
		return "", errors.New("Could not login")
	}

	if err = json.Unmarshal(*authResp.Data["sid"], &s.sid); err != nil {
		glog.Errorf("Failed to parse auth authResp.Data.sid: %s(%v)", authResp, err)
		return "", err

	}

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
	defer func() {
		if err := securityResp.Body.Close(); err != nil {
			glog.Errorf("Failed closing the body: %v", err)
		}
	}()

	securityConf := securityData{}
	if err = json.Unmarshal(body, &securityConf); err != nil {
		glog.Errorf("Failed to parse auth response: %s(%v)", body, err)
		return "", err
	}

	s.timeoutMinute = securityConf.Timeout

	now := time.Now()
	s.lastLoginTime = &now

	glog.Infof("Logged in. Timeout minute: %d", s.timeoutMinute)

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

func (s *session) Login(options *options.SynologyOptions) (string, error) {
	s.options = options

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

	defer func() {
		if err := resp.Body.Close(); err != nil {
			glog.Errorf("Failed closing the body: %v", err)
		}
	}()

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

// Post sends 'POST' request to the endpoint for the method with the parameters
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

	defer func() {
		if err := resp.Body.Close(); err != nil {
			glog.Errorf("Failed closing the body: %v", err)
		}
	}()

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
