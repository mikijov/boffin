/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

var deleteDuplicates bool

// findDuplicatesCmd represents the findDuplicates command
var findDuplicatesCmd = &cobra.Command{
	Use:   "find-duplicates",
	Short: "find and display duplicate files",
	Long:  `find and display duplicate files`,
	Run: func(cmd *cobra.Command, args []string) {
		if dbDir == "" {
			var err error
			dbDir, err = lib.FindBoffinDir(dbDir)
			if err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}
		}

		local, err := lib.LoadBoffin(dbDir)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}

		for hash, files := range lib.FilesToHashMap(local.GetFiles()) {
			if len(files) > 1 {
				fmt.Printf("%s:\n", hash)
				keep := true
				for _, file := range files {
					if deleteDuplicates && !keep {
						fmt.Printf(" -%s\n", file.Path())
						if !dryRun {
							path := filepath.Join(local.GetBaseDir(), file.Path())
							if err := os.Remove(path); err != nil {
								log.Printf("%v", err)
							}
						}
					} else {
						fmt.Printf("  %s\n", file.Path())
						keep = false
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(findDuplicatesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	findDuplicatesCmd.PersistentFlags().BoolVar(&deleteDuplicates, "delete", false, "delete all but one of the duplicates")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// findDuplicatesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
