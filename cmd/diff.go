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
package cmd

import (
	"fmt"
	"os"

	"github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff <remote-repo>",
	Short: "Show differences between current and remote repo.",
	Long: `Diff will use meta-data from the repository and compare their contents.
	It will show added, removed and changed files. If the file by the same name
	exists in both repositories, but they do not share the same history, a
	conflict will be reported.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbDir, err := lib.FindBoffinDir(dbDirFlag)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		left, err := lib.LoadBoffin(dbDir)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		dbDir, err = lib.FindBoffinDir(args[0])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		right, err := lib.LoadBoffin(dbDir)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		for _, diff := range left.Diff(right) {
			if diff.Result == lib.DiffEqual {
				fmt.Printf("=:%s\n", diff.Left.Path)
			} else if diff.Result == lib.DiffLeftAdded {
				fmt.Printf("L:%s\n", diff.Left.Path)
			} else if diff.Result == lib.DiffRightAdded {
				fmt.Printf("R:%s\n", diff.Right.Path)
			} else if diff.Result == lib.DiffLeftDeleted {
				fmt.Printf("+:%s\n", diff.Left.Path)
			} else if diff.Result == lib.DiffRightAdded {
				fmt.Printf("-:%s\n", diff.Left.Path)
			} else if diff.Result == lib.DiffLeftChanged {
				fmt.Printf(">:%s\n", diff.Left.Path)
			} else if diff.Result == lib.DiffRightChanged {
				fmt.Printf("<:%s\n", diff.Left.Path)
			} else if diff.Result == lib.DiffConflict {
				fmt.Printf("~:%s\n", diff.Left.Path)
			} else {
				fmt.Printf("~:%s\n", diff.Left.Path)
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
