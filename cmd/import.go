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
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/mikijov/boffin/lib"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <remote-repo>",
	Short: "Import changes made in the remote repository.",
	Long: `Import will use meta-data from the local and remote repository silimarly
	to 'diff' and compare their contents. Any files that have been added or
	modified in the remote repository will be imported into local repository.
	Options can be used to control which changes will be imported.`,
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

		exit := 0

		for _, diff := range local.Diff2(remote) {
			if diff.Result == lib.DiffEqual {
				// fmt.Printf("=:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffLocalAdded {
				// fmt.Printf("L:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffRemoteAdded {
				fmt.Printf("R:%s\n", diff.Remote.Path())

				src := filepath.Join(remote.GetBaseDir(), diff.Remote.Path())
				dest := filepath.Join(local.GetImportDir(), diff.Remote.Path())

				if err := copyFile(src, dest); err != nil {
					exit = 1
					break
				}
			} else if diff.Result == lib.DiffLocalDeleted {
				// fmt.Printf("+:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffRemoteDeleted {
				// fmt.Printf("-:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffLocalChanged {
				// fmt.Printf("<:%s\n", diff.Local.Path())
			} else if diff.Result == lib.DiffRemoteChanged {
				fmt.Printf(">:%s\n", diff.Local.Path())

				src := filepath.Join(remote.GetBaseDir(), diff.Remote.Path())
				dest := filepath.Join(local.GetBaseDir(), diff.Local.Path())

				if err := copyFile(src, dest); err != nil {
					exit = 1
					break
				}
			} else if diff.Result == lib.DiffConflict {
				fmt.Printf("!:%s\n", diff.Local.Path())
				exit = 2
			} else {
				fmt.Printf("?:%s\n", diff.Local.Path())
				exit = 2
			}
		}

		os.Exit(exit)
	},
}

// Copy the src file to dest. Any existing file will be overwritten and will not
// copy file attributes.
func copyFile(src, dest string) error {
	fmt.Printf("%s => %s\n", src, dest)

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// copy new file to temporary file
	tempDest := dest + ".boffin-tmp"
	out, err := os.OpenFile(tempDest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.Remove(tempDest)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Chmod(stat.Mode())
	if err != nil {
		return err
	}
	err = out.Close()
	if err != nil {
		return err
	}
	err = os.Chtimes(tempDest, stat.ModTime(), stat.ModTime())
	if err != nil {
		return err
	}

	// put temporary file into final desination
	backupDest := dest + ".boffin-old"
	if err := os.Rename(dest, backupDest); err != nil {
		return err
	}
	defer os.Rename(backupDest, dest)
	if err := os.Rename(tempDest, dest); err != nil {
		return err
	}
	if err := os.Remove(backupDest); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
