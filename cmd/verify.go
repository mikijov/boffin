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
	// "fmt"
	// "os"

	// "github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify integrity of all files in the directory",
	Long:  `Verify directory for changes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// dbDir := ""
		// if dbDirFlag != "" {
		// 	dbDir = dbDirFlag
		// } else {
		// 	wd, err := os.Getwd()
		// 	if err != nil {
		// 		fmt.Printf("ERROR: %v\n", err)
		// 		os.Exit(1)
		// 	}
		// 	dbDir, err = lib.FindDbDir(wd)
		// 	if err != nil {
		// 		fmt.Printf("ERROR: %v\n", err)
		// 		os.Exit(1)
		// 	}
		// }
		// db, err := lib.LoadFileDB(dbDir)
		// if err != nil {
		// 	fmt.Printf("ERROR: %v\n", err)
		// 	os.Exit(1)
		// }
		// if err = db.Verify(); err != nil {
		// 	fmt.Printf("ERROR: %v\n", err)
		// 	os.Exit(1)
		// }
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
