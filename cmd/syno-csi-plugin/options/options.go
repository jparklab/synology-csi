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

package options

import (
	"errors"
	"fmt"
	"github.com/jparklab/synology-csi/pkg/synology/options"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/golang/glog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/jparklab/synology-csi/pkg/driver"
)

// RunOptions stores option values
type RunOptions struct {
	NodeID       string
	Endpoint     string
	SynologyConf string
	CheckLogin   bool // Check if app is able to log into Synology and exit immediately
}

// NewRunOptions creates a default option object
func NewRunOptions() *RunOptions {
	return &RunOptions{
		NodeID:   "CSINode",
		Endpoint: "unix:///var/lib/kubelet/plugins/" + driver.DriverName + "/csi.sock",
	}
}

// ReadConfig reads synology configuration file
func ReadConfig(path string) (*options.SynologyOptions, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		glog.V(1).Infof("Unable to open config file: %v", err)
		return nil, err
	}

	conf := options.NewSynologyOptions()
	err = yaml.Unmarshal(f, &conf)
	if err != nil {
		glog.V(1).Infof("Failed to parse config: %v", err)
		return nil, err
	}

	if conf.LoginApiVersion <= 0 {
		conf.LoginApiVersion = 2
	}
	if conf.LoginApiVersion >= 3 {
		if conf.EnableSynoToken != nil {
			val := strings.ToLower(*conf.EnableSynoToken)
			if val == "yes" || val == "1" || val == "true" {
				yes := "yes"
				conf.EnableSynoToken = &yes
			} else {
				no := "no"
				conf.EnableSynoToken = &no
			}
		}
	}
	if conf.LoginApiVersion >= 6 {
		if conf.DeviceId == nil || *conf.DeviceId == "" {
			deviceId := os.Getenv("DEVICE_ID")
			conf.DeviceId = &deviceId
			if deviceId != "" {
				glog.Info("Using DEVICE_ID from environment variables: %v", deviceId)
			}
		}
		if conf.EnableDeviceToken != nil {
			val := strings.ToLower(*conf.EnableDeviceToken)
			if val == "yes" || val == "1" || val == "true" {
				yes := "yes"
				conf.EnableDeviceToken = &yes
			} else {
				no := "no"
				conf.EnableDeviceToken = &no
			}
		}
	}

	conf.LoginHttpMethod = strings.TrimSpace(strings.ToUpper(conf.LoginHttpMethod))
	if conf.LoginHttpMethod == "AUTO" {
		if conf.LoginApiVersion >= 6 {
			conf.LoginHttpMethod = "POST"
		} else {
			conf.LoginHttpMethod = "GET"
		}
	}

	if conf.LoginHttpMethod != "GET" && conf.LoginHttpMethod != "POST" {
		glog.V(1).Infof("Invalid login method in config: %v", conf.LoginHttpMethod)
		return nil, errors.New(fmt.Sprintf("Invalid login method in config: %v", conf.LoginHttpMethod))
	}

	return &conf, nil
}

// AddFlags adds command line options
func (o *RunOptions) AddFlags(cmd *cobra.Command, fs *pflag.FlagSet) {
	fs.StringVar(&o.NodeID, "nodeid", o.NodeID, "Node ID")
	fs.StringVar(&o.Endpoint, "endpoint", o.Endpoint, "CSI endpoint")

	fs.StringVar(&o.SynologyConf, "synology-config", o.SynologyConf, "Synology config yaml file")
	fs.BoolVar(&o.CheckLogin, "check-login", o.CheckLogin, "Just try to login and exit")

	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("synology-config")
}
