/*
Copyright (C) 2020 Milutin Jovanvović

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package cmd ...
package cmd

import (
	"fmt"
	"log"

	"github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff <remote-repo>",
	Short: "Show differences between local and remote repo.",
	Long: `Diff will use meta-data from the repository and compare their contents.
	It will show added, removed and changed files. If the file by the same name
	exists in both repositories, but they do not share the same history, a
	conflict will be reported.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbDir, err := lib.FindBoffinDir(dbDirFlag)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}
		local, err := lib.LoadBoffin(dbDir)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}

		dbDir, err = lib.FindBoffinDir(args[0])
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}
		remote, err := lib.LoadBoffin(dbDir)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}

		for _, diff := range local.Diff2(remote) {
			if diff.Result == lib.DiffEqual {
				fmt.Printf("=:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffLocalAdded {
				fmt.Printf("L:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffRemoteAdded {
				fmt.Printf("R:%s\n", diff.Remote.Path())
			} else if diff.Result == lib.DiffLocalDeleted {
				fmt.Printf("+:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffRemoteDeleted {
				fmt.Printf("-:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffLocalChanged {
				fmt.Printf("<:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffRemoteChanged {
				fmt.Printf(">:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffConflict {
				fmt.Printf("~:%s\n", diff.Local.Path())
			} else {
				fmt.Printf("~:%s\n", diff.Local.Path())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// diffCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// diffCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
