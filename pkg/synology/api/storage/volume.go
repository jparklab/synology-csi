/*
 * Copyright 2019 Ji-Young Park(jiyoung.park.dev@gmail.com)
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

package storage

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/jparklab/synology-csi/pkg/synology/core"
	"net/url"
)

const (

	// FSTypeEXT4 is ext4 file system
	FSTypeExt4 = "ext4"
	// FSTypeBtrfs is btrfs file system
	FSTypeBtrfs = "btrfs"
)

/*************************************************************
* Volume Object
* Example Volume
   {
       "volume": {
           "container": "internal",
           "display_name": "Volume 1",
           "eppool_used_byte": "7279022080",
           "fs_type": "btrfs",
           "location": "internal",
           "raid_type": "shr_1",
           "readonly": false,
           "single_volume": false,
           "size_free_byte": "3738201395200",
           "size_total_byte": "7676309151744",
           "status": "normal",
           "volume_id": 1,
           "volume_path": "/volume1"
       }
   }
*/

type Volume struct {
	VolumeId   int    `json:"volume_id"`
	VolumePath string `json:"volume_path"`

	FSType        string `json:"fs_type"`
	SizeFreeByte  string `json:"size_free_byte"`
	SizeTotalByte string `json:"size_total_byte"`

	Status string `json:"status"`
}

/*************************************************************
 * API for Volume
 *************************************************************/
type VolumeAPI interface {
	List() ([]Volume, error)
	Get(volumePath string) (*Volume, error)
}

type volumeAPI struct {
	apiEntry core.APIEntry
}

// NewVolumeAPI creates a VolumeAPI object
func NewVolumeAPI(s core.Session) VolumeAPI {
	entry := core.NewAPIEntry(s, Path, "SYNO.Core.Storage.Volume", "1")

	return &volumeAPI{
		apiEntry: entry,
	}
}

func (v *volumeAPI) List() ([]Volume, error) {
	data, err := v.apiEntry.Get("list", url.Values{
		"limit":    {"-1"},
		"offset":   {"0"},
		"location": {"internal"},
	})

	if err != nil {
		return nil, err
	}

	var volumes []Volume
	if jsonErr := json.Unmarshal(*data["volumes"], &volumes); jsonErr != nil {
		glog.Errorf("Failed to parse volume list: %s(%s)", *data["volumes"], jsonErr)
		return nil, jsonErr
	}

	return volumes, nil
}

func (v *volumeAPI) Get(volumePath string) (*Volume, error) {
	data, err := v.apiEntry.Get("get", url.Values{
		"volume_path": {volumePath},
	})

	if err != nil {
		return nil, err
	}

	var volume Volume
	if jsonErr := json.Unmarshal(*data["volume"], &volume); jsonErr != nil {
		glog.Errorf("Failed to parse volume: %s(%s)", *data["volume"], jsonErr)
		return nil, jsonErr
	}

	return &volume, nil
}
