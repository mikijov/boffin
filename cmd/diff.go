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

	"git.voreni.com/miki/boffin/lib"
	"github.com/spf13/cobra"
)

var (
	diffHideUnchanged      = false
	diffHideMetadataChange = false
	diffHideMoved          = false
	diffHideLocalOnly      = false
	diffHideLocalOld       = false
	diffHideRemoteOnly     = false
	diffHideRemoteOld      = false
	diffHideLocalDeleted   = false
	diffHideRemoteDeleted  = false
	diffHideLocalChanged   = false
	diffHideRemoteChanged  = false
	diffHideConflict       = false
)

type diffAction struct {
}

func (a *diffAction) Unchanged(localFile, remoteFile *lib.FileInfo) {
	if !diffHideUnchanged {
		fmt.Printf("==:%s\n", localFile.Path())
	}
}

func (a *diffAction) MetaDataChanged(localFile, remoteFile *lib.FileInfo) {
	if !diffHideMetadataChange {
		fmt.Printf("MD:%s\n", localFile.Path())
	}
}

func (a *diffAction) Moved(localFile, remoteFile *lib.FileInfo) {
	if !diffHideMoved {
		fmt.Printf("=>:%s => %s\n", localFile.Path(), remoteFile.Path())
	}
}

func (a *diffAction) LocalOnly(localFile *lib.FileInfo) {
	if !diffHideLocalOnly {
		fmt.Printf("L+:%s\n", localFile.Path())
	}
}

func (a *diffAction) LocalOld(localFile *lib.FileInfo) {
	if !diffHideLocalOld {
		// fmt.Printf("L+:%s\n", localFile.Path())
	}
}

func (a *diffAction) RemoteOnly(remoteFile *lib.FileInfo) {
	if !diffHideRemoteOnly {
		fmt.Printf("R+:%s\n", remoteFile.Path())
	}
}

func (a *diffAction) RemoteOld(remoteFile *lib.FileInfo) {
	if !diffHideRemoteOld {
		// fmt.Printf("R+:%s\n", remoteFile.Path())
	}
}

func (a *diffAction) LocalDeleted(localFile, remoteFile *lib.FileInfo) {
	if !diffHideLocalDeleted {
		fmt.Printf("L-:%s\n", localFile.Path())
	}
}

func (a *diffAction) RemoteDeleted(localFile, remoteFile *lib.FileInfo) {
	if !diffHideRemoteDeleted {
		fmt.Printf("R-:%s\n", remoteFile.Path())
	}
}

func (a *diffAction) LocalChanged(localFile, remoteFile *lib.FileInfo) {
	if !diffHideLocalChanged {
		fmt.Printf(">>:%s\n", localFile.Path())
	}
}

func (a *diffAction) RemoteChanged(localFile, remoteFile *lib.FileInfo) {
	if !diffHideRemoteChanged {
		fmt.Printf("<<:%s\n", remoteFile.Path())
	}
}

func (a *diffAction) ConflictPath(localFile, remoteFile *lib.FileInfo) {
	if !diffHideConflict {
		fmt.Printf("!!:%s ! %s\n", localFile.Path(), remoteFile.Path())
	}
}

func (a *diffAction) ConflictHash(localFiles, remoteFiles []*lib.FileInfo) {
	// if len(localFiles) == 1 && len(remoteFiles) == 1 {
	// 	localFile := localFiles[0]
	// 	remoteFile := remoteFiles[0]
	// 	fmt.Printf("=>:%s => %s\n", localFile.Path(), remoteFile.Path())
	// 	localFile.History = append(localFile.History, &lib.FileEvent{
	// 		Path:     remoteFile.Path(),
	// 		Time:     remoteFile.Time(),
	// 		Size:     remoteFile.Size(),
	// 		Checksum: remoteFile.Checksum(),
	// 	})
	// 	return
	// }
	//
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

	diffCmd.Flags().BoolVar(&diffHideUnchanged, "hide-unchanged", false, "hide files that have not changed")
	diffCmd.Flags().BoolVar(&diffHideMetadataChange, "hide-metadata-change", false, "hide files where only metadata has changed, but are otherwise same")
	diffCmd.Flags().BoolVar(&diffHideMoved, "hide-moved", false, "hide files that have moved")
	diffCmd.Flags().BoolVar(&diffHideLocalOnly, "hide-local-only", false, "hide files that only exist in local repo")
	diffCmd.Flags().BoolVar(&diffHideLocalOld, "hide-local-old", false, "hide files whose local version is old")
	diffCmd.Flags().BoolVar(&diffHideRemoteOnly, "hide-remote-only", false, "hide files that only exist in remote repo")
	diffCmd.Flags().BoolVar(&diffHideRemoteOld, "hide-remote-old", false, "hide files whole remote version is old")
	diffCmd.Flags().BoolVar(&diffHideLocalDeleted, "hide-local-deleted", false, "hide files that were locally deleted, but still exist in remote repo")
	diffCmd.Flags().BoolVar(&diffHideRemoteDeleted, "hide-remote-deleted", false, "hide files that were remotely deleted, but still exist in local repo")
	diffCmd.Flags().BoolVar(&diffHideLocalChanged, "hide-local-changed", false, "hide changed files which local version is newest")
	diffCmd.Flags().BoolVar(&diffHideRemoteChanged, "hide-remote-changed", false, "hide changed files which remote version is newest")
	diffCmd.Flags().BoolVar(&diffHideConflict, "hide-conflict", false, "hide files which have conflicting changes in both local and remote repo")
}
