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

	volumeID := name

	if len(prefix) == 0 {
		// If string is greater than max, truncate it, otherwise return original string
		if len(volumeID) > VolumeNameMaxLength {
			// Skip over 'pvc-'
			if len(volumeID) >= 4 && volumeID[0:4] == "pvc-" {
				volumeID = volumeID[4:]
			}
			// Skip over 'snapshot-'
			if len(volumeID) >= 9 && volumeID[0:9] == "snapshot-" {
				volumeID = volumeID[9:]
			}
			volumeID = strings.ReplaceAll(volumeID, "-", "")
			volumeID = volumeID[:VolumeNameMaxLength]
		}
	} else {
		// Skip over 'pvc-' and remove all dashes
		uuid := volumeID
		if len(volumeID) >= 4 && volumeID[0:4] == "pvc-" {
			uuid = volumeID[4:]
			klog.Infof("TranslateName(pvc): uuid=%q", uuid)
		}
		if len(volumeID) >= 9 && volumeID[0:9] == "snapshot-" {
			uuid = volumeID[9:]
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
			volumeID = prefix + uuid[len(uuid)-truncate:]
		} else {
			volumeID = prefix + uuid
		}
	}

	klog.Infof("TranslateName %q[%d], prefix %q[%d], result %q[%d]", name, len(name), prefix, len(prefix), volumeID, len(volumeID))

	return volumeID, nil
}
