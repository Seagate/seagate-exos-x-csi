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
	"errors"
	"fmt"
	"strings"

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
	cs commonService
}

type sasStorage struct {
	cs commonService
}

// buildCommonService:
func buildCommonService(config map[string]string, secretMap map[string]string) (commonService, error) {
	commonserv := commonService{}
	if config != nil {
		commonserv.driverVersion = config["driverversion"]
	}
	klog.V(2).Infof("buildCommonService commonservice configuration done.")
	return commonserv, nil
}

//NewStorageNode : To return specific implementation of storage
func NewStorageNode(storageProtocol string, configparams ...map[string]string) (StorageOperations, error) {
	comnserv, err := buildCommonService(configparams[0], configparams[1])
	if err == nil {
		storageProtocol = strings.TrimSpace(storageProtocol)
		klog.V(2).Infof("NewStorageNode for (%s)", storageProtocol)
		if storageProtocol == "fc" {
			return &fcStorage{cs: comnserv}, nil
		} else if storageProtocol == "sas" {
			return &sasStorage{cs: comnserv}, nil
		} else if storageProtocol == "iscsi" {
			return &iscsiStorage{cs: comnserv}, nil
		}
		return nil, errors.New(fmt.Sprintf("Error: Invalid storage protocol (%s)", storageProtocol))
	}
	return nil, err
}
