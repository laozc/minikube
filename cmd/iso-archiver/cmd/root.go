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
	goflag "flag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"

	// initflag must be imported before any other minikube pkg.
	// Fix for https://github.com/kubernetes/minikube/issues/4866
	_ "k8s.io/minikube/pkg/initflag"

	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/translate"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "iso-archiver",
	Short: "iso-archiver is a tool for ISO archive manipulation.",
	Long:  `iso-archiver is a CLI tool that creates/extracts/patches ISO archive.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.Short = translate.T(RootCmd.Short)
	RootCmd.Long = translate.T(RootCmd.Long)
	RootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Usage = translate.T(flag.Usage)
	})

	if err := RootCmd.Execute(); err != nil {
		// Cobra already outputs the error, typically because the user provided an unknown command.
		os.Exit(exit.BadUsage)
	}
}

func init() {
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		exit.WithError("Unable to bind flags", err)
	}
	RootCmd.Flags().String("source", "", "Path of the source ISO")
	RootCmd.Flags().String("out", "", "Path of the patched ISO")
	RootCmd.Flags().String("baseDir", "", "The base dir where patched files locate")
	RootCmd.Flags().StringSlice("options", []string{}, "ISO options")
	RootCmd.MarkFlagRequired("source")
	RootCmd.MarkFlagRequired("out")
	RootCmd.MarkFlagRequired("baseDir")
}

func main() {
	Execute()
}
