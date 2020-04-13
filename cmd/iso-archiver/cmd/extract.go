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

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extracts ISO archive content.",
	Long:  `Extracts ISO archive content to the specified directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("source")
		out, _ := cmd.Flags().GetString("out")
		if source == "" {
			exit.WithCodeT(exit.BadUsage, "source cannot be empty")
		}
		if out == "" {
			exit.WithCodeT(exit.BadUsage, "out cannot be empty")
		}
		if err := libarchive.ExtractISO(source, out); err != nil {
			exit.WithError("Cannot extract ISO", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(extractCmd)
	extractCmd.Flags().String("source", "", "Path of the ISO")
	extractCmd.Flags().String("out", "", "The directory to store all the extracted files")
	_ = extractCmd.MarkFlagRequired("source")
	_ = extractCmd.MarkFlagRequired("out")
}
