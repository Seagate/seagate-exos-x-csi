//
// Copyright (c) 2021 Seagate Technology LLC and/or its Affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// For any questions about this software or licensing,
// please email opensource@seagate.com or cortx-questions@seagate.com.

package storage

import (
	"strings"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog"
)

type StorageOperations interface {
	csi.NodeServer
}

type commonService struct {
	storagePoolIdName map[int64]string
	driverVersion     string
}

type fcStorage struct {
	cs commonService
}

type iscsiStorage struct {
	cs            commonService
	iscsiInfoPath string
}

type sasStorage struct {
	cs commonService
}

// buildCommonService:
func buildCommonService(config map[string]string) (commonService, error) {
	commonserv := commonService{}
	commonserv.driverVersion = config["driverversion"]
	klog.V(2).Infof("buildCommonService commonservice configuration done.")
	return commonserv, nil
}

//NewStorageNode : To return specific implementation of storage
func NewStorageNode(storageProtocol string, config map[string]string) (StorageOperations, error) {
	comnserv, err := buildCommonService(config)
	if err == nil {
		storageProtocol = strings.TrimSpace(storageProtocol)
		klog.V(2).Infof("NewStorageNode for (%s)", storageProtocol)
		if storageProtocol == common.StorageProtocolFC {
			return &fcStorage{cs: comnserv}, nil
		} else if storageProtocol == common.StorageProtocolSAS {
			return &sasStorage{cs: comnserv}, nil
		} else if storageProtocol == common.StorageProtocolISCSI {
			return &iscsiStorage{cs: comnserv, iscsiInfoPath: config["iscsiInfoPath"]}, nil
		} else {
			klog.Warningf("Invalid or no storage protocol specified (%s)", storageProtocol)
			klog.Warningf("Expecting storageProtocol (iscsi, fc, sas, etc) in StorageClass YAML. Default of (%s) used.", common.StorageProtocolISCSI)
			return &iscsiStorage{cs: comnserv, iscsiInfoPath: config["iscsiInfoPath"]}, nil
		}
	}
	return nil, err
}

// ValidateStorageProtocol: Verifies that a correct protocol is chosen or returns a valid default storage protocol.
func ValidateStorageProtocol(storageProtocol string) string {
	if storageProtocol == common.StorageProtocolFC || storageProtocol == common.StorageProtocolISCSI || storageProtocol == common.StorageProtocolSAS {
		return storageProtocol
	} else {
		klog.Warningf("Invalid or no storage protocol specified (%s)", storageProtocol)
		klog.Warningf("Expecting storageProtocol (iscsi, fc, sas, etc) in StorageClass YAML. Default of (%s) used.", common.StorageProtocolISCSI)
		return common.StorageProtocolISCSI
	}
}
