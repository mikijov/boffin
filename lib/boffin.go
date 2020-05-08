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

package lib

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const defaultDbDir = ".boffin"
const filesFilename = "files.json"
const newFilesFilename = "files.json.tmp"

const changed = "changed"
const deleted = "deleted"

// FileEvent ...
type FileEvent struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Type     string    `json:"event"`
	Time     time.Time `json:"time"`
	Checksum string    `json:"checksum,omitempty"`
}

// FileInfo ...
type FileInfo struct {
	History []*FileEvent `json:"history,omitempty"`
}

type FilterFunc func(info os.FileInfo, local *FileInfo) bool

// Boffin ...
type Boffin interface {
	GetFiles() []*FileInfo

	GetDbDir() string
	GetBaseDir() string
	GetImportDir() string

	Save() error
	Sort()
	// Update(filter FilterFunc) error
	Diff3(remote Boffin, action DiffAction) error
}

type DiffAction interface {
	Unchanged(localFile, remoteFile *FileInfo)
	Moved(localFile, remoteFile *FileInfo)
	LocalOnly(localFile *FileInfo)
	RemoteOnly(remoteFile *FileInfo)
	LocalDeleted(localFile, remoteFile *FileInfo)
	RemoteDeleted(localFile, remoteFile *FileInfo)
	LocalChanged(localFile, remoteFile *FileInfo)
	RemoteChanged(localFile, remoteFile *FileInfo)
	Conflict(localFile, remoteFile []*FileInfo)
}

// Checksum ...
func (fi *FileInfo) Checksum() string {
	return fi.History[len(fi.History)-1].Checksum
	// for i := range fi.History {
	// 	event := fi.History[len(fi.History)-1-i]
	// 	if event.Checksum != "" {
	// 		return event.Checksum
	// 	}
	// }
	// return ""
}

// Path ...
func (fi *FileInfo) Path() string {
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Path
		}
	}
	return ""
}

// Size ...
func (fi *FileInfo) Size() int64 {
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Size
		}
	}
	return 0
}

// Time ...
func (fi *FileInfo) Time() time.Time {
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Time
		}
	}
	return time.Time{}
}

// IsDeleted ...
func (fi *FileInfo) IsDeleted() bool {
	if len(fi.History) == 0 {
		return true
	}
	return fi.History[len(fi.History)-1].Checksum == ""
}

// func (fi *FileInfo) inheritsFrom(checksum string) bool {
// 	for _, event := range fi.History {
// 		if event.Checksum == checksum {
// 			return true
// 		}
// 	}
// 	return false
// }

func (fi *FileInfo) markDeleted() {
	if !fi.IsDeleted() {
		fi.History = append(fi.History, &FileEvent{
			Path: fi.Path(),
			Type: deleted,
			Time: time.Now().UTC(),
		})
	}
}

//       dP dP
//       88 88
// .d888b88 88d888b.
// 88'  `88 88'  `88
// 88.  .88 88.  .88
// `88888P8 88Y8888'

type db struct {
	dbDir        string
	absBaseDir   string
	absImportDir string

	// this is simply kept for saving purposes
	baseDir   string
	importDir string
	files     []*FileInfo
}

// GetDbDir ...
func (db *db) GetDbDir() string {
	return db.dbDir
}

// GetBaseDir ...
func (db *db) GetBaseDir() string {
	return db.absBaseDir
}

// GetImportDir ...
func (db *db) GetImportDir() string {
	return db.absImportDir
}

// GetFiles ...
func (db *db) GetFiles() []*FileInfo {
	return append([]*FileInfo{}, db.files...)
}

func (db *db) Sort() {
	sort.Slice(db.files, func(i, j int) bool {
		return db.files[i].Path() < db.files[j].Path()
	})
}

func filesToPathMap(files []*FileInfo) map[string]*FileInfo {
	fileMap := make(map[string]*FileInfo)

	for _, file := range files {
		if !file.IsDeleted() {
			fileMap[file.Path()] = file
		}
	}

	return fileMap
}

// filesToHashMap ...
func filesToHashMap(files []*FileInfo) map[string][]*FileInfo {
	fileMap := make(map[string][]*FileInfo)

	for _, file := range files {
		if !file.IsDeleted() {
			fi, found := fileMap[file.Checksum()]
			if found {
				fileMap[file.Checksum()] = append(fi, file)
			} else {
				fileMap[file.Checksum()] = []*FileInfo{file}
			}
		}
	}

	return fileMap
}

// filesToHistoricHashMap ...
func filesToHistoricHashMap(files []*FileInfo) map[string][]int {
	fileMap := make(map[string][]int)

	for fileIndex, file := range files {
		// for _, event := range file.History[:len(file.History)-1] {
		for _, event := range file.History {
			if event.Type != deleted {
				fi, found := fileMap[event.Checksum]
				// does the checksum exist in the list
				if found {
					found = false
					// do not same file multiple times
					for _, otherFile := range fi { // optimisation; no need to include last hash as current hashes are already processes
						if fileIndex == otherFile {
							found = true
							break
						}
					}
					if !found {
						fileMap[event.Checksum] = append(fi, fileIndex)
					}
				} else {
					fileMap[event.Checksum] = []int{fileIndex}
				}
			}
		}
	}

	return fileMap
}

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

// func (d *db) Update(filter FilterFunc) error {
// 	if filter == nil {
// 		filter = CheckIfMetaChanged
// 	}
//
// 	dir := d.GetBaseDir()
//
// 	info, err := os.Stat(dir)
// 	if err != nil {
// 		return fmt.Errorf("base directory '%s' does not exist", dir)
// 	}
// 	if !info.IsDir() {
// 		return fmt.Errorf("base directory '%s' is not a directory", dir)
// 	}
//
// 	// TODO: can you optimise this into another loop?
// 	for _, localFile := range d.files {
// 		localFile.checked = false
// 	}
//
// 	localByPath := filesToPathMap(d.files)
//
// 	checkedFiles := &db{
// 		dbDir:        d.dbDir,
// 		absBaseDir:   d.absBaseDir,
// 		absImportDir: d.absImportDir,
// 		baseDir:      d.baseDir,
// 		importDir:    d.importDir,
// 		files:        []*FileInfo{},
// 	}
//
// 	// # get list of files that should be checked
// 	// - for each file on the file system
// 	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			if os.IsPermission(err) {
// 				log.Printf("%s: permission denied", path)
// 			} else {
// 				return fmt.Errorf("%s: error getting file info: %s", path, err)
// 			}
// 		}
// 		if info.IsDir() {
// 			if info.Name() == defaultDbDir { // skip DB directory
// 				// fmt.Printf("skip %s\n", path)
// 				return filepath.SkipDir
// 			} else if strings.HasPrefix(info.Name(), ".") {
// 				// fmt.Printf("skip %s\n", path)
// 				return filepath.SkipDir
// 			}
// 			// fmt.Printf("dir %s\n", path)
// 			return nil
// 		}
//
// 		root := path[:len(dir)]
// 		if dir != root {
// 			// this should never happen
// 			log.Panicf("unexpected error; root mismatch '%s' != '%s'", dir, root)
// 		}
//
// 		relPath := path[len(dir)+1:]
//
// 		//   - if the path matches exactly
// 		//     - if forced check or if size/date changed
// 		//       - put file in to-be-checked list
// 		//     - else
// 		//       - mark checked
// 		//   - else
// 		//     - put file in to-be-checked list
// 		localFile, ok := localByPath[relPath]
// 		if filter(info, localFile) {
// 			hash, err := CalculateChecksum(path)
// 			if err != nil {
// 				return err
// 			}
// 			log.Printf("%s: %s\n", hash, relPath)
//
// 			checkedFiles.files = append(checkedFiles.files, &FileInfo{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Type:     changed,
// 						Path:     relPath,
// 						Time:     info.ModTime(),
// 						Size:     info.Size(),
// 						Checksum: hash,
// 					},
// 				},
// 			})
// 		} else if ok {
// 			localFile.checked = true
// 			// fmt.Printf("=%s\n", localFile.Path())
// 			delete(localByPath, relPath)
// 		}
//
// 		return nil
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	return d.Diff2(checkedFiles, &updateAction{})
// }

// type updateAction struct {
// }
//
// func (a *updateAction) Equal(localFile, remoteFile *FileInfo) {
// 	fmt.Printf("=%s\n", localFile.Path())
// }
//
// func (a *updateAction) LocalOnly(localFile *FileInfo) {
// 	panic("local added should never happen for updateAction")
// }
//
// func (a *updateAction) RemoteOnly(remoteFile *FileInfo) {
// 	fmt.Printf("+%s\n", remoteFile.Path())
//
// 	localFile.History = append(localFile.History, &FileEvent{
// 		Type:     changed,
// 		Path:     remoteFile.Path(),
// 		Time:     remoteFile.Time(),
// 		Size:     remoteFile.Size(),
// 		Checksum: remoteFile.Checksum(),
// 	})
// }
//
// func (a *updateAction) LocalDeleted(localFile, remoteFile *FileInfo) {
// 	fmt.Printf("-%s\n", localFile.Path())
// 	localFile.markDeleted()
// }
//
// func (a *updateAction) RemoteDeleted(localFile, remoteFile *FileInfo) {
// 	panic("remote deleted should never happen for updateAction")
// }
//
// func (a *updateAction) LocalChanged(localFile, remoteFile *FileInfo) {
// 	panic("local changed should never happen for updateAction")
// }
//
// func (a *updateAction) RemoteChanged(localFile, remoteFile *FileInfo) {
// 	fmt.Printf("~%s => %s\n", localFile.Path(), remoteFile.Path())
// 	localFile.History = append(localFile.History, &FileEvent{
// 		Type:     changed,
// 		Path:     remoteFile.Path(),
// 		Time:     remoteFile.Time(),
// 		Size:     remoteFile.Size(),
// 		Checksum: remoteFile.Checksum(),
// 	})
// }
//
// func (a *updateAction) Conflict(localFile, remoteFile *FileInfo) {
// 	if localFile != nil {
// 		fmt.Printf("!%s\n", localFile.Path())
// 	}
// 	if remoteFile != nil {
// 		fmt.Printf("!%s\n", remoteFile.Path())
// 	}
// }

func (d *db) Diff3(remote Boffin, action DiffAction) error {
	localFiles := d.GetFiles()
	remoteFiles := remote.GetFiles()
	var err error

	localFiles, remoteFiles, err =
		matchRemoteToLocalUsingPathAndCurrentHashes(localFiles, remoteFiles, action)
		// equal
	localFiles, remoteFiles, err =
		matchRemoteToLocalUsingCurrentHashes(localFiles, remoteFiles, action)
		// moved/renamed
	localFiles, remoteFiles, err =
		matchCurrentRemoteToHistoricalLocalUsingHashes(localFiles, remoteFiles, action)
		// moved/renamed and changed; conflict if multiple matches
	localFiles, remoteFiles, err =
		matchCurrentLocalToHistoricalRemoteUsingHashed(localFiles, remoteFiles, action)
		// moved/renamed and changed; conflict if multiple matches
	localFiles, remoteFiles, err =
		matchUsingHistoricalHashes(localFiles, remoteFiles, action)
		// conflict
	localFiles, remoteFiles, err =
		matchUsingPath(localFiles, remoteFiles, action)
		// conflict

	for _, file := range localFiles {
		action.LocalOnly(file)
	}
	for _, file := range remoteFiles {
		action.RemoteOnly(file)
	}

	return err
}

// Match all files that have identical paths and current hashes and report them
// as equal/unchanged.
func matchRemoteToLocalUsingPathAndCurrentHashes(local, remote []*FileInfo, action DiffAction) (newLocal, newRemote []*FileInfo, err error) {
	// sort by path to merge lists easily
	sort.Slice(local, func(i, j int) bool {
		return local[i].Path() < local[j].Path()
	})
	sort.Slice(remote, func(i, j int) bool {
		return remote[i].Path() < remote[j].Path()
	})
	newLocal = make([]*FileInfo, 0, len(local))
	newRemote = make([]*FileInfo, 0, len(remote))

	i, j := 0, 0
	for {
		cmp := strings.Compare(local[i].Path(), remote[j].Path())
		// if paths are different just mark them for further processing
		if cmp < 0 {
			newLocal = append(newLocal, local[i])
			i++
			if i >= len(local) {
				break
			}
		} else if cmp > 0 {
			newRemote = append(newRemote, remote[j])
			j++
			if j >= len(local) {
				break
			}
		} else {
			// if paths match, are not deleted and checksums match, mark them equal
			if !local[i].IsDeleted() && !remote[j].IsDeleted() &&
				local[i].Checksum() == remote[j].Checksum() {
				action.Unchanged(local[i], remote[j])
			} else {
				newLocal = append(newLocal, local[i])
				newRemote = append(newRemote, remote[j])
			}

			i++
			if i >= len(local) {
				break
			}

			j++
			if j >= len(local) {
				break
			}
		}
	}

	// add any elements that might not have been processed by the loop, as often
	// one list is shorter than the other
	newLocal = append(newLocal, local[i:]...)
	newRemote = append(newRemote, remote[j:]...)

	return newLocal, newRemote, nil
}

// Match all files that have identical current hashes but different current
// paths, and mark them as moved/renamed. In case of multiple matches, report
// them as conflict.
func matchRemoteToLocalUsingCurrentHashes(local, remote []*FileInfo, action DiffAction) (newLocal, newRemote []*FileInfo, err error) {
	// copy all deleted files as we will not be handling them
	newLocal = make([]*FileInfo, 0, len(local))
	for _, file := range local {
		if file.IsDeleted() {
			newLocal = append(newLocal, file)
		}
	}
	newRemote = make([]*FileInfo, 0, len(remote))
	for _, file := range remote {
		if file.IsDeleted() {
			newRemote = append(newRemote, file)
		}
	}

	// maps do not include deleted files
	localByHash := filesToHashMap(local)
	remoteByHash := filesToHashMap(remote)

	for hash, localFiles := range localByHash {
		remoteFiles, match := remoteByHash[hash]
		if match {
			if len(localFiles) == 1 && len(remoteFiles) == 1 {
				action.Moved(localFiles[0], remoteFiles[0])
			} else {
				newLocal = append(newLocal, localFiles...)
				newRemote = append(newRemote, remoteFiles...)
			}

			delete(remoteByHash, hash)
		} else {
			newLocal = append(newLocal, localFiles...)
		}
	}

	for _, remoteFiles := range remoteByHash {
		newRemote = append(newRemote, remoteFiles...)
	}

	return newLocal, newRemote, nil
}

// Match all remote files to local files, where current remote hash matches
// historical local hash, and mark remote file as a changed version of the local
// file. In case that the same hash appears multiple times on either remote or
// local side, mark them as conflicts.
func matchCurrentRemoteToHistoricalLocalUsingHashes(local, remote []*FileInfo, action DiffAction) (newLocal, newRemote []*FileInfo, err error) {
	// copy all deleted files as we will not be handling them
	newLocal = make([]*FileInfo, 0, len(local))

	newRemote = make([]*FileInfo, 0, len(remote))
	for _, file := range remote {
		if file.IsDeleted() {
			newRemote = append(newRemote, file)
		}
	}

	localByHash := filesToHistoricHashMap(local)
	remoteByHash := filesToHashMap(remote)

	for remoteHash, remoteFiles := range remoteByHash {
		localFileIndices, ok := localByHash[remoteHash]
		if ok {
			if len(localFileIndices) == 1 && len(remoteFiles) == 1 {
				action.LocalChanged(local[localFileIndices[0]], remoteFiles[0])
				local[localFileIndices[0]] = nil
			} else {
				localFiles := make([]*FileInfo, 0, len(localFileIndices))
				for _, localFileIndex := range localFileIndices {
					localFiles = append(localFiles, local[localFileIndex])
					local[localFileIndex] = nil
				}
				action.Conflict(localFiles, remoteFiles)
			}
		} else {
			newRemote = append(newRemote, remoteFiles...)
		}
	}

	for _, localFile := range local {
		if localFile != nil {
			newLocal = append(newLocal, localFile)
		}
	}

	return newLocal, newRemote, nil
}

// Match all local files to remote files, where current local hash matches
// historical remote hash, and mark local file as a changed version of the
// remote file. In case that the same hash appears multiple times on either
// remote or local side, mark them as conflicts.
func matchCurrentLocalToHistoricalRemoteUsingHashed(local, remote []*FileInfo, action DiffAction) (newLocal, newRemote []*FileInfo, err error) {
	// copy all deleted files as we will not be handling them
	newLocal = make([]*FileInfo, 0, len(local))
	for _, file := range local {
		if file.IsDeleted() {
			newLocal = append(newLocal, file)
		}
	}

	newRemote = make([]*FileInfo, 0, len(remote))

	localByHash := filesToHashMap(local)
	remoteByHash := filesToHistoricHashMap(remote)

	for localHash, localFiles := range localByHash {
		remoteFileIndices, ok := remoteByHash[localHash]
		if ok {
			if len(remoteFileIndices) == 1 && len(localFiles) == 1 {
				action.RemoteChanged(localFiles[0], remote[remoteFileIndices[0]])
				remote[remoteFileIndices[0]] = nil
			} else {
				remoteFiles := make([]*FileInfo, 0, len(remoteFileIndices))
				for _, remoteFileIndex := range remoteFileIndices {
					remoteFiles = append(remoteFiles, remote[remoteFileIndex])
					remote[remoteFileIndex] = nil
				}
				action.Conflict(localFiles, remoteFiles)
			}
		} else {
			newLocal = append(newLocal, localFiles...)
		}
	}

	for _, remoteFile := range remote {
		if remoteFile != nil {
			newRemote = append(newRemote, remoteFile)
		}
	}

	return newLocal, newRemote, nil
}

func matchUsingHistoricalHashes(local, remote []*FileInfo, action DiffAction) (newLocal, newRemote []*FileInfo, err error) {
	newLocal = make([]*FileInfo, 0, len(local))
	newRemote = make([]*FileInfo, 0, len(remote))

	localByHash := filesToHistoricHashMap(local)
	remoteByHash := filesToHistoricHashMap(remote)

	for localHash, localFileIndices := range localByHash {
		remoteFileIndices, ok := remoteByHash[localHash]
		if ok {
			localFiles := make([]*FileInfo, 0, len(localFileIndices))
			for _, localFileIndex := range localFileIndices {
				if local[localFileIndex] != nil {
					localFiles = append(localFiles, local[localFileIndex])
					local[localFileIndex] = nil
				}
			}

			remoteFiles := make([]*FileInfo, 0, len(remoteFileIndices))
			for _, remoteFileIndex := range remoteFileIndices {
				if remote[remoteFileIndex] != nil {
					remoteFiles = append(remoteFiles, remote[remoteFileIndex])
					remote[remoteFileIndex] = nil
				}
			}

			action.Conflict(localFiles, remoteFiles)
		}
	}

	for _, localFile := range local {
		if localFile != nil {
			newLocal = append(newLocal, localFile)
		}
	}
	for _, remoteFile := range remote {
		if remoteFile != nil {
			newRemote = append(newRemote, remoteFile)
		}
	}

	return newLocal, newRemote, nil
}

func matchUsingPath(local, remote []*FileInfo, action DiffAction) (newLocal, newRemote []*FileInfo, err error) {
	newLocal = make([]*FileInfo, 0, len(local))
	for _, file := range local {
		if file.IsDeleted() {
			newLocal = append(newLocal, file)
		}
	}

	newRemote = make([]*FileInfo, 0, len(remote))
	for _, file := range remote {
		if file.IsDeleted() {
			newRemote = append(newRemote, file)
		}
	}

	localByPath := filesToPathMap(local)
	remoteByPath := filesToPathMap(remote)

	for localPath, localFile := range localByPath {
		remoteFile, ok := remoteByPath[localPath]
		if ok {
			action.Conflict([]*FileInfo{localFile}, []*FileInfo{remoteFile})
			delete(remoteByPath, localPath)
		} else {
			// pass through any unmatched files
			newLocal = append(newLocal, localFile)
		}
	}

	// pass through any unmatched files
	for _, remoteFile := range remoteByPath {
		newRemote = append(newRemote, remoteFile)
	}

	return newLocal, newRemote, nil
}

// func (d *db) Diff2(remote Boffin, action DiffAction) error {
// 	localByHash := filesToHashMap(d.files)
// 	remoteByHash := filesToHashMap(remote.GetFiles())
//
// 	// # match everything you can using hashes
// 	// - for each update
// 	for hash, remoteFiles := range remoteByHash {
// 		//   - get current files with matching hashes
// 		localFiles, found := localByHash[hash]
// 		// fmt.Printf("checking %s\n", hash)
// 		//
// 		// fmt.Printf("results\n")
// 		// for _, result := range results {
// 		// 	fmt.Printf("  %s\n", result.relPath)
// 		// }
// 		// fmt.Printf("localFiles\n")
// 		// for _, localFile := range localFiles {
// 		// 	fmt.Printf("  %s\n", localFile.Path())
// 		// }
// 		if found {
// 			// # prioritize matching of files with same meta data
// 			//   - for each current file in matching hashes
// 			//     - for each result in update
// 			//       - if whole path for existing file and result matches
// 			//         - mark as unchanged
// 			//         - mark existing file as checked
// 			//         - remove existing file from matching hashes
// 			//         - remove result from update
// 			lcount := 0
// 			// fmt.Printf("match by path\n")
// 			for _, localFile := range localFiles {
// 				if localFile.checked {
// 					panic(fmt.Sprintf("this should never happen:1: %s", localFile.Path()))
// 				}
// 				localFiles[lcount] = localFile
// 				lcount++
// 				rcount := 0
// 				for _, result := range remoteFiles {
// 					if result.checked {
// 						panic("this should never happen:2:")
// 					}
// 					if localFile.Path() == result.Path() {
// 						localFile.checked = true
// 						result.checked = true
// 						lcount-- // effectively undo the localFiles[lcount] = localFile
// 						action.Equal(localFile, result)
// 					} else {
// 						remoteFiles[rcount] = result
// 						rcount++
// 					}
// 				}
// 				remoteFiles = remoteFiles[:rcount]
// 				// fmt.Printf("remoteFiles:1:\n")
// 				// for _, result := range remoteFiles {
// 				// 	fmt.Printf("  %s\n", result.relPath)
// 				// }
// 			}
// 			localFiles = localFiles[:lcount]
// 			// fmt.Printf("localFiles:1:\n")
// 			// for _, localFile := range localFiles {
// 			// 	fmt.Printf("  %s\n", localFile.Path())
// 			// }
//
// 			// # prioritize matching of files with same name, i.e. moved files
// 			//   - for each current file in matching hashes
// 			//     - for each result in update
// 			//       - if filename for existing file and result matches
// 			//         - mark as moved
// 			//         - mark existing file as checked
// 			//         - remove existing file from matching hashes
// 			//         - remove result from update
// 			lcount = 0
// 			// fmt.Printf("match by file name\n")
// 			for _, localFile := range localFiles {
// 				if localFile.checked {
// 					panic("this should never happen:3:")
// 				}
// 				localFiles[lcount] = localFile
// 				lcount++
// 				rcount := 0
// 				for _, result := range remoteFiles {
// 					if result.checked {
// 						panic("this should never happen:4:")
// 					}
// 					if filepath.Base(localFile.Path()) == filepath.Base(result.Path()) {
// 						localFile.checked = true
// 						result.checked = true
// 						lcount-- // effectively undo the localFiles[lcount] = localFile
// 						action.RemoteChanged(localFile, result)
// 					} else {
// 						remoteFiles[rcount] = result
// 						rcount++
// 					}
// 				}
// 				remoteFiles = remoteFiles[:rcount]
// 				// fmt.Printf("remoteFiles:2:\n")
// 				// for _, result := range remoteFiles {
// 				// 	fmt.Printf("  %s\n", result.relPath)
// 				// }
// 			}
// 			localFiles = localFiles[:lcount]
// 			// fmt.Printf("localFiles:2:\n")
// 			// for _, localFile := range localFiles {
// 			// 	fmt.Printf("  %s\n", localFile.Path())
// 			// }
//
// 			// # simple case of file being renamed
// 			//   - if only one file in matching hashes and one result in updates
// 			//     - mark as moved
// 			//     - mark existing file as checked
// 			//     - remove existing file from matching hashes
// 			//     - remove result from update
// 			// fmt.Printf("single match\n")
// 			if len(localFiles) == 1 && len(remoteFiles) == 1 {
// 				localFile := localFiles[0]
// 				result := remoteFiles[0]
// 				if localFile.checked || result.checked {
// 					panic("this should never happen:5:")
// 				}
// 				localFile.checked = true
// 				result.checked = true
// 				action.RemoteChanged(localFile, result)
//
// 				localFiles = localFiles[:0]
// 				remoteFiles = remoteFiles[:0]
// 			}
//
// 			//   - if only updates remain
// 			// fmt.Printf("only updates\n")
// 			if len(localFiles) == 0 && len(remoteFiles) > 0 {
// 				//     - for every result in updates
// 				//       - mark as new
// 				//       - remove result from update
// 				for _, result := range remoteFiles {
// 					if result.checked {
// 						panic("this should never happen:6:")
// 					}
// 					result.checked = true
// 					// fi := &FileInfo{
// 					// 	checked: true,
// 					// }
// 					// d.files = append(d.files, fi)
// 					action.RemoteOnly(nil, result)
// 				}
// 				remoteFiles = remoteFiles[:0]
// 			} else {
// 				//   - else
// 				//     - for every result in updates
// 				//       - mark as conflict
// 				//       - remove result from update
// 				for _, result := range remoteFiles {
// 					if result.checked {
// 						panic("this should never happen:7:")
// 					}
// 					result.checked = true
// 					action.Conflict(nil, result)
// 					// TODO: is it safe to continue here?
// 				}
// 				remoteFiles = remoteFiles[:0]
// 				//     - for every current file in matching hashes
// 				//       - mark as conflict
// 				for _, localFile := range localFiles {
// 					if localFile.checked {
// 						panic("this should never happen:8:")
// 					}
// 					localFile.checked = true
// 					action.Conflict(localFile, nil)
// 					// TODO: is it safe to continue here?
// 				}
// 				localFiles = localFiles[:0]
// 			}
// 			// skip len(localFiles) > 0 && len(remoteFiles) == 0 as deletes will be handled below
// 		} // found
//
// 		// update map with only unprocessed remoteFiles
// 		remoteByHash[hash] = remoteFiles
//
// 		// no need to update localByHash with updated localFiles as localByHash is not used any more
// 	}
//
// 	// regenerate localByPath to account for any files matched by hash, i.e. moved
// 	localByPath := filesToPathMap(d.files)
//
// 	// # match everything you can using paths
// 	// - for each update
// 	//   - if existing file exists with the same path
// 	//     - mark as changed
// 	//     - mark as checked
// 	//   - else
// 	//     - mark as new
// 	// fmt.Printf("match by path\n")
// 	for _, remoteFiles := range remoteByHash {
// 		for _, result := range remoteFiles {
// 			if result.checked {
// 				panic("this should never happen:9:")
// 			}
// 			localFile, found := localByPath[result.Path()]
// 			if found {
// 				if localFile.checked {
// 					panic("this should never happen:10:")
// 				}
// 				localFile.checked = true
// 				action.RemoteChanged(localFile, result)
// 			} else {
// 				fmt.Printf("+%s\n", result.Path())
// 				fi := &FileInfo{
// 					checked: true,
// 				}
// 				d.files = append(d.files, fi)
// 				action.RemoteOnly(fi, result)
// 			}
// 		}
// 	}
//
// 	// every update has been addressed by now
//
// 	// # any file not checked means it was deleted
// 	// - for each existing file
// 	//   - if not checked
// 	//     - mark as deleted
// 	//     - mark as checked
// 	// fmt.Printf("unchecked\n")
// 	for _, localFile := range d.files {
// 		if !localFile.checked && !localFile.IsDeleted() {
// 			localFile.checked = true
// 			action.LocalDeleted(localFile, nil)
// 		}
// 	}
//
// 	return nil
// }

// const (
// 	// DiffUnchanged indicates that files in both repositories are identical.
// 	DiffUnchanged = iota
// 	// DiffLocalOnly indicates that file exists only in the local repository.
// 	DiffLocalOnly
// 	// DiffLocalChanged indicates that file has been changed in the local repository.
// 	DiffLocalChanged
// 	// DiffLocalDeleted indicates that file has been deleted in the local repository.
// 	DiffLocalDeleted
// 	// DiffRemoteOnly indicates that file exists only in the remote repository.
// 	DiffRemoteOnly
// 	// DiffRemoteChanged indicates that file has been changed in the remote repository.
// 	DiffRemoteChanged
// 	// DiffRemoteDeleted indicates that file has been deleted in the remote repository.
// 	DiffRemoteDeleted
// 	// DiffConflict indicates that file has changed in both repositories.
// 	DiffConflict
// )
//
// // DiffResult ...
// type DiffResult struct {
// 	Result int
// 	Local  *FileInfo
// 	Remote *FileInfo
// }
//
// type diffAction struct {
// 	results []DiffResult
// }
//
// func (a *diffAction) Equal(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffUnchanged,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) LocalOnly(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffLocalOnly,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) RemoteOnly(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffRemoteOnly,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) LocalDeleted(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffLocalDeleted,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) RemoteDeleted(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffRemoteDeleted,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) LocalChanged(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffLocalChanged,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) RemoteChanged(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffRemoteChanged,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }
//
// func (a *diffAction) Conflict(localFile, remoteFile *FileInfo) {
// 	a.results = append(a.results, DiffResult{
// 		Result: DiffConflict,
// 		Local:  localFile,
// 		Remote: remoteFile,
// 	})
// }

// // Diff ...
// func (db *db) Diff(remote Boffin) []DiffResult {
// 	action := &diffAction{
// 		results: make([]DiffResult, 0, len(db.files)+len(remote.GetFiles())),
// 	}
//
// 	db.Diff2(remote, action)
//
// 	// sort results according to local path names
// 	sort.Slice(action.results, func(i, j int) bool {
// 		var iFile string
// 		iResult := action.results[i]
// 		if iResult.Local != nil {
// 			iFile = iResult.Local.Path()
// 		} else {
// 			iFile = iResult.Remote.Path()
// 		}
//
// 		var jFile string
// 		jResult := action.results[j]
// 		if jResult.Local != nil {
// 			jFile = jResult.Local.Path()
// 		} else {
// 			jFile = jResult.Remote.Path()
// 		}
//
// 		return iFile < jFile
// 	})
//
// 	return action.results
// }

// DiffOld ...
// func (db *db) DiffOld(remote Boffin) []DiffResult {
// 	localFiles := make(map[string]*FileInfo)
// 	for _, localFile := range db.files {
// 		localFile.checked = false
// 		for _, localEvent := range localFile.History {
// 			if localEvent.Checksum != "" {
// 				localFiles[localEvent.Checksum] = localFile
// 			}
// 		}
// 	}
//
// 	// allocate for the worst case
// 	results := make([]DiffResult, 0, len(db.files)+len(remote.GetFiles()))
//
// 	// - for each file in the remote repo
// 	for _, remoteFile := range remote.GetFiles() {
// 		localFile, localFound := localFiles[remoteFile.Checksum()]
// 		//   - if not deleted and checksum exists in local repo
// 		if !remoteFile.IsDeleted() && localFound {
// 			//     - mark local file as checked
// 			if localFile.checked {
// 				printDiff(results)
// 				log.Panicf("already checked:%s:%s", localFile.Checksum(), localFile.Path())
// 			}
// 			localFile.checked = true
//
// 			//     - if match is current file version in local repo
// 			if remoteFile.IsDeleted() && localFile.IsDeleted() {
// 				results = append(results, DiffResult{
// 					Result: DiffUnchanged,
// 					Local:  localFile,
// 					Remote: remoteFile,
// 				})
// 				//     - else - if match is older version
// 			} else if localFile.IsDeleted() {
// 				results = append(results, DiffResult{
// 					Result: DiffLocalDeleted,
// 					Local:  localFile,
// 					Remote: remoteFile,
// 				})
// 			} else if remoteFile.Checksum() == localFile.Checksum() {
// 				results = append(results, DiffResult{
// 					Result: DiffUnchanged,
// 					Local:  localFile,
// 					Remote: remoteFile,
// 				})
// 				//     - else - if match is older version
// 			} else {
// 				results = append(results, DiffResult{
// 					Result: DiffLocalChanged,
// 					Local:  localFile,
// 					Remote: remoteFile,
// 				})
// 			}
// 		} else { // localFound
// 			//     - for each checksum in file history
// 			foundLocalMatch := false
// 			// TODO: comparing first checksum is pointless
// 			for i := range remoteFile.History {
// 				remoteEvent := remoteFile.History[len(remoteFile.History)-1-i]
// 				//       - if checksum exists in local repo
// 				localFile, localFound := localFiles[remoteEvent.Checksum]
// 				if localFound {
// 					//     - mark local file as checked
// 					if localFile.checked {
// 						printDiff(results)
// 						log.Panicf("already checked:%s:%s", localFile.Checksum(), localFile.Path())
// 					}
// 					localFile.checked = true
// 					//         - if match is current file version in local repo
// 					if remoteEvent.Checksum == localFile.Checksum() {
// 						if remoteFile.IsDeleted() {
// 							results = append(results, DiffResult{
// 								Result: DiffRemoteDeleted,
// 								Local:  localFile,
// 								Remote: remoteFile,
// 							})
// 						} else {
// 							results = append(results, DiffResult{
// 								Result: DiffRemoteChanged,
// 								Local:  localFile,
// 								Remote: remoteFile,
// 							})
// 						}
// 					} else if remoteFile.IsDeleted() && localFile.IsDeleted() {
// 						//         - else if both files deleted
// 						// TODO: merge histories?
// 						results = append(results, DiffResult{
// 							Result: DiffRemoteChanged,
// 							Local:  localFile,
// 							Remote: remoteFile,
// 						})
// 					} else { // both local and remote file have changes after the matched checksums
// 						//         - else - if match is older version and they are not both deleted
// 						results = append(results, DiffResult{
// 							Result: DiffConflict,
// 							Local:  localFile,
// 							Remote: remoteFile,
// 						})
// 					}
//
// 					foundLocalMatch = true
// 					break // stop on first match
// 				} // localFound
// 			} // for remote history
//
// 			// if none of the historic hashes found in local repo
// 			if !foundLocalMatch {
// 				//     - else - checksum not matched
// 				if remoteFile.IsDeleted() {
// 					// TODO: keep history?
// 					results = append(results, DiffResult{
// 						Result: DiffUnchanged,
// 						Remote: remoteFile,
// 					})
// 				} else {
// 					results = append(results, DiffResult{
// 						Result: DiffRemoteOnly,
// 						Remote: remoteFile,
// 					})
// 				}
// 			}
// 		}
// 	}
//
// 	// - for each file in the local repo
// 	//   - if not checked
// 	//     - if deleted
// 	//       - 'LocalDeleted'
// 	//     - else
// 	//       - 'LocalOnly'
// 	for _, localFile := range db.files {
// 		if !localFile.checked {
// 			if localFile.IsDeleted() {
// 				results = append(results, DiffResult{
// 					Result: DiffLocalDeleted,
// 					Local:  localFile,
// 				})
// 			} else {
// 				results = append(results, DiffResult{
// 					Result: DiffLocalOnly,
// 					Local:  localFile,
// 				})
// 			}
// 		}
// 	}
//
// 	// sort results according to local path names
// 	sort.Slice(results, func(i, j int) bool {
// 		var iFile string
// 		iResult := results[i]
// 		if iResult.Local != nil {
// 			iFile = iResult.Local.Path()
// 		} else {
// 			iFile = iResult.Remote.Path()
// 		}
//
// 		var jFile string
// 		jResult := results[j]
// 		if jResult.Local != nil {
// 			jFile = jResult.Local.Path()
// 		} else {
// 			jFile = jResult.Remote.Path()
// 		}
//
// 		return iFile < jFile
// 	})
//
// 	return results
// }

func cleanPath(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// dP                                dP      d8' .d88888b
// 88                                88     d8'  88.    "'
// 88        .d8888b. .d8888b. .d888b88    d8'   `Y88888b. .d8888b. dP   .dP .d8888b.
// 88        88'  `88 88'  `88 88'  `88   d8'          `8b 88'  `88 88   d8' 88ooood8
// 88        88.  .88 88.  .88 88.  .88  d8'     d8'   .8P 88.  .88 88 .88'  88.  ...
// 88888888P `88888P' `88888P8 `88888P8 88        Y88888P  `88888P8 8888P'   `88888P'

type jsonStruct struct {
	V1 *v1Struct `json:"v1,omitempty"`
}

type v1Struct struct {
	BaseDir   string      `json:"base-dir"`
	ImportDir string      `json:"import-dir"`
	Files     []*FileInfo `json:"files"`
}

// InitDbDir ...
func InitDbDir(dbDir, baseDir string) (Boffin, error) {
	baseDir, err := cleanPath(baseDir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("'%s' does not exist", baseDir)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", baseDir)
	}

	dbDir, err = cleanPath(dbDir)
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(dbDir)
	if err == nil {
		return nil, fmt.Errorf("'%s' already exists", dbDir)
	}
	err = os.Mkdir(dbDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	db := &db{
		dbDir:      dbDir,
		absBaseDir: baseDir,
	}

	if relDir, err := filepath.Rel(dbDir, baseDir); err == nil {
		// if we can deduce relative path, use it instead of absolute one
		db.baseDir = relDir
	} else {
		db.baseDir = baseDir
	}

	if err = db.Save(); err != nil {
		return nil, err
	}

	return db, nil
}

// Save ...
func (db *db) Save() error {
	rawJSON := &jsonStruct{
		V1: &v1Struct{
			BaseDir: db.baseDir,
			Files:   db.files,
		},
	}

	newFilename := filepath.Join(db.dbDir, newFilesFilename)
	defer os.Remove(newFilename) // cleanup; will work only if the old file could not be replaced

	{ // write new file
		file, err := os.OpenFile(newFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if encoder == nil {
			return fmt.Errorf("failed to create json encoder")
		}
		encoder.SetIndent("", "  ")

		encoder.Encode(rawJSON)
	}

	{ // now replace old file with the new one
		filename := filepath.Join(db.dbDir, filesFilename)

		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to overwrite '%s'", filename)
		}
		if err := os.Rename(newFilename, filename); err != nil {
			return fmt.Errorf("critical error; failed to rename '%s' to '%s'", newFilename, filename)
		}

		if fi, err := os.Stat(filename); err == nil {
			os.Chmod(filename, fi.Mode()&0444)
		} else {
			log.Printf("warning: failed to make repo file read only")
		}
	}

	return nil
}

// LoadBoffin ...
func LoadBoffin(dbDir string) (Boffin, error) {
	boffinPath := filepath.Join(dbDir, filesFilename)

	boffinFile, err := os.Open(boffinPath)
	if err != nil {
		return nil, err
	}
	defer boffinFile.Close()

	decoder := json.NewDecoder(boffinFile)

	rawJSON := &jsonStruct{}
	if err := decoder.Decode(&rawJSON); err != nil {
		return nil, err
	}

	// ensure there is nothing after the first json object
	dummy := &jsonStruct{}
	if err = decoder.Decode(&dummy); err != io.EOF {
		return nil, fmt.Errorf("unexpected contents at the end of config file")
	}

	if rawJSON.V1 == nil {
		return nil, fmt.Errorf("config file is empty")
	}

	db := &db{
		dbDir:     dbDir,
		baseDir:   rawJSON.V1.BaseDir,
		importDir: rawJSON.V1.ImportDir,
		files:     rawJSON.V1.Files,
	}

	if filepath.IsAbs(db.baseDir) {
		db.absBaseDir, err = cleanPath(db.baseDir)
	} else {
		db.absBaseDir, err = cleanPath(filepath.Join(dbDir, db.baseDir))
	}
	if filepath.IsAbs(db.importDir) {
		db.absImportDir, err = cleanPath(db.importDir)
	} else {
		db.absImportDir, err = cleanPath(filepath.Join(db.absBaseDir, db.importDir))
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ConstuctDbPath ...
func ConstuctDbPath(baseDir string) string {
	return filepath.Join(baseDir, defaultDbDir)
}

// FindBoffinDir ...
func FindBoffinDir(dir string) (string, error) {
	// if dir is empty, start in current directory
	if dir == "" {
		var err error
		if dir, err = os.Getwd(); err != nil {
			return "", err
		}
	}

	dir, err := cleanPath(dir)
	if err != nil {
		return "", err
	}

	// look into current or any parent directory for a root which has defaultDbDir
	for true {
		dbDir := filepath.Join(dir, defaultDbDir)
		info, err := os.Stat(dbDir)
		if err == nil && info.IsDir() {
			return dbDir, nil
		}
		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("could not find %s dir", defaultDbDir)
}

func CalculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// func printDiff(diffs []DiffResult) {
// 	for _, diff := range diffs {
// 		if diff.Result == DiffUnchanged {
// 			if diff.Local != nil {
// 				fmt.Printf("=:%s\n", diff.Local.Path())
// 			} else {
// 				fmt.Printf("=:%s\n", diff.Remote.Path())
// 			}
// 		} else if diff.Result == DiffLocalOnly {
// 			fmt.Printf("L:%s\n", diff.Local.Path())
// 		} else if diff.Result == DiffRemoteOnly {
// 			fmt.Printf("R:%s\n", diff.Remote.Path())
// 		} else if diff.Result == DiffLocalDeleted {
// 			fmt.Printf("+:%s\n", diff.Local.Path())
// 		} else if diff.Result == DiffRemoteDeleted {
// 			fmt.Printf("-:%s\n", diff.Local.Path())
// 		} else if diff.Result == DiffLocalChanged {
// 			fmt.Printf("<:%s\n", diff.Local.Path())
// 		} else if diff.Result == DiffRemoteChanged {
// 			fmt.Printf(">:%s\n", diff.Local.Path())
// 		} else if diff.Result == DiffConflict {
// 			fmt.Printf("~:%s\n", diff.Local.Path())
// 		} else {
// 			fmt.Printf("~:%s\n", diff.Local.Path())
// 		}
// 	}
// }
