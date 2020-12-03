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
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"io/ioutil"

	"github.com/golang/glog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"

	"k8s.io/kubernetes/pkg/util/resizefs"
	"k8s.io/utils/exec"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"

	"github.com/jparklab/synology-csi/pkg/synology/api/iscsi"
)

const (
	probeDeviceInterval = 1 * time.Second
	probeDeviceTimeout  = 60 * time.Second
)

type nodeServer struct {
	*csicommon.DefaultNodeServer

	targetAPI iscsi.TargetAPI
	lunAPI    iscsi.LunAPI

	iscsiDrv iscsiDriver
}

func getDevicePath(targetDevPath string) string {
	diskDevPath := "/dev/disk/by-path"

	if entries, err := ioutil.ReadDir(diskDevPath); err == nil {
		for _, f := range entries {
			// example:
			//    ip-192.168.1.196:3260-iscsi-iqn.2000-01.com.synology:JPNAS02.Target-23.cf8d920aa9-lun-1
			glog.V(5).Info(f.Name())
			if strings.Index(f.Name(), targetDevPath) != -1 {
				return strings.Join([]string{diskDevPath, f.Name()}, "/")
			}
		}
	}

	return ""
}

func probeDevice(targetDevPath string) (string, error) {
	ticker := time.NewTicker(probeDeviceInterval)
	defer ticker.Stop()
	timer := time.NewTimer(probeDeviceTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			if devicePath := getDevicePath(targetDevPath); devicePath != "" {
				return devicePath, nil
			}
		case <-timer.C:
			return "", fmt.Errorf("Timed out while waiting for device for %s", targetDevPath)

		}
	}
}

// NodePublishVolume mounts the volume to target path
func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	// TODO: support chap
	// secrets := req.GetNodePublishSecrets()

	targetID, mappingIndex, err := parseVolumeID(volID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	target, err := ns.targetAPI.Get(targetID)
	if err != nil {
		msg := fmt.Sprintf(
			"Unable to find target of ID: %d", targetID)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	// run discovery to add target
	if err = ns.iscsiDrv.discovery(); err != nil {
		msg := fmt.Sprintf("Failed to run ISCSI discovery: %v", err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	hasSession, err := ns.hasSession(target.IQN)
	if err != nil {
		return nil, err
	}

	if hasSession {
		glog.V(5).Infof("Found an existing session for %s", target.IQN)
	} else {
		// login
		if err = ns.iscsiDrv.login(target); err != nil {
			msg := fmt.Sprintf("Failed to run ISCSI login: %v", err)
			glog.V(3).Info(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		defer func() {
			// logout target when we fail to mount
			if err != nil {
				_ = ns.iscsiDrv.logout(target)
			}
		}()
	}

	// find device mapped to the target
	targetDevPath := fmt.Sprintf("%s-lun-%d", target.IQN, mappingIndex)

	devicePath, err := probeDevice(targetDevPath)
	if err != nil {
		msg := fmt.Sprintf("Failed to find device for %s", targetDevPath)
		glog.V(3).Info(msg)
		return nil, errors.New(msg)
	}

	glog.V(5).Infof("Target path: %s", targetPath)

	/*
		notMnt, err := isLikelyNotMountPointAttach(targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	*/
	notMnt := true

	if notMnt {
		exists, err := mount.PathExists(devicePath)
		if !exists || err != nil {
			msg := fmt.Sprintf("Could not find ISCSI device: %s", devicePath)
			glog.V(3).Info(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		// mount device to the target path
		mounter := &mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      utilexec.New(),
		}

		options := []string{"rw"}
		mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
		options = append(options, mountFlags...)

		glog.V(5).Infof(
			"Mounting %s to %s(fstype: %s, options: %v)",
			devicePath, targetPath, fsType, options)
		err = mounter.FormatAndMount(devicePath, targetPath, fsType, options)
		if err != nil {
			msg := fmt.Sprintf(
				"Failed to mount %s to %s(fstype: %s, options: %v): %v",
				devicePath, targetPath, fsType, options, err)
			glog.V(5).Info(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		// TODO(jpark):
		// change owner of the root path:
		// https://github.com/kubernetes/kubernetes/pull/62486
		//	 https://github.com/kubernetes/kubernetes/pull/62486/files
		// https://github.com/kubernetes/kubernetes/issues/66323
		//	https://github.com/kubernetes/kubernetes/pull/67280/files

		glog.V(5).Infof(
			"Mounted %s to %s(fstype: %s, options: %v)",
			devicePath, targetPath, fsType, options)
	} else {
		glog.V(5).Infof("%s is already mounted", targetPath)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	targetID, _, err := parseVolumeID(volID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	target, err := ns.targetAPI.Get(targetID)
	if err != nil {
		msg := fmt.Sprintf(
			"Unable to find target of ID: %d", targetID)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	/*
		notMnt, err := isLikelyNotMountPointDetach(targetPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	*/
	notMnt := false

	if notMnt {
		msg := fmt.Sprintf("Path %s not mounted", targetPath)
		glog.V(3).Info(msg)
		return nil, status.Errorf(codes.NotFound, msg)
	}

	mounter := &mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      utilexec.New(),
	}

	if err = mounter.Unmount(targetPath); err != nil {
		msg := fmt.Sprintf("Failed to unmount %s: %v", targetPath, err)
		glog.V(3).Info(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	// logout target
	// NOTE: we can safely log out because we do not share the target
	//	and we only support targets with a single lun
	if err = ns.iscsiDrv.logout(target); err != nil {
		msg := fmt.Sprintf(
			"Failed to logout(iqn: %s): %v", target.IQN, err)
		glog.V(3).Info(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeStageVolume temporarily mounts the volume to a staging path
func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	// No staging is necessary since we do not share volumes
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	// No staging is necessary since we do not share volumes
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	volID := req.GetVolumeId()
	if volID == "" {
		msg := fmt.Sprintf("Cannot find volume id")
		glog.V(3).Info(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	volumePath := req.GetVolumePath()
	if volumePath == "" {
		msg := fmt.Sprintf("Cannot find volume path")
		glog.V(3).Info(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	if fsType == "" {
		msg := fmt.Sprintf("Cannot detect filesystem type")
		glog.V(3).Info(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	mounter := &mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      utilexec.New(),
	}

	// ex) devicePath = /dev/sdX
	args := []string{"-o", "source", "--noheadings", "--target", volumePath}
	output, err := mounter.Exec.Command("findmnt", args...).CombinedOutput()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot detect device path for volume %s: %v", volumePath, err)
	}
	devicePath := strings.TrimSpace(string(output))

	// ex) /sys/block/sdX/device/rescan is rescan device path
	blockDeviceRescanPath := ""
	parts := strings.Split(devicePath, "/")
	if len(parts) == 3 && strings.HasPrefix(parts[1], "dev") {
		d := filepath.Join("/sys/block", parts[2], "device", "rescan")
		blockDeviceRescanPath, err = filepath.EvalSymlinks(d)
		if err != nil {
			return nil, status.Error(codes.Internal, "")
		}
	} else {
		msg := fmt.Sprintf("device path %s is invalid format", devicePath)
		return nil, status.Error(codes.Internal, msg)
	}
	// write data for triggering to rescan
	err = ioutil.WriteFile(blockDeviceRescanPath, []byte{'1'}, 0666)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// resize file system
	r := resizefs.NewResizeFs(mounter)
	if _, err := r.Resize(devicePath, volumePath); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not resize volume %s to %s: %s", devicePath, volumePath, err.Error())
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}

// Check if session exists for the given IQN
func (ns *nodeServer) hasSession(iqn string) (bool, error) {
	// check if we already have a session
	sessions, err := ns.iscsiDrv.session()
	if err != nil {
		if exiterr, ok := err.(exec.ExitError); ok {
			if exiterr.ExitStatus() == 21 {
				// This is OK -- this means "no sessions"
				return false, nil
			}
		}

		msg := fmt.Sprintf("Unable to list existing sessions: %v", err)
		glog.V(3).Info(msg)
		return false, status.Error(codes.Internal, msg)
	}

	for _, sess := range sessions {
		if sess.IQN == iqn {
			return true, nil
		}
	}

	return false, nil
}

func (ns *nodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	capabilities := []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	}

	caps := make([]*csi.NodeServiceCapability, len(capabilities))
	for i, capability := range capabilities {
		caps[i] = &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: capability,
				},
			},
		}
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}
