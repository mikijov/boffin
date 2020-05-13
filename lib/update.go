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

type FilterFunc func(info os.FileInfo, local *FileInfo) bool

func CheckIfMetaChanged(info os.FileInfo, localFile *FileInfo) bool {
	if localFile == nil {
		return true
	}
	return localFile.IsDeleted() ||
		info.Size() != localFile.Size() ||
		!info.ModTime().Equal(localFile.Time())
}

func ForceCheck(info os.FileInfo, local *FileInfo) bool {
	return true
}

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

		root := path[:len(dir)]
		if dir != root {
			// this should never happen
			log.Panicf("unexpected error; root mismatch '%s' != '%s'", dir, root)
		}

		relPath := path[len(dir)+1:]

		//   - if the path matches exactly
		//     - if forced check or if size/date changed
		//       - put file in to-be-checked list
		//     - else
		//       - mark checked
		//   - else
		//     - put file in to-be-checked list
		localFile, ok := localByPath[relPath]
		if filter(info, localFile) {
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
		} else if ok {
			// fmt.Printf("=%s\n", localFile.Path())
			delete(localByPath, relPath)
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
	fmt.Printf("=%s\n", localFile.Path())
}

func (a *updateAction) Moved(localFile, remoteFile *FileInfo) {
	fmt.Printf("@%s => %s\n", localFile.Path(), remoteFile.Path())
	localFile.History = append(localFile.History, remoteFile.History...)
}

func (a *updateAction) LocalOnly(localFile *FileInfo) {
	fmt.Printf("-%s\n", localFile.Path())
	localFile.MarkDeleted()
}

func (a *updateAction) RemoteOnly(remoteFile *FileInfo) {
	fmt.Printf("+%s\n", remoteFile.Path())

	a.repo.AddFile(remoteFile)
}

func (a *updateAction) LocalDeleted(localFile, remoteFile *FileInfo) {
	log.Panicf("local deleted; should never happen for updateAction: %s", localFile.Path())
}

func (a *updateAction) RemoteDeleted(localFile, remoteFile *FileInfo) {
	panic("remote deleted should never happen for updateAction")
}

func (a *updateAction) LocalChanged(localFile, remoteFile *FileInfo) {
	panic("local changed should never happen for updateAction")
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
	if len(localFiles) == 1 && len(remoteFiles) == 1 {
		localFile := localFiles[0]
		remoteFile := remoteFiles[0]
		fmt.Printf("~%s => %s\n", localFile.Path(), remoteFile.Path())
		localFile.History = append(localFile.History, &FileEvent{
			Path:     remoteFile.Path(),
			Time:     remoteFile.Time(),
			Size:     remoteFile.Size(),
			Checksum: remoteFile.Checksum(),
		})
		return
	}

	for _, file := range localFiles {
		fmt.Printf("!%s\n", file.Path())
	}
	for _, file := range remoteFiles {
		fmt.Printf("!%s\n", file.Path())
	}
}
