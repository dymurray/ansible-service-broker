//
// Copyright (c) 2017 Red Hat, Inc.
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
// Red Hat trademarks are not licensed under Apache License, Version 2.
// No permission is granted to use or replicate Red Hat trademarks that
// are incorporated in this software or its documentation.
//

package adapters

import (
	"testing"

	ft "github.com/openshift/ansible-service-broker/pkg/fusortest"
)

func TestSpecLabel(t *testing.T) {
	ft.AssertEqual(t, BundleSpecLabel, "com.redhat.apb.spec", "spec label does not match dockerhub")
}

func TestVersionCheck(t *testing.T) {
	// Test equal versions
	ft.AssertTrue(t, isCompatibleVersion("1.0", "1.0", "1.0"))
	// Test out of range by major version
	ft.AssertFalse(t, isCompatibleVersion("2.0", "1.0", "1.0"))
	// Test out of range by minor version
	ft.AssertTrue(t, isCompatibleVersion("1.10", "1.0", "1.0"))
	// Test out of range by major and minor version
	ft.AssertTrue(t, isCompatibleVersion("2.4", "1.0", "2.0"))
	// Test in range with differing  major and minor version
	ft.AssertTrue(t, isCompatibleVersion("1.10", "1.0", "2.0"))
	// Test out of range by major and minor version
	ft.AssertFalse(t, isCompatibleVersion("0.6", "1.0", "2.0"))
	// Test out of range by major and minor version and invalid version
	ft.AssertFalse(t, isCompatibleVersion("0.1.0", "1.0", "1.0"))
	// Test in range of long possible window
	ft.AssertTrue(t, isCompatibleVersion("2.5", "1.0", "3.0"))
	// Test invalid version
	ft.AssertFalse(t, isCompatibleVersion("1", "1.0", "3.0"))
	// Test invalid version
	ft.AssertFalse(t, isCompatibleVersion("2.5", "3.0", "4.0"))
}
