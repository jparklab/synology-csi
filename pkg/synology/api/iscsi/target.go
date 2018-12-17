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
	"errors"
	"fmt"
	"strconv"

	"encoding/json"
	"net/url"

	"github.com/golang/glog"
	"github.com/jparklab/synology-csi/pkg/synology/core"
)

const (
	// TargetAuthTypeNone is for no auth
	TargetAuthTypeNone = 0
	// TargetAuthTypeSingleChap is for single chap
	TargetAuthTypeSingleChap = 1
	// TargetAuthTypeMutualChap is for mutual chap
	TargetAuthTypeMutualChap = 2
)

var (
	// AdditionalTargetFields contains list of additional fields to query
	AdditionalTargetFields = []string {
		"acls",
		"connected_sessions",
		"mapped_lun",
		"status",
	}
)

/*************************************************************
 * Target Object
 * 
 * An example target object(TODO: move to test)
 * 
 *
	{
	"acls": [
		{
		"iqn": "iqn.2000-01.com.synology:default.acl",
		"permission": "rw"
		}
	],
	"auth_type": 0,
	"connected_sessions": [],
	"has_data_checksum": false,
	"has_header_checksum": false,
	"iqn": "iqn.2000-01.com.synology:kubernetes-target-3",
	"is_enabled": true,
	"mapped_luns": [G
		{
		"lun_uuid": "7a9121bc-ef10-4fbe-80ef-2d552b94aaaa",
		"mapping_index": 1
		}
	],
	"mapping_index": -1,
	"max_recv_seg_bytes": 262144,
	"max_send_seg_bytes": 262144,
	"max_sessions": 1,
	"mutual_password": "",
	"mutual_user": "",
	"name": "target-3",
	"network_portals": [
		{
		"interface_name": "all",
		"ip": "",
		"port": 3260
		}
	],
	"password": "",
	"status": "online",
	"target_id": 12,
	"user": ""
	}
 *
 *************************************************************/

// Target represents a target object
type Target struct {

	Name string `json:"name"`
	IQN string `json:"iqn"`
	TargetID int `json:"target_id"`

	AuthType int `json:"auth_type"`
	User string `json:"user"`
	Password string `json:"password"`
	MutualUser string `json:"mutual_user"`
	MutualPassword string `json:"mutual_password"`


	MappedLuns []struct {
		LunUUID string `json:"lun_uuid"`
		MappingIndex int `json:"mapping_index"`
	} `json:"mapped_luns"`

	MaxSessions int `json:"max_sessions"`
	IsEnabled bool `json:"is_enabled"`
	Status string `json:"status"`
}

 
/*************************************************************
 * API for Target
 *************************************************************/

// TargetAPI defines Target object
type TargetAPI interface {
	List() ([]Target, error)
	Get(id int) (*Target, error)
	Create(
		name string,		// name of the target
		iqn string,			// iqn
		authType int,		// see TargetAuthType
		user string,		// username, can be nil when authType is 0
		password string,	// password, can be nil when authType is 0
	) (*Target, error)
	Delete(id int) error

	MapLun(targetID int, lunUUIDs []string) error
	UnmapLun(targetID int, lunUUIDs []string) error
}


type targetAPI struct {
	apiEntry core.APIEntry
}

// NewTargetAPI creates a LunAPI object
func NewTargetAPI(s core.Session) TargetAPI {
	entry := core.NewAPIEntry(s, Path, "SYNO.Core.ISCSI.Target", "1")

	return &targetAPI{
		apiEntry: entry,
	}
}


func (t *targetAPI) List() ([]Target, error) {
	additional, _ := json.Marshal(AdditionalTargetFields)

	data, err := t.apiEntry.Get("list", url.Values{
		"additional": { string(additional) },
	})
	if err != nil {
		return nil, err
	}

	var targets []Target 
	if jsonErr := json.Unmarshal(*data["targets"], &targets); jsonErr != nil {
		glog.Errorf("Failed to parse target list: %s(%s)", *data["targets"], jsonErr)
		return nil, jsonErr
	}

	return targets, nil
}

func (t *targetAPI) Get(id int) (*Target, error) {
	additional, _ := json.Marshal(AdditionalTargetFields)

	data, err := t.apiEntry.Get("get", url.Values{
		"additional": { string(additional) },
		"target_id": { fmt.Sprintf("\"%d\"", id) },
	})
	if err != nil {
		return nil, err
	}

	var target Target
	if jsonErr := json.Unmarshal(*data["target"], &target); jsonErr != nil {
		glog.Errorf("Failed to parse target: %s(%s)", *data["target"], jsonErr)
		return nil, jsonErr
	}

	return &target, nil
}

func (t *targetAPI) Create(
	name string,
	iqn string,
	authType int,
	user string,
	password string,
) (*Target, error) {

	params := url.Values{
		"name": { name },
		"iqn": { iqn },
		"auth_type": { fmt.Sprintf("%d", authType) },
	}

	if authType == TargetAuthTypeNone {
		params.Set("chap", "false")
	} else {
		params.Set("chap", "true")
		// NOTE(jpark):
		//	synology supports mutual chap, but we do not support here.
		params.Set("mutual_chap", "false")
		params.Set("user", user)
		params.Set("password", password)
		params.Set("password_confirm", password)
	}

	data, err := t.apiEntry.Post("create", params)
	if err != nil {
		return nil, err
	}

	targetIDStr := string(*data["target_id"])
	fmt.Printf("Created TargetID: %s", targetIDStr)

	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid target ID: %s", targetIDStr))
	}

	return t.Get(targetID)
}

func (t *targetAPI) Delete(id int) error {
	_, err := t.apiEntry.Post("delete", url.Values{ 
		"target_id": { fmt.Sprintf("\"%d\"", id) },
	})

	return err
}

func (t *targetAPI) MapLun(targetID int, lunUUIDs []string) error {
	_, err := t.apiEntry.Post("map_lun", url.Values{
		"target_id": { fmt.Sprintf("\"%d\"", targetID) },
		"lun_uuids": { fmt.Sprintf("[\"%s\"]", lunUUIDs[0]) },
	})

	return err
}
func (t *targetAPI) UnmapLun(targetID int, lunUUIDs []string) error {
	encodedUUIDs, _ := json.Marshal(lunUUIDs)

	_, err := t.apiEntry.Post("unmap_lun", url.Values{
		"target_id": { fmt.Sprintf("\"%d\"", targetID) },
		"lun_uuids": { string(encodedUUIDs) },
	})

	return err
}