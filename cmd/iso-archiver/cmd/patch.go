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

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Patches an ISO archive",
	Long:  `Patches an ISO archive with files in the specified base directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("source")
		out, _ := cmd.Flags().GetString("out")
		baseDir, _ := cmd.Flags().GetString("base-dir")
		options, _ := cmd.Flags().GetStringSlice("options")
		if source == "" {
			exit.WithCodeT(exit.BadUsage, "source cannot be empty")
		}
		if out == "" {
			exit.WithCodeT(exit.BadUsage, "out cannot be empty")
		}
		files, err := getISOFileMapping(baseDir)
		if err != nil {
			exit.WithError("Cannot locate files in base dir", err)
		}
		if err := libarchive.PatchISO(source, out, files, options); err != nil {
			exit.WithError("Cannot patch ISO", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(patchCmd)
	patchCmd.Flags().String("source", "", "Path of the source ISO")
	patchCmd.Flags().String("out", "", "Path of the patched ISO")
	patchCmd.Flags().String("base-dir", "", "The base dir where patched files locate")
	patchCmd.Flags().StringSlice("options", []string{}, "ISO options")
	_ = patchCmd.MarkFlagRequired("source")
	_ = patchCmd.MarkFlagRequired("out")
	_ = patchCmd.MarkFlagRequired("base-dir")
}
