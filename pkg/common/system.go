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

package common

import (
	"fmt"
	"strings"
	"unicode"

	"k8s.io/klog"
)

// ValidateName verifies that the string only includes spaces and printable UTF-8 characters except: " , < \
func ValidateName(s string) bool {
	klog.V(2).Infof("ValidateName %q", s)
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == ',' || s[i] == '<' || s[i] == '\\' {
			return false
		} else if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// TranslateName converts the passed in volume name to the translated volume name
func TranslateName(name, prefix string) (string, error) {

	klog.V(2).Infof("TranslateName VolumeNameMaxLength=%d name=[%d]%q prefix=[%d]%q", VolumeNameMaxLength, len(name), name, len(prefix), prefix)
	volumeName := name

	if len(prefix) == 0 {
		// If string is greater than max, truncate it, otherwise return original string
		if len(volumeName) > VolumeNameMaxLength {
			// Skip over 'pvc-'
			if len(volumeName) >= 4 && volumeName[0:4] == "pvc-" {
				volumeName = volumeName[4:]
			}
			// Skip over 'snapshot-'
			if len(volumeName) >= 9 && volumeName[0:9] == "snapshot-" {
				volumeName = volumeName[9:]
			}
			volumeName = strings.ReplaceAll(volumeName, "-", "")
			klog.V(2).Infof("volumeName=[%d]%q", len(volumeName), volumeName)
			if len(volumeName) > VolumeNameMaxLength {
				volumeName = volumeName[:VolumeNameMaxLength]
			}
		}
	} else {
		// Skip over 'pvc-' and remove all dashes
		uuid := volumeName
		if len(volumeName) >= 4 && volumeName[0:4] == "pvc-" {
			uuid = volumeName[4:]
			klog.Infof("TranslateName(pvc): uuid=%q", uuid)
		}
		if len(volumeName) >= 9 && volumeName[0:9] == "snapshot-" {
			uuid = volumeName[9:]
			klog.Infof("TranslateName(snapshot): uuid=%q", uuid)
		}
		uuid = strings.ReplaceAll(uuid, "-", "")

		// Verify that the prefix is the required length, and truncate as needed, add an underscore
		if len(prefix) > VolumePrefixMaxLength {
			prefix = prefix[:VolumePrefixMaxLength]
		}
		prefix = prefix + "_"

		if len(prefix)+len(uuid) > VolumeNameMaxLength {
			truncate := VolumeNameMaxLength - len(prefix)
			volumeName = prefix + uuid[len(uuid)-truncate:]
		} else {
			volumeName = prefix + uuid
		}
	}

	klog.Infof("TranslateName %q[%d], prefix %q[%d], result %q[%d]", name, len(name), prefix, len(prefix), volumeName, len(volumeName))

	return volumeName, nil
}

// VolumeIdGetName: Decode the augmented volume identifier and return the name only
func VolumeIdGetName(volumeId string) (string, error) {
	tokens := strings.Split(volumeId, AugmentKey)

	if len(tokens) > 0 {
		return tokens[0], nil
	} else {
		return "", fmt.Errorf("Unable to retrieve volume name from (%s)", volumeId)
	}
}

// VolumeIdGetStorageProtocol: Decode the augmented volume identifier and return the storage protocol only
func VolumeIdGetStorageProtocol(volumeId string) (string, error) {
	tokens := strings.Split(volumeId, AugmentKey)

	if len(tokens) > 1 {
		return tokens[1], nil
	} else {
		return "", fmt.Errorf("Unable to retrieve storage protocol from (%s)", volumeId)
	}
}

// VolumeIdGetWwn: Decode the augmented volume identifier and return the WWN
func VolumeIdGetWwn(volumeId string) (string, error) {
	tokens := strings.Split(volumeId, AugmentKey)

	if len(tokens) > 2 {
		return tokens[2], nil
	} else {
		return "", fmt.Errorf("Unable to retrieve wwn from (%s)", volumeId)
	}
}

// VolumeIdAugment: Extend the volume name by augmenting it with storage protocol
func VolumeIdAugment(volumename, storageprotocol, wwn string) string {

	volumeId := volumename + AugmentKey + storageprotocol + AugmentKey + wwn
	klog.V(2).Infof("VolumeIdAugment: %s", volumeId)
	return volumeId
}
