/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/spf13/cobra"

	"k8s.io/minikube/pkg/libarchive"
	"k8s.io/minikube/pkg/minikube/exit"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new ISO archive.",
	Long:  `Creates a new ISO archive with all the content in the specified directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		out, _ := cmd.Flags().GetString("out")
		baseDir, _ := cmd.Flags().GetString("base-dir")
		options, _ := cmd.Flags().GetStringSlice("options")
		if out == "" {
			exit.WithCodeT(exit.BadUsage, "out cannot be empty")
		}
		files, err := getISOFileMapping(baseDir)
		if err != nil {
			exit.WithError("Cannot locate files in base dir", err)
		}
		if err := libarchive.CreateISO(out, files, options); err != nil {
			exit.WithError("Cannot create ISO", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	createCmd.Flags().String("out", "", "Path of the new ISO")
	createCmd.Flags().String("base-dir", "", "The base dir where all files to add locate")
	createCmd.Flags().StringSlice("options", []string{}, "ISO options")
	_ = createCmd.MarkFlagRequired("out")
	_ = createCmd.MarkFlagRequired("base-dir")
}
