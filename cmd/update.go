/*
Copyright (C) 2019 Milutin Jovanvović

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
	"os"

	"github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update repository with meta-data for any changed files.",
	Long: `Update looks for any added, removed or changed files in the
	repository and updates meta-data correspondingly. By default, only if file
	size or modification timestamp are changed will the file checksum be checked.`,
	// Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbDir, err := lib.FindBoffinDir(dbDirFlag)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		boffin, err := lib.LoadBoffin(dbDir)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		if err = boffin.Update(); err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		if err = boffin.Save(); err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
