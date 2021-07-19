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
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	. "github.com/onsi/gomega"
)

func init() {
	fmt.Printf("Test Setup:\n")
	fmt.Printf("    VolumeNameMaxLength   = %d\n", VolumeNameMaxLength)
	fmt.Printf("    VolumePrefixMaxLength = %d\n", VolumePrefixMaxLength)
	fmt.Printf("\n")
}

func createRequestVolume(name string, prefix string) (csi.CreateVolumeRequest, error) {
	// Create a CSI CreateVolumeRequest and Response

	req := csi.CreateVolumeRequest{
		Name:       name,
		Parameters: map[string]string{VolumePrefixKey: prefix},
	}

	return req, nil
}

func runTest(t *testing.T, idin string, idout string, prefix string) {
	req, _ := createRequestVolume(idin, prefix)
	id, err := TranslateVolumeName(&req)

	g := NewWithT(t)
	g.Expect(err).To(BeNil())
	g.Expect(id).To(Equal(idout))
}

func TestTranslate(t *testing.T) {

	// Test empty name
	req := csi.CreateVolumeRequest{
		Name:       "",
		Parameters: map[string]string{"volPrefix": "csi"},
	}
	_, err := TranslateVolumeName(&req)
	g := NewWithT(t)
	g.Expect(err).ToNot(BeNil())

	// Test with no prefix
	runTest(t, "pvc-03c551d9-7e77-43ff-993e-c2308d2f09a1", "03c551d97e7743ff993ec2308d2f09a1", "")
	runTest(t, "03c551d97e7743ff993ec2308d2f09a1", "03c551d97e7743ff993ec2308d2f09a1", "")
	runTest(t, "8d2f09a1", "8d2f09a1", "")
	runTest(t, "51d9-7e77-43ff-993e-c2308d2f09a1", "51d9-7e77-43ff-993e-c2308d2f09a1", "")

	// Test with prefix
	runTest(t, "pvc-03c551d9-7e77-43ff-993e-c2308d2f09a1", "csi_51d97e7743ff993ec2308d2f09a1", "csi")
	runTest(t, "pvc-51d9-7e77-43ff-993e-c2308d2f09a1", "csi_51d97e7743ff993ec2308d2f09a1", "csi")
	runTest(t, "51d9-7e77-43ff-993e-c2308d2f09a1", "csi_51d97e7743ff993ec2308d2f09a1", "csi")
	runTest(t, "51d97e7743ff993ec2308d2f09a1", "csi_51d97e7743ff993ec2308d2f09a1", "csi")
	runTest(t, "51d97e7743ff993ec2308d2f09a1", "csi_51d97e7743ff993ec2308d2f09a1", "csi_123")
	runTest(t, "pvc-51d9-7e77-43ff-993e-c2308d2f09a1", "cd_51d97e7743ff993ec2308d2f09a1", "cd")
	runTest(t, "pvc-51d9-7e77-43ff-993e-c2308d2f09a1", "c_51d97e7743ff993ec2308d2f09a1", "c")
}

func TestValidate(t *testing.T) {
	g := NewWithT(t)
	g.Expect(ValidateVolumeName("abcdefghijklmnopqrstuvwxyz")).To(BeTrue())
	g.Expect(ValidateVolumeName("ABCDEFGHIJKLMNOPQRSTUVWXYZ")).To(BeTrue())
	g.Expect(ValidateVolumeName("a b _ . - c")).To(BeTrue())

	// 	Test unaccepable characters: " , < \
	g.Expect(ValidateVolumeName("\"abc")).To(BeFalse())
	g.Expect(ValidateVolumeName("abc,")).To(BeFalse())
	g.Expect(ValidateVolumeName("abc<def")).To(BeFalse())
	g.Expect(ValidateVolumeName("abc\\def")).To(BeFalse())
}
