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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/jparklab/synology-csi/cmd/syno-csi-plugin/options"
	"github.com/jparklab/synology-csi/pkg/driver"
)

func main() {
	runOptions := options.NewRunOptions()

	rootCmd := &cobra.Command{
		Use:  "synology-csi-plugin",
		Long: "Synology CSI(Container Storage Interface) plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			flag.CommandLine.Parse([]string{"-v", "8", "--logtostderr=1"})

			endpoint := runOptions.Endpoint
			nodeID := runOptions.NodeID

			synoOption, err := options.ReadConfig(runOptions.SynologyConf)
			if err != nil {
				fmt.Printf("Failed to read config: %v\n", err)
				return err
			}

			if runOptions.CheckLogin {
				_, _, err := driver.Login(synoOption)
				if err != nil {
					fmt.Printf("Failed to login: %v\n", err)
				}
				return err
			}

			drv, err := driver.NewDriver(nodeID, endpoint, synoOption)
			if err != nil {
				fmt.Printf("Failed to create driver: %v\n", err)
				return err
			}
			drv.Run()

			return nil
		},
		SilenceUsage: true,
	}

	runOptions.AddFlags(rootCmd, rootCmd.PersistentFlags())
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
