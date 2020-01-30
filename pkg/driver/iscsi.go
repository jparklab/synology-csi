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
	"regexp"
	"strings"

	"github.com/golang/glog"
	utilexec "k8s.io/utils/exec"

	"github.com/jparklab/synology-csi/pkg/synology/api/iscsi"
)

type iscsiDriver struct {
	synologyHost string
}

type Session struct {
	IQN string
}

/************************************************************
 * helper functions
 ************************************************************/
func parseSessionOutput(output string) []Session {
	lines := strings.Split(output, "\n")
	iqn_re, _ := regexp.Compile(".*(iqn.\\S+)\\s.*")

	var sessions []Session
	for _, line := range lines {
		match := iqn_re.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}

		iqn := match[1]
		sessions = append(sessions, Session{iqn})
	}

	return sessions
}

/************************************************************
 * iscsiDriver functions
 ************************************************************/

/*
iscsiadm --mode discovery --type sendtargets --portal 192.168.8.10 --discover
iscsiadm --mode node --targetname "iqn.2000-01.com.synology:kube-csi-pvc-b96e0cf4-da74-49e3-88d6-f31b2e25c514" --portal 192.168.8.10 --login
iscsiadm --mode session

*/

func iscsiadm(cmdArgs ...string) utilexec.Cmd {
	// iscsiadm can be/is a shell script which just chroots to /host and exectues iscsi on the host.
	// hence a "sh -c" call is required
	command := "iscsiadm " + strings.Join(cmdArgs, " ")
	executor := utilexec.New()
	cmd := executor.Command("sh", "-c", command)
	glog.V(5).Infof("[EXECUTING] %s", command)
	return cmd
}

func (d *iscsiDriver) discovery() error {
	cmd := iscsiadm("--mode", "discovery", "--type", "sendtargets", "--portal", d.synologyHost, "--discover")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Error running iscsiadm discovery: %s(%v)", out, err)
		glog.V(3).Info(msg)
		return errors.New(msg)
	}
	return nil
}

func (d *iscsiDriver) login(target *iscsi.Target) error {
	cmd := iscsiadm("--mode", "node", "--targetname", target.IQN, "--portal", d.synologyHost, "--login")
	_, err := cmd.CombinedOutput()
	if err != nil {
		glog.V(3).Infof("Error running iscsiadm login: %v", err)
		return err
	}
	return nil
}

func (d *iscsiDriver) session() ([]Session, error) {
	cmd := iscsiadm("--mode", "session")
	out, err := cmd.CombinedOutput()
	if err != nil {
		glog.V(3).Infof("Error running iscsiadm session: %v", err)
		return nil, err
	}
	return parseSessionOutput(string(out)), nil
}

func (d *iscsiDriver) logout(target *iscsi.Target) error {
	cmd := iscsiadm("--mode", "node", "--targetname", target.IQN, "--logout")
	_, err := cmd.CombinedOutput()
	if err != nil {
		glog.V(3).Infof("Error running iscsiadm logout: %v", err)
		return err
	}
	return nil
}
