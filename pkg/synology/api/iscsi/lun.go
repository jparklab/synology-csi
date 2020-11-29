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

package iscsi

import (
	"fmt"
	"strings"

	"github.com/golang/glog"

	"encoding/json"
	"net/url"

	"github.com/jparklab/synology-csi/pkg/synology/core"
)

const (

	// NOTE(jpark)
	//	following types are extracted from the captured request
	//  using google chrome

	LunTypeBlock           = "BLOCK"
	LunTypeFile            = "FILE"
	LunTypeThin            = "THIN"
	LunTypeAdv             = "ADV"
	LunTypeSink            = "SINK"
	LunTypeCinder          = "CINDER"
	LunTypeCinderBLUN      = "CINDER_BLUN"
	LunTypeCinderBLUNThick = "CINDER_BLUN_THICK"

	// LunTypeBlun is used for thin provision
	// This type is mapped to type 263
	LunTypeBlun = "BLUN"
	// LunTypeBlunThick is used for no thin provision
	// This type is mapped to type 259
	LunTypeBlunThick     = "BLUN_THICK"
	LunTypeBlunSink      = "BLUN_SINK"
	LunTypeBlunThickSink = "BLUN_THICK_SINK"
)

var (
	// AdditionalLunFields contains list of additional fields to query
	AdditionalLunFields = []string{
		"is_action_locked",
		"is_mapped",
		"extent_size",
		"allocated_size",
		"status",
		"allow_bkpobj",
		"flashcache_status",
		"family_config",
		"sync_progress",
	}
)

/*************************************************************
 * LUN Object
 * Example LUN
 {
	"dev_attribs": [
	  {
		"dev_attrib": "emulate_3pc",
		"enable": 1
	  },
	  {
		"dev_attrib": "emulate_tpws",
		"enable": 1
	  },
	  {
		"dev_attrib": "emulate_caw",
		"enable": 1
	  },
	  {
		"dev_attrib": "emulate_tpu",
		"enable": 0
	  }
	],
	"location": "/volume3",
	"lun_id": 2,
	"name": "kube-lun-1",
	"restored_time": 0,
	"size": 53687091200,
	"type": 263,
	"uuid": "fd993a34-15ba-44e6-a60c-62d17a3430c8",
	"vpd_unit_sn": "fd993a34-15ba-44e6-a60c-62d17a3430c8"
  }
*/

type Lun struct {
	Location string      `json:"location"`
	LunID    int         `json:"lun_id"`
	Name     string      `json:"name"`
	Size     int64       `json:"size"`
	Type     interface{} `json:"type"` // type can be either int or string
	UUID     string      `json:"uuid"`

	IsMapped bool   `json:"is_mapped"`
	Status   string `json:"status"`
}

/*************************************************************
 * API for LUN
 *************************************************************/

type LunAPI interface {
	List() ([]Lun, error)
	Get(id string) (*Lun, error)
	Create(
		name string, // name of the volume
		location string, // location(e.g. /volume1)
		size int64, // size of the volume(in bytes)
		volType string, // type of the volume, see LunType for available types
	) (*Lun, error)
	Delete(id string) error
	Update(
		id string,
		size int64,
	) error
}

type lunAPI struct {
	apiEntry core.APIEntry
}

// NewLunAPI creates a LunAPI object
func NewLunAPI(s core.Session) LunAPI {
	entry := core.NewAPIEntry(s, Path, "SYNO.Core.ISCSI.LUN", "1")

	return &lunAPI{
		apiEntry: entry,
	}
}

func (l *lunAPI) List() ([]Lun, error) {
	additional, _ := json.Marshal(AdditionalLunFields)

	data, err := l.apiEntry.Get("list", url.Values{
		"additional": {string(additional)},
	})
	if err != nil {
		return nil, err
	}

	var luns []Lun
	if jsonLunErr := json.Unmarshal(*data["luns"], &luns); jsonLunErr != nil {
		glog.Errorf("Failed to parse Lun list: %s(%s)", *data["luns"], jsonLunErr)
		return nil, jsonLunErr
	}

	return luns, nil
}

// Get finds lun for the given ID(either UUID or name of the LUN)
func (l *lunAPI) Get(id string) (*Lun, error) {
	additional, _ := json.Marshal(AdditionalLunFields)

	data, err := l.apiEntry.Get("get", url.Values{
		"uuid":       {fmt.Sprintf("\"%s\"", id)},
		"additional": {string(additional)},
	})
	if err != nil {
		return nil, err
	}

	var lun Lun
	if jsonLunErr := json.Unmarshal(*data["lun"], &lun); jsonLunErr != nil {
		glog.Errorf("Failed to parse Lun: %s(%s)", *data["lun"], jsonLunErr)
		return nil, jsonLunErr
	}

	return &lun, nil
}

func (l *lunAPI) Create(
	name string,
	location string,
	size int64,
	volType string,
) (*Lun, error) {
	data, err := l.apiEntry.Post("create", url.Values{
		"name":     {name},
		"location": {location},
		"type":     {volType},
		"size":     {fmt.Sprintf("%d", size)},
	})

	if err != nil {
		return nil, err
	}

	uuid := string(*data["uuid"])
	// uuid can be quoted
	uuid = strings.Trim(uuid, "\"")

	glog.V(5).Infof("Created a LUN: %s", uuid)

	return l.Get(uuid)
}

func (l *lunAPI) Delete(id string) error {
	_, err := l.apiEntry.Post("delete", url.Values{
		"uuid": {fmt.Sprintf("\"%s\"", id)},
	})

	return err
}

func (l *lunAPI) Update(
	id string,
	size int64,
) error {
	_, err := l.apiEntry.Post("set", url.Values{
		"uuid":     {fmt.Sprintf("\"%s\"", id)},
		"new_size": {fmt.Sprintf("%d", size)},
	})

	glog.V(5).Infof("Updated a LUN: %s", id)

	return err
}
