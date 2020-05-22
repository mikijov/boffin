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

type diffAction struct {
}

func (a *diffAction) Unchanged(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf("==:%s\n", localFile.Path())
}

func (a *diffAction) Moved(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf("=>:%s => %s\n", localFile.Path(), remoteFile.Path())
}

func (a *diffAction) LocalOnly(localFile *lib.FileInfo) {
	fmt.Printf("L+:%s\n", localFile.Path())
}

func (a *diffAction) LocalOld(localFile *lib.FileInfo) {
	// fmt.Printf("L+:%s\n", localFile.Path())
}

func (a *diffAction) RemoteOnly(remoteFile *lib.FileInfo) {
	fmt.Printf("R+:%s\n", remoteFile.Path())
}

func (a *diffAction) RemoteOld(remoteFile *lib.FileInfo) {
	// fmt.Printf("R+:%s\n", remoteFile.Path())
}

func (a *diffAction) LocalDeleted(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf("L-:%s\n", localFile.Path())
}

func (a *diffAction) RemoteDeleted(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf("R-:%s\n", remoteFile.Path())
}

func (a *diffAction) LocalChanged(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf(">>:%s\n", localFile.Path())
}

func (a *diffAction) RemoteChanged(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf("<<:%s\n", remoteFile.Path())
}

func (a *diffAction) ConflictPath(localFile, remoteFile *lib.FileInfo) {
	fmt.Printf("!!:%s ! %s\n", localFile.Path(), remoteFile.Path())
}

func (a *diffAction) ConflictHash(localFiles, remoteFiles []*lib.FileInfo) {
	if len(localFiles) == 1 && len(remoteFiles) == 1 {
		localFile := localFiles[0]
		remoteFile := remoteFiles[0]
		fmt.Printf("=>:%s => %s\n", localFile.Path(), remoteFile.Path())
		localFile.History = append(localFile.History, &lib.FileEvent{
			Path:     remoteFile.Path(),
			Time:     remoteFile.Time(),
			Size:     remoteFile.Size(),
			Checksum: remoteFile.Checksum(),
		})
		return
	}

	for _, file := range localFiles {
		fmt.Printf("!!:%s\n", file.Path())
	}
	for _, file := range remoteFiles {
		fmt.Printf("!!:%s\n", file.Path())
	}
}

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
		if dbDir == "" {
			var err error
			dbDir, err = lib.FindBoffinDir(dbDir)
			if err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}
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

		if err = lib.Diff(local, remote, &diffAction{}); err != nil {
			log.Fatalf("ERROR: %v\n", err)
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
