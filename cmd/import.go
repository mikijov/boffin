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

		action := &importAction{
			local:  local,
			remote: remote,
		}

		if err = lib.Diff(local, remote, action); err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}
		if !dryRun {
			if err = boffin.Save(); err != nil {
				log.Fatalf("ERROR: %v\n", err)
			}
		}

		if action.exit != 0 {
			os.Exit(action.exit)
		}
	},
}

type importAction struct {
	exit   int
	local  lib.Boffin
	remote lib.Boffin
}

func (a *importAction) Unchanged(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf("==:%s\n", localFile.Path())
}

func (a *importAction) Moved(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf("=>:%s => %s\n", localFile.Path(), remoteFile.Path())
}

func (a *importAction) LocalOnly(localFile *lib.FileInfo) {
	// fmt.Printf("L+:%s\n", localFile.Path())
}

func (a *importAction) LocalOld(localFile *lib.FileInfo) {
	// do nothing
}

func (a *importAction) RemoteOnly(remoteFile *lib.FileInfo) {
	// fmt.Printf("R+:%s\n", remoteFile.Path())

	src := filepath.Join(a.remote.GetBaseDir(), remoteFile.Path())
	dest := filepath.Join(a.local.GetImportDir(), remoteFile.Path())

	if err := addFile(src, dest); err != nil {
		log.Printf("%v", err)
		a.exit = 1
	} else {
		remoteFile.History = append(remoteFile.History, &lib.FileEvent{
			Path:     filepath.Join(a.local.GetRelImportDir(), remoteFile.Path()),
			Time:     remoteFile.Time(),
			Size:     remoteFile.Size(),
			Checksum: remoteFile.Checksum(),
		})
		a.local.AddFile(remoteFile)
	}
}

func (a *importAction) RemoteOld(remoteFile *lib.FileInfo) {
	// do nothing
}

func (a *importAction) LocalDeleted(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf("L-:%s\n", localFile.Path())
}

func (a *importAction) RemoteDeleted(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf("R-:%s\n", remoteFile.Path())
}

func (a *importAction) LocalChanged(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf(">>:%s\n", localFile.Path())
}

func (a *importAction) RemoteChanged(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf("<<:%s\n", remoteFile.Path())

	src := filepath.Join(a.remote.GetBaseDir(), remoteFile.Path())
	dest := filepath.Join(a.local.GetBaseDir(), localFile.Path())

	if err := replaceFile(src, dest); err != nil {
		log.Printf("%v", err)
		a.exit = 1
	} else {
		localFile.History = append(localFile.History, &lib.FileEvent{
			Path:     localFile.Path(),
			Time:     remoteFile.Time(),
			Size:     remoteFile.Size(),
			Checksum: remoteFile.Checksum(),
		})
	}
}

func (a *importAction) ConflictPath(localFile, remoteFile *lib.FileInfo) {
	// fmt.Printf("!!:%s ! %s\n", localFile.Path(), remoteFile.Path())
}

func (a *importAction) ConflictHash(localFiles, remoteFiles []*lib.FileInfo) {
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

func addFile(src, dest string) error {
	if fi, err := os.Stat(dest); err == nil {
		if fi.IsDir() {
			return fmt.Errorf("destination is a directory for addFile operation: %s", dest)
		}
		return fmt.Errorf("destination file exists for addFile operation: %s", dest)
	} else if os.IsNotExist(err) {
		fmt.Printf("Add: %s => %s\n", src, dest)
		// this path is what is expected
		return _copyFile(src, dest)
	} else {
		return fmt.Errorf("unexpected error when checking '%s' during addFile operation: %s", dest, err)
	}
}

func replaceFile(src, dest string) error {
	if fi, err := os.Stat(dest); err == nil {
		if fi.IsDir() {
			return fmt.Errorf("destination is a directory: %s", dest)
		}
		fmt.Printf("Update: %s => %s\n", src, dest)
		// this path is what is expected
		return _copyFile(src, dest)
	} else if os.IsNotExist(err) {
		return fmt.Errorf("destination file missing for replaceFile operation: %s", dest)
	} else {
		return fmt.Errorf("unexpected error when checking '%s' during replaceFile operation: %s", dest, err)
	}
}

// Copy the src file to dest. Any existing file will be overwritten and will not
// copy file attributes.
func _copyFile(src, dest string) error {
	if dryRun {
		return nil
	}

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0777); err != nil {
		return err
	}

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
	var backupErr error
	if backupErr = os.Rename(dest, backupDest); backupErr != nil {
		if !os.IsNotExist(backupErr) {
			return backupErr
		}
	} else {
		defer os.Rename(backupDest, dest)
	}
	if err := os.Rename(tempDest, dest); err != nil {
		return err
	}
	if backupErr == nil {
		if err := os.Remove(backupDest); err != nil {
			return err
		}
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
