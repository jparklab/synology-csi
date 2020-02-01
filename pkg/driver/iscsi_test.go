/*
 * Copyright 2019 Ji-Young Park(jiyoung.park.dev@gmail.com)
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
	"testing"

	"github.com/stretchr/testify/assert"
)

/************************************************************
 * Tests
 ************************************************************/
func TestParseSessionOutput(t *testing.T) {
    output := `
tcp: [11] 10.0.0.1:3260,1 iqn.2000-01.com.synology:kube-csi-pvc-e27d9fe3-7460-11e9-b909-74d02b7bd3f6 (non-flash)
tcp: [13] 10.0.0.1:3260,1 iqn.2000-01.com.synology:kube-csi-pvc-af2b5087-f07d-11e8-ae4a-74d02b7bd3f6 (non-flash)
tcp: [36] 10.0.0.1:3260,1 iqn.2000-01.com.synology:kube-csi-pvc-3d22dbbb-8132-4e32-bf17-a78b045ba2d1 (non-flash)
tcp: [38] 10.0.0.1:3260,1 iqn.2000-01.com.synology:kube-csi-pvc-25d6faf9-15bd-4031-8546-5d32ded6693e (non-flash)
tcp: [7] 10.0.0.1:3260,1 iqn.2000-01.com.synology:kube-csi-pvc-521ce1b7-f07c-11e8-ae4a-74d02b7bd3f6 (non-flash)
`
    sessions := parseSessionOutput(output)

    assert.Equal(t, len(sessions), 5)
    assert.Equal(t, sessions[0].IQN, "iqn.2000-01.com.synology:kube-csi-pvc-e27d9fe3-7460-11e9-b909-74d02b7bd3f6")
    assert.Equal(t, sessions[4].IQN, "iqn.2000-01.com.synology:kube-csi-pvc-521ce1b7-f07c-11e8-ae4a-74d02b7bd3f6")
}