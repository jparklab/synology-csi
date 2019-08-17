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
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/url"
	"testing"
)

type testApiEntry struct {
    mock.Mock
}

func (m *testApiEntry) Get(method string, params url.Values) (map[string]*json.RawMessage, error) {
    args := m.Called(method, params)

    if args.Get(0) == nil {
		return nil, args.Error(1)
	} else {
		return args.Get(0).(map[string]*json.RawMessage), nil
	}
}

func (m *testApiEntry) Post(method string, params url.Values) (map[string]*json.RawMessage, error) {
	args := m.Called(method, params)

	return args.Get(0).(map[string]*json.RawMessage), args.Error(1)
}

/************************************************************
 * Tests
 ************************************************************/
func TestListVolumes(t *testing.T) {
    entry := testApiEntry{}
    encodedVol := []byte(`[
		{
			"atime_checked": true,
			"atime_opt": "relatime",
			"container": "internal",
			"crashed": false,
			"description": "Located on Storage Pool 1, SHR",
			"display_name": "Volume 1",
			"fs_type": "btrfs",
			"location": "internal",
			"pool_path": "reuse_1",
			"raid_type": "shr_1",
			"readonly": false,
			"single_volume": false,
			"size_free_byte": "3738590121984",
			"size_total_byte": "7676309151744",
			"status": "normal",
			"volume_id": 1,
			"volume_path": "/volume1"
		},
		{
			"atime_checked": true,
			"atime_opt": "relatime",
			"container": "internal",
			"crashed": false,
			"description": "Application Data(SSD)",
			"display_name": "Volume 2",
			"fs_type": "btrfs",
			"location": "internal",
			"pool_path": "reuse_2",
			"raid_type": "raid_1",
			"readonly": false,
			"single_volume": true,
			"size_free_byte": "449451450368",
			"size_total_byte": "475363291136",
			"status": "normal",
			"volume_id": 2,
			"volume_path": "/volume2"
		}]`)
	data := map[string]*json.RawMessage{ "volumes": (*json.RawMessage)(&encodedVol) }

	entry.On("Get", "list", mock.Anything).Return(data, nil)

    api := &volumeAPI{
        apiEntry: &entry,
    }

    volumes, err := api.List()

    assert.NoError(t, err)
    assert.Equal(t, len(volumes), 2)
    assert.Equal(t, volumes[0].VolumePath, "/volume1")
	assert.Equal(t, volumes[1].VolumePath, "/volume2")
}

func TestGetVolume(t *testing.T) {
	entry := testApiEntry{}
	encodedVol := []byte(`{
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
	}`)

	vol1Data := map[string]*json.RawMessage{
		"volume": (*json.RawMessage)(&encodedVol),
	}

	notFoundError := errors.New("volume not found")

	entry.
		On("Get", "get", url.Values{"volume_path":[]string{"/volume1"}}).Return(vol1Data, nil).
		On("Get", "get", url.Values{"volume_path":[]string{"/volume2"}}).Return(nil, notFoundError)

	api := &volumeAPI{
		apiEntry: &entry,
	}

	// Test if Get returns a volume
	vol1, err := api.Get("/volume1")
	assert.NoError(t, err)
	assert.NotNil(t, vol1)
	assert.Equal(t, vol1.VolumePath, "/volume1")

	// Test if Get returns err when no volume found
	vol2, err := api.Get("/volume2")
	assert.Error(t, err)
	assert.Nil(t, vol2)
}