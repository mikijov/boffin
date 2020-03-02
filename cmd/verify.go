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
	"log"
	"path/filepath"

	"github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify integrity of all files in the directory",
	Long:  `Verify directory for changes.`,
	// Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbDir, err := lib.FindBoffinDir(dbDirFlag)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		local, err := lib.LoadBoffin(dbDir)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}

		for _, file := range local.GetFiles() {
			if file.IsDeleted() {
				continue
			}
			path := filepath.Join(local.GetBaseDir(), file.Path())
			checksum, err := lib.CalculateChecksum(path)
			if err != nil {
				log.Printf("ERROR: %v", err)
			} else if checksum != file.Checksum() {
				log.Printf("%s: checksum does not match", file.Path())
			} else {
				log.Printf("%s: OK", file.Path())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// verifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
