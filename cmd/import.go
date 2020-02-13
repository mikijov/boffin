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

// import (
// 	"fmt"
// 	"io"
// 	"os"
//
// 	"github.com/mikijov/boffin/lib"
// 	"github.com/spf13/cobra"
// )
//
// // importCmd represents the import command
// var importCmd = &cobra.Command{
// 	Use:   "import <remote-repo>",
// 	Short: "Import changes made in the remote repository.",
// 	Long: `Import will use meta-data from the local and remote repository silimarly
// 	to 'diff' and compare their contents. Any files that have been added or
// 	modified in the remote repository will be imported into local repository.
// 	Options can be used to control which changes will be imported.`,
// 	Args: cobra.ExactArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		dbDir, err := lib.FindBoffinDir(dbDirFlag)
// 		if err != nil {
// 			fmt.Printf("ERROR: %v\n", err)
// 			os.Exit(1)
// 		}
//
// 		dest, err := lib.LoadBoffin(dbDir)
// 		if err != nil {
// 			fmt.Printf("ERROR: %v\n", err)
// 			os.Exit(1)
// 		}
//
// 		dbDir, err = lib.FindBoffinDir(args[0])
// 		if err != nil {
// 			fmt.Printf("ERROR: %v\n", err)
// 			os.Exit(1)
// 		}
//
// 		source, err := lib.LoadBoffin(dbDir)
// 		if err != nil {
// 			fmt.Printf("ERROR: %v\n", err)
// 			os.Exit(1)
// 		}
//
// 		for _, diff := range dest.Diff(source) {
// 			if diff.Result == lib.DiffEqual {
// 				// fmt.Printf("=:%s\n", diff.Left.Path)
// 			} else if diff.Result == lib.DiffLeftAdded {
// 				fmt.Printf("%s: copy new file\n", diff.Left.Path)
// 			} else if diff.Result == lib.DiffRightAdded {
// 				// fmt.Printf("R:%s\n", diff.Right.Path)
// 			} else if diff.Result == lib.DiffLeftDeleted {
// 				// fmt.Printf("+:%s\n", diff.Left.Path)
// 			} else if diff.Result == lib.DiffRightAdded {
// 				// fmt.Printf("-:%s\n", diff.Left.Path)
// 			} else if diff.Result == lib.DiffLeftChanged {
// 				fmt.Printf("%s: copy changed file\n", diff.Left.Path)
// 			} else if diff.Result == lib.DiffRightChanged {
// 				// fmt.Printf("<:%s\n", diff.Left.Path)
// 			} else if diff.Result == lib.DiffConflict {
// 				fmt.Printf("%s: conflict\n", diff.Left.Path)
// 			} else {
// 				panic("unexpected result")
// 			}
// 		}
// 	},
// }
//
// // Copy the src file to dst. Any existing file will be overwritten and will not
// // copy file attributes.
// func copyFile(src, dst string) error {
// 	fmt.Printf("%s => %s\n", src, dst)
//
// 	stat, err := os.Stat(src)
// 	if err != nil {
// 		return err
// 	}
//
// 	in, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer in.Close()
//
// 	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()
//
// 	_, err = io.Copy(out, in)
// 	if err != nil {
// 		return err
// 	}
// 	err = out.Chmod(stat.Mode())
// 	if err != nil {
// 		return err
// 	}
// 	err = out.Close()
// 	if err != nil {
// 		return err
// 	}
// 	err = os.Chtimes(dst, stat.ModTime(), stat.ModTime())
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// func init() {
// 	rootCmd.AddCommand(importCmd)
//
// 	// Here you will define your flags and configuration settings.
//
// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// importCmd.PersistentFlags().String("foo", "", "A help for foo")
//
// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
