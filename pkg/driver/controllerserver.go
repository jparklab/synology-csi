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
	"strings"

	"github.com/golang/glog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pborman/uuid"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"

	"github.com/jparklab/synology-csi/pkg/synology/api/iscsi"
)

const (
	defaultVolumeSize = int64(1 * 1024 * 1024 * 1024)
	defaultLocation   = "/volume1"
	defaultVolumeType = iscsi.LunTypeBlun

	targetNamePrefix = "kube-csi"
	lunNamePrefix    = "kube-csi"

	iqnPrefix = "iqn.2000-01.com.synology:kube-csi"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
	targetAPI iscsi.TargetAPI
	lunAPI    iscsi.LunAPI
}

// CreateVolume creates a LUN and a target for a volume
func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	// Volume name
	volName := req.GetName()
	if len(volName) == 0 {
		volName = uuid.NewUUID().String()
	}

	// Volume size
	volSizeByte := defaultVolumeSize
	if req.GetCapacityRange() != nil {
		volSizeByte = int64(req.GetCapacityRange().GetRequiredBytes())
	}

	//
	// Create volumes
	//
	params := req.GetParameters()
	location, present := params["location"]
	if !present {
		location = defaultLocation
	}

	volType, present := params["type"]
	if !present {
		volType = defaultVolumeType
	}

	lunName := fmt.Sprintf("%s-%s", lunNamePrefix, volName)
	targetName := fmt.Sprintf("%s-%s", targetNamePrefix, volName)
	targetIQN := fmt.Sprintf("%s-%s", iqnPrefix, volName)

	// check if lun already exists
	lun, err := cs.lunAPI.Get(lunName)
	if lun != nil {
		return nil, status.Errorf(
			codes.AlreadyExists,
			"Volume %s already exists, found LUN %s", volName, lunName,
		)
	}

	// create a lun
	newLun, err := cs.lunAPI.Create(
		lunName,
		location,
		volSizeByte,
		volType,
	)

	if err != nil {
		msg := fmt.Sprintf(
			"Failed to create a LUN(name: %s, location: %s, size: %d, type: %s): %v",
			lunName, location, volSizeByte, volType, err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("LUN %s(%s) created", lunName, newLun.UUID)

	// create a target
	var newTarget *iscsi.Target

	secrets := req.GetSecrets()
	user, present := secrets["user"]
	if present {
		password, present := secrets["password"]
		if !present {
			glog.V(3).Info("Password is required to provide chap authentication")
			return nil, status.Error(codes.InvalidArgument, "Password is missing")
		}
		newTarget, err = cs.targetAPI.Create(
			targetName,
			targetIQN,
			iscsi.TargetAuthTypeNone,
			user, password,
		)

	} else {
		// create a target
		newTarget, err = cs.targetAPI.Create(
			targetName,
			targetIQN,
			iscsi.TargetAuthTypeNone,
			"", "",
		)
	}

	if err != nil {
		msg := fmt.Sprintf(
			"Failed to create target(name: %s, iqn: %s): %v",
			targetName, targetIQN, err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("Target %s(ID: %d) created", targetName, newTarget.TargetID)

	// map lun
	err = cs.targetAPI.MapLun(
		newTarget.TargetID, []string{newLun.UUID})
	if err != nil {
		msg := fmt.Sprintf(
			"Failed to map LUN %s(%s) to target %s(%d): %v",
			newLun.Name, newLun.UUID, newTarget.Name, newTarget.TargetID, err)
		glog.V(5).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("Mapped LUN %s(%s) to target %s(ID: %d)",
		newLun.Name, newLun.UUID, newTarget.Name, newTarget.TargetID)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      makeVolumeID(newTarget.TargetID, 1),
			CapacityBytes: volSizeByte,
			VolumeContext: map[string]string{
				"targetID":     fmt.Sprintf("%d", newTarget.TargetID),
				"iqn":          newTarget.IQN,
				"mappingIndex": "1",
			},
		},
	}, nil
}

// DeleteVolume deletes the LUN and the target created for the volume
func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	volID := req.GetVolumeId()
	targetID, mappingIndex, err := parseVolumeID(volID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	target, err := cs.targetAPI.Get(targetID)
	if err != nil {
		msg := fmt.Sprintf(
			"Unable to find target of ID(%d): %v", targetID, err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	if len(target.MappedLuns) < mappingIndex {
		msg := fmt.Sprintf("Target %s(%d) does not have mapping for index %d",
			target.Name, target.TargetID, mappingIndex)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	mapping := target.MappedLuns[mappingIndex-1]
	lun, err := cs.lunAPI.Get(mapping.LunUUID)
	if err != nil {
		msg := fmt.Sprintf(
			"Unable to find LUN of UUID: %s(mapped to target %s(%d))",
			mapping.LunUUID, target.Name, target.TargetID)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	// unmap lun
	err = cs.targetAPI.UnmapLun(target.TargetID, []string{lun.UUID})
	if err != nil {
		msg := fmt.Sprintf(
			"Failed to unmap LUN %s(%s) to target %s(%d): %v",
			lun.Name, lun.UUID, target.Name, target.TargetID, err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	glog.V(5).Infof("Unmapped LUN %s(%s) to target %s(ID: %d)",
		lun.Name, lun.UUID, target.Name, target.TargetID)

	// delete target
	err = cs.targetAPI.Delete(target.TargetID)
	if err != nil {
		msg := fmt.Sprintf(
			"Failed to delete target %s(%d): %v",
			target.Name, target.TargetID, err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	glog.V(5).Infof("Deleted target %s(%d)",
		target.Name, target.TargetID)

	// delete lun
	err = cs.lunAPI.Delete(lun.UUID)
	if err != nil {
		msg := fmt.Sprintf(
			"Failed to delete lun %s(%s): %v",
			lun.Name, lun.UUID, err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	glog.V(5).Infof("Deleted lun %s(%s)",
		lun.Name, lun.UUID)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	// nothing needs to be done
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (cs *controllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	targets, err := cs.targetAPI.List()
	if err != nil {
		msg := fmt.Sprintf("Failed to list targets: %v", err)
		glog.V(3).Info(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	var entries []*csi.ListVolumesResponse_Entry
	for _, t := range targets {

		if !strings.HasPrefix(t.Name, targetNamePrefix) {
			// I was not able to find a good way to flag volumes created by csi
			// other than using prefix..
			continue
		}

		for _, mapping := range t.MappedLuns {
			lun, err := cs.lunAPI.Get(mapping.LunUUID)
			if err != nil {
				msg := fmt.Sprintf("Failed to get LUN(%s): %v", mapping.LunUUID, err)
				glog.V(3).Info(msg)
				return nil, status.Error(codes.Internal, msg)

			}
			if lun == nil {
				continue
			}

			entry := csi.ListVolumesResponse_Entry{
				Volume: &csi.Volume{
					VolumeId:      fmt.Sprintf("%d.%d", t.TargetID, mapping.MappingIndex),
					CapacityBytes: lun.Size,
					VolumeContext: map[string]string{
						"targetID":     fmt.Sprintf("%d", t.TargetID),
						"iqn":          t.IQN,
						"mappingIndex": fmt.Sprintf("%d", mapping.MappingIndex),
					},
				},
			}

			entries = append(entries, &entry)
		}
	}

	return &csi.ListVolumesResponse{
		Entries: entries,
	}, nil
}
