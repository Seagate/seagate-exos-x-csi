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

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

// ValidateVolumeName verifies that the string only includes spaces and printable UTF-8 characters except: " , < \
func ValidateVolumeName(s string) bool {
	klog.Infof("ValidateVolumeName %q", s)
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == ',' || s[i] == '<' || s[i] == '\\' {
			return false
		} else if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// TranslateVolumeName converts the passed in volume name to the translated volume name
func TranslateVolumeName(req *csi.CreateVolumeRequest) (string, error) {

	if req.GetName() == "" {
		return "", status.Error(codes.InvalidArgument, "cannot create volume with an empty name")
	}

	originalVolumeID := req.GetName()
	volumeID := originalVolumeID
	parameters := req.GetParameters()
	prefix := parameters[VolumePrefixKey]

	if len(prefix) == 0 {
		// If string is greater than max, truncate it, otherwise return original string
		if len(volumeID) > VolumeNameMaxLength {
			// Skip over 'pvc-'
			if volumeID[0:4] == "pvc-" {
				volumeID = volumeID[4:]
			}
			volumeID = strings.ReplaceAll(volumeID, "-", "")
			volumeID = volumeID[:VolumeNameMaxLength]
		}
	} else {
		// Skip over 'pvc-' and remove all dashes
		uuid := volumeID
		if volumeID[0:4] == "pvc-" {
			uuid = volumeID[4:]
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

	klog.Infof("TranslateVolumeName %q[%d], prefix %q[%d], result %q[%d]", originalVolumeID, len(originalVolumeID), prefix, len(prefix), volumeID, len(volumeID))

	return volumeID, nil
}
