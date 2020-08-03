/*
Copyright 2020 Smorgasbord Authors

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

package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newServerCmd(out io.Writer) *cobra.Command {
	var (
		addr string
	)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints version information",
		Long:  `Prints version information and the commit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&addr, "addr", "a", "0.0.0.0:8080", "which address the server will listen on")

	return cmd
}
