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

package lib

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FilterFunc is function type, that determines if a file should be processed or
// not. Return true to process file, or false if it should be skipped.
type FilterFunc func(info os.FileInfo, local *FileInfo) bool

// CheckIfMetaChanged implements FilterFunc, and will return true, i.e. it will
// trigger file check, if any of the file size of time has changes.
func CheckIfMetaChanged(info os.FileInfo, localFile *FileInfo) bool {
	if localFile == nil {
		return true
	}
	return localFile.IsDeleted() ||
		info.Size() != localFile.Size() ||
		!info.ModTime().Equal(localFile.Time())
}

// ForceCheck implements FilterFunc, and will force every file to be checked.
func ForceCheck(info os.FileInfo, local *FileInfo) bool {
	return true
}

// Update will compare the boffin repo with the files in the monitored directory
// and update the repo with any changes.
func Update(repo Boffin, filter FilterFunc) error {
	if filter == nil {
		filter = CheckIfMetaChanged
	}

	dir := repo.GetBaseDir()

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("base directory '%s' does not exist", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("base directory '%s' is not a directory", dir)
	}

	localByPath := filesToPathMap(repo.GetFiles())

	checkedFiles := &db{
		dbDir:        repo.GetDbDir(),
		absBaseDir:   repo.GetBaseDir(),
		absImportDir: repo.GetImportDir(),
		baseDir:      repo.GetBaseDir(),
		importDir:    repo.GetImportDir(),
		files:        []*FileInfo{},
	}

	// # get list of files that should be checked
	// - for each file on the file system
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				log.Printf("%s: permission denied", path)
			} else {
				return fmt.Errorf("%s: error getting file info: %s", path, err)
			}
		}
		if info.IsDir() {
			if info.Name() == defaultDbDir { // skip DB directory
				// fmt.Printf("skip %s\n", path)
				return filepath.SkipDir
			} else if strings.HasPrefix(info.Name(), ".") {
				// fmt.Printf("skip %s\n", path)
				return filepath.SkipDir
			}
			// fmt.Printf("dir %s\n", path)
			return nil
		}

		// sanity check which has never fired
		root := path[:len(dir)]
		if dir != root {
			// this should never happen
			log.Panicf("unexpected error; root mismatch '%s' != '%s'", dir, root)
		}

		relPath := path[len(dir)+1:]

		localFile, ok := localByPath[relPath]
		var checkFile bool
		if ok {
			delete(localByPath, relPath)
			checkFile = filter(info, localFile)
		} else {
			checkFile = true
		}

		if checkFile {
			// fmt.Printf("CC%s\n", relPath)
			hash, err := CalculateChecksum(path)
			if err != nil {
				return err
			}
			log.Printf("%s: %s\n", hash, relPath)

			checkedFiles.files = append(checkedFiles.files, &FileInfo{
				History: []*FileEvent{
					&FileEvent{
						Path:     relPath,
						Time:     info.ModTime(),
						Size:     info.Size(),
						Checksum: hash,
					},
				},
			})
		} else { // no need to check, assume identical
			// fmt.Printf("==%s\n", localFile.Path())
			checkedFiles.files = append(checkedFiles.files, localFile)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return Diff(repo, checkedFiles, &updateAction{
		repo: repo,
	})
}

type updateAction struct {
	repo Boffin
}

func (a *updateAction) Unchanged(localFile, remoteFile *FileInfo) {
	// fmt.Printf("=%s\n", localFile.Path())
}

func (a *updateAction) MetaDataChanged(localFile, remoteFile *FileInfo) {
	fmt.Printf("M%s\n", localFile.Path())
	localFile.History = append(localFile.History, remoteFile.History...)
}

func (a *updateAction) Moved(localFile, remoteFile *FileInfo) {
	fmt.Printf("@%s => %s\n", localFile.Path(), remoteFile.Path())
	localFile.History = append(localFile.History, remoteFile.History...)
}

func (a *updateAction) LocalOnly(localFile *FileInfo) {
	fmt.Printf("-%s\n", localFile.Path())
	localFile.MarkDeleted()
}

func (a *updateAction) LocalOld(localFile *FileInfo) {
	// do nothing
}

func (a *updateAction) RemoteOnly(remoteFile *FileInfo) {
	fmt.Printf("+%s\n", remoteFile.Path())
	a.repo.AddFile(remoteFile)
}

func (a *updateAction) RemoteOld(remoteFile *FileInfo) {
	// do nothing
}

func (a *updateAction) LocalDeleted(localFile, remoteFile *FileInfo) {
	log.Panicf("local deleted; should never happen for updateAction: %s", localFile.Path())
}

func (a *updateAction) RemoteDeleted(localFile, remoteFile *FileInfo) {
	panic("remote deleted should never happen for updateAction")
}

func (a *updateAction) LocalChanged(localFile, remoteFile *FileInfo) {
	// panic("local changed should never happen for updateAction")
	fmt.Printf("WARNING: Local should not change during update: ~%s => %s\n", localFile.Path(), remoteFile.Path())
}

func (a *updateAction) RemoteChanged(localFile, remoteFile *FileInfo) {
	fmt.Printf("~%s => %s\n", localFile.Path(), remoteFile.Path())
	localFile.History = append(localFile.History, &FileEvent{
		Path:     remoteFile.Path(),
		Time:     remoteFile.Time(),
		Size:     remoteFile.Size(),
		Checksum: remoteFile.Checksum(),
	})
}

func (a *updateAction) ConflictPath(localFile, remoteFile *FileInfo) {
	fmt.Printf("~%s => %s\n", localFile.Path(), remoteFile.Path())
	localFile.History = append(localFile.History, &FileEvent{
		Path:     remoteFile.Path(),
		Time:     remoteFile.Time(),
		Size:     remoteFile.Size(),
		Checksum: remoteFile.Checksum(),
	})
}

func (a *updateAction) ConflictHash(localFiles, remoteFiles []*FileInfo) {
	if len(localFiles) == 1 {
		for _, remoteFile := range remoteFiles {
			fmt.Printf("+%s\n", remoteFile.Path())
			a.repo.AddFile(remoteFile)
		}
	}

	for _, file := range localFiles {
		fmt.Printf("!%s\n", file.Path())
	}
	for _, file := range remoteFiles {
		fmt.Printf("!%s\n", file.Path())
	}
}
