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

package driver

import (
	"fmt"

	"github.com/golang/glog"

	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"

	csi "github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/jparklab/synology-csi/pkg/synology/api/iscsi"
	"github.com/jparklab/synology-csi/pkg/synology/api/storage"
	"github.com/jparklab/synology-csi/pkg/synology/core"
	"github.com/jparklab/synology-csi/pkg/synology/options"
)

const (
	// DriverName is the name of csi driver for synology
	DriverName = "csi.synology.com"

	version = "0.2.0"
)

// Driver is top interface to run server
type Driver interface {
	Run()
}

type driver struct {
	csiDriver *csicommon.CSIDriver

	endpoint string

	synologyHost string
	session      core.Session
}

func Login(synoOption *options.SynologyOptions) (*core.Session, string, error) {
	var proto = "http"
	if synoOption.SslVerify {
		proto = "https"
	}

	synoAPIUrl := fmt.Sprintf(
		"%s://%s:%d/webapi", proto,
		synoOption.Host, synoOption.Port)

	glog.V(1).Infof("Use Synology: %s", synoAPIUrl)

	session := core.NewSession(synoAPIUrl, synoOption.SessionName)
	loginResult, err := session.Login(synoOption)

	return &session, loginResult, err
}

// NewDriver creates a Driver object
func NewDriver(nodeID string, endpoint string, synoOption *options.SynologyOptions) (Driver, error) {
	glog.Infof("Driver: %v", DriverName)

	session, _, err := Login(synoOption)
	if err != nil {
		glog.V(3).Infof("Failed to login: %v", err)
		return nil, err
	}

	d := &driver{
		endpoint:     endpoint,
		synologyHost: synoOption.Host,
		session:      *session,
	}

	csiDriver := csicommon.NewCSIDriver(DriverName, version, nodeID)
	csiDriver.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
			csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		})
	csiDriver.AddVolumeCapabilityAccessModes(
		[]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	d.csiDriver = csiDriver

	return d, nil
}

func (d *driver) Run() {
	csicommon.RunControllerandNodePublishServer(d.endpoint, d.csiDriver, newControllerServer(d), newNodeServer(d))
}

func newControllerServer(d *driver) *controllerServer {
	glog.V(3).Infof("Create controller: %v", d.csiDriver)
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.csiDriver),
		lunAPI:                  iscsi.NewLunAPI(d.session),
		targetAPI:               iscsi.NewTargetAPI(d.session),
		volumeAPI:               storage.NewVolumeAPI(d.session),
	}
}

func newNodeServer(d *driver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.csiDriver),
		lunAPI:            iscsi.NewLunAPI(d.session),
		targetAPI:         iscsi.NewTargetAPI(d.session),
		iscsiDrv:          iscsiDriver{synologyHost: d.synologyHost},
	}
}
