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
	checked bool
}

// Boffin ...
type Boffin interface {
	GetFiles() []*FileInfo

	GetDbDir() string
	GetBaseDir() string
	GetImportDir() string

	Save() error
	Sort()
	Update() error
	Update2() error
	Diff(other Boffin) []DiffResult
	Diff2(remote Boffin) []DiffResult
}

// Checksum ...
func (fi *FileInfo) Checksum() string {
	// return fi.History[len(fi.History)-1].Checksum
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Checksum
		}
	}
	return ""
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

func (fi *FileInfo) inheritsFrom(checksum string) bool {
	for _, event := range fi.History {
		if event.Checksum == checksum {
			return true
		}
	}
	return false
}

func (fi *FileInfo) markDeleted() {
	if !fi.IsDeleted() {
		fi.History = append(fi.History, &FileEvent{
			Path: fi.Path(),
			Type: deleted,
			Time: time.Now().UTC(),
		})
	}
}

func (fi *FileInfo) update(path, relPath string, info os.FileInfo) error {
	fi.checked = true
	// fmt.Printf("checking %s\n", path)

	if !fi.IsDeleted() { // check size/time only if not marked as deleted
		if info.Size() == fi.Size() && info.ModTime().UTC() == fi.Time() {
			// size and time matches, assume no change
			return nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	hashStr := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	// fmt.Printf("Hash: %s\n", hashStr)

	if len(fi.History) == 0 || hashStr != fi.History[len(fi.History)-1].Checksum {
		fi.History = append(fi.History, &FileEvent{
			Path:     relPath,
			Size:     info.Size(),
			Type:     changed,
			Time:     info.ModTime().UTC(),
			Checksum: hashStr,
		})
	}

	return nil
}

//       dP dP
//       88 88
// .d888b88 88d888b.
// 88'  `88 88'  `88
// 88.  .88 88.  .88
// `88888P8 88Y8888'

type db struct {
	dbDir      string
	absBaseDir string

	// this is simply kept for saving purposes
	baseDir string
	files   []*FileInfo
}

// GetDbDir ...
func (db *db) GetDbDir() string {
	return db.dbDir
}

// GetBaseDir ...
func (db *db) GetBaseDir() string {
	return db.absBaseDir
}

func (db *db) GetImportDir() string {
	return filepath.Join(db.GetBaseDir(), "import")
}

func (db *db) GetFiles() []*FileInfo {
	return db.files
}

func (db *db) Sort() {
	sort.Slice(db.files, func(i, j int) bool {
		return db.files[i].Path() < db.files[j].Path()
	})
}

func filesToPathMap(files []*FileInfo) map[string]*FileInfo {
	fileMap := make(map[string]*FileInfo)

	for _, file := range files {
		if !file.checked && !file.IsDeleted() {
			fileMap[file.Path()] = file
		}
	}

	return fileMap
}

// FilesToHashMap ...
func FilesToHashMap(files []*FileInfo) map[string][]*FileInfo {
	fileMap := make(map[string][]*FileInfo)

	for _, file := range files {
		if !file.checked && !file.IsDeleted() {
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

// Update ...
func (db *db) Update() error {
	dir := db.GetBaseDir()

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("base directory '%s' does not exist", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("base directory '%s' is not a directory", dir)
	}

	// TODO: consider if this should deep copy if object should stay unchanged on error
	fileMap := filesToPathMap(db.files)

	// fmt.Printf("walking %s\n", dir)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%s: error getting file info: %s", path, err)
		}
		if info.IsDir() {
			if info.Name() == defaultDbDir { // skip DB directory
				return filepath.SkipDir
			}
			return nil
		}

		root := path[:len(dir)]
		if dir != root {
			// this should never happen
			return fmt.Errorf("unexpected error; root mismatch '%s' != '%s'", dir, root)
		}

		relPath := path[len(dir)+1:]

		file, ok := fileMap[relPath]
		if !ok {
			file = &FileInfo{}
			fileMap[relPath] = file
			db.files = append(db.files, file)
		}

		err = file.update(path, relPath, info)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, file := range db.files {
		if !file.checked {
			file.markDeleted()
		}
	}

	return nil
}

type toCheck struct {
	path    string
	relPath string
	fi      os.FileInfo
	hash    string
	matched bool // TODO: remove after debugging
}

func (db *db) Update2() error {
	dir := db.GetBaseDir()

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("base directory '%s' does not exist", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("base directory '%s' is not a directory", dir)
	}

	// TODO: can you optimise this into another loop?
	for _, localFile := range db.files {
		localFile.checked = false
	}

	filesByPath := filesToPathMap(db.files)

	filesToCheck := []*toCheck{}

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
			}
			// fmt.Printf("dir %s\n", path)
			return nil
		}

		root := path[:len(dir)]
		if dir != root {
			// this should never happen
			// fmt.Printf("err1 %s\n", path)
			return fmt.Errorf("unexpected error; root mismatch '%s' != '%s'", dir, root)
		}

		relPath := path[len(dir)+1:]

		//   - if the path matches exactly
		//     - if forced check or if size/date changed
		//       - put file in to-be-checked list
		//     - else
		//       - mark checked
		//   - else
		//     - put file in to-be-checked list
		localFile, ok := filesByPath[relPath]
		if !ok {
			// fmt.Printf("checking %s\n", relPath)
			filesToCheck = append(filesToCheck, &toCheck{
				path:    path,
				relPath: relPath,
				fi:      info,
			})
		} else if localFile.IsDeleted() {
			// fmt.Printf("checking %s\n", relPath)
			filesToCheck = append(filesToCheck, &toCheck{
				path:    path,
				relPath: relPath,
				fi:      info,
			})
		} else if info.Size() != localFile.Size() || !info.ModTime().Equal(localFile.Time()) {
			// fmt.Printf("checking %s\n", relPath)
			filesToCheck = append(filesToCheck, &toCheck{
				path:    path,
				relPath: relPath,
				fi:      info,
			})
		} else {
			localFile.checked = true
			// fmt.Printf("=%s\n", localFile.Path())
			delete(filesByPath, relPath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// # bulk calculate checksums
	// - for each file in the to-be-checked list
	//   - calculate checksum
	//   - put checksum in the 'updates' map[hash][]result

	updates := make(map[string][]*toCheck)

	for _, f := range filesToCheck {
		var err error
		f.hash, err = CalculateChecksum(f.path)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s\n", f.hash, f.relPath)

		hashes, found := updates[f.hash]
		if found {
			updates[f.hash] = append(hashes, f)
		} else {
			updates[f.hash] = []*toCheck{f}
		}
	}

	filesByHash := FilesToHashMap(db.files)

	// # match everything you can using hashes
	// - for each update
	for hash, results := range updates {
		//   - get current files with matching hashes
		localFiles, found := filesByHash[hash]
		// fmt.Printf("checking %s\n", hash)
		//
		// fmt.Printf("results\n")
		// for _, result := range results {
		// 	fmt.Printf("  %s\n", result.relPath)
		// }
		// fmt.Printf("localFiles\n")
		// for _, localFile := range localFiles {
		// 	fmt.Printf("  %s\n", localFile.Path())
		// }
		if found {
			// # prioritize matching of files with same meta data
			//   - for each current file in matching hashes
			//     - for each result in update
			//       - if whole path for existing file and result matches
			//         - mark as unchanged
			//         - mark existing file as checked
			//         - remove existing file from matching hashes
			//         - remove result from update
			lcount := 0
			// fmt.Printf("match by path\n")
			for _, localFile := range localFiles {
				if localFile.checked {
					panic(fmt.Sprintf("this should never happen:1: %s", localFile.Path()))
				}
				localFiles[lcount] = localFile
				lcount++
				rcount := 0
				for _, result := range results {
					if result.matched {
						panic("this should never happen:2:")
					}
					if localFile.Path() == result.relPath {
						localFile.checked = true
						result.matched = true
						lcount-- // effectively undo the localFiles[lcount] = localFile
						// fmt.Printf("=%s\n", localFile.Path())
					} else {
						results[rcount] = result
						rcount++
					}
				}
				results = results[:rcount]
				// fmt.Printf("results:1:\n")
				// for _, result := range results {
				// 	fmt.Printf("  %s\n", result.relPath)
				// }
			}
			localFiles = localFiles[:lcount]
			// fmt.Printf("localFiles:1:\n")
			// for _, localFile := range localFiles {
			// 	fmt.Printf("  %s\n", localFile.Path())
			// }

			// # prioritize matching of files with same name, i.e. moved files
			//   - for each current file in matching hashes
			//     - for each result in update
			//       - if filename for existing file and result matches
			//         - mark as moved
			//         - mark existing file as checked
			//         - remove existing file from matching hashes
			//         - remove result from update
			lcount = 0
			// fmt.Printf("match by file name\n")
			for _, localFile := range localFiles {
				if localFile.checked {
					panic("this should never happen:3:")
				}
				localFiles[lcount] = localFile
				lcount++
				rcount := 0
				for _, result := range results {
					if result.matched {
						panic("this should never happen:4:")
					}
					if filepath.Base(localFile.Path()) == filepath.Base(result.relPath) {
						localFile.checked = true
						result.matched = true
						lcount-- // effectively undo the localFiles[lcount] = localFile
						// fmt.Printf("~%s => %s\n", localFile.Path(), result.relPath)
						localFile.History = append(localFile.History, &FileEvent{
							Type:     changed,
							Path:     result.relPath,
							Time:     result.fi.ModTime(),
							Size:     result.fi.Size(),
							Checksum: result.hash,
						})
					} else {
						results[rcount] = result
						rcount++
					}
				}
				results = results[:rcount]
				// fmt.Printf("results:2:\n")
				// for _, result := range results {
				// 	fmt.Printf("  %s\n", result.relPath)
				// }
			}
			localFiles = localFiles[:lcount]
			// fmt.Printf("localFiles:2:\n")
			// for _, localFile := range localFiles {
			// 	fmt.Printf("  %s\n", localFile.Path())
			// }

			// # simple case of file being renamed
			//   - if only one file in matching hashes and one result in updates
			//     - mark as moved
			//     - mark existing file as checked
			//     - remove existing file from matching hashes
			//     - remove result from update
			// fmt.Printf("single match\n")
			if len(localFiles) == 1 && len(results) == 1 {
				localFile := localFiles[0]
				result := results[0]
				if localFile.checked || result.matched {
					panic("this should never happen:5:")
				}
				// fmt.Printf("~%s => %s\n", localFile.Path(), result.relPath)
				localFile.History = append(localFile.History, &FileEvent{
					Type:     changed,
					Path:     result.relPath,
					Time:     result.fi.ModTime(),
					Size:     result.fi.Size(),
					Checksum: result.hash,
				})
				localFile.checked = true
				result.matched = true

				localFiles = localFiles[:0]
				results = results[:0]
			}

			//   - if only updates remain
			// fmt.Printf("only updates\n")
			if len(localFiles) == 0 && len(results) > 0 {
				//     - for every result in updates
				//       - mark as new
				//       - remove result from update
				for _, result := range results {
					if result.matched {
						panic("this should never happen:6:")
					}
					result.matched = true
					// fmt.Printf("+%s\n", result.relPath)

					db.files = append(db.files, &FileInfo{
						History: []*FileEvent{
							&FileEvent{
								Type:     changed,
								Path:     result.relPath,
								Time:     result.fi.ModTime(),
								Size:     result.fi.Size(),
								Checksum: result.hash,
							},
						},
						checked: true, // new files are automatically checked
					})
				}
				results = results[:0]
			} else {
				//   - else
				//     - for every result in updates
				//       - mark as conflict
				//       - remove result from update
				for _, result := range results {
					if result.matched {
						panic("this should never happen:7:")
					}
					result.matched = true
					// fmt.Printf("!%s\n", result.relPath)
					// TODO: is it safe to continue here?
				}
				results = results[:0]
				//     - for every current file in matching hashes
				//       - mark as conflict
				for _, localFile := range localFiles {
					if localFile.checked {
						panic("this should never happen:8:")
					}
					localFile.checked = true
					// fmt.Printf("!%s\n", localFile.Path())
					// TODO: is it safe to continue here?
				}
				localFiles = localFiles[:0]
			}
			// skip len(localFiles) > 0 && len(results) == 0 as deletes will be handled below
		} // found

		// update map with only unprocessed results
		updates[hash] = results

		// no need to update filesByHash with updated localFiles as filesByHash is not used any more
	}

	// regenerate filesByPath to account for any files matched by hash, i.e. moved
	filesByPath = filesToPathMap(db.files)

	// # match everything you can using paths
	// - for each update
	//   - if existing file exists with the same path
	//     - mark as changed
	//     - mark as checked
	//   - else
	//     - mark as new
	// fmt.Printf("match by path\n")
	for _, results := range updates {
		for _, result := range results {
			if result.matched {
				panic("this should never happen:9:")
			}
			localFile, found := filesByPath[result.relPath]
			if found {
				if localFile.checked {
					panic("this should never happen:10:")
				}
				// fmt.Printf("~%s\n", localFile.Path())
				localFile.checked = true
				localFile.History = append(localFile.History, &FileEvent{
					Type:     changed,
					Path:     result.relPath,
					Time:     result.fi.ModTime(),
					Size:     result.fi.Size(),
					Checksum: result.hash,
				})
			} else {
				// fmt.Printf("+%s\n", result.relPath)
				db.files = append(db.files, &FileInfo{
					History: []*FileEvent{
						&FileEvent{
							Type:     changed,
							Path:     result.relPath,
							Time:     result.fi.ModTime(),
							Size:     result.fi.Size(),
							Checksum: result.hash,
						},
					},
					checked: true, // new files are automatically checked
				})
			}
		}
	}

	// every update has been addressed by now

	// # any file not checked means it was deleted
	// - for each existing file
	//   - if not checked
	//     - mark as deleted
	//     - mark as checked
	// fmt.Printf("unchecked\n")
	for _, localFile := range db.files {
		if !localFile.checked && !localFile.IsDeleted() {
			localFile.checked = true
			// fmt.Printf("-%s\n", localFile.Path())
			localFile.markDeleted()
		}
	}

	return nil
}

const (
	// DiffEqual indicates that files in both repositories are identical.
	DiffEqual = iota
	// DiffLocalAdded indicates that file exists only in the local repository.
	DiffLocalAdded
	// DiffLocalChanged indicates that file has been changed in the local repository.
	DiffLocalChanged
	// DiffLocalDeleted indicates that file has been deleted in the local repository.
	DiffLocalDeleted
	// DiffRemoteAdded indicates that file exists only in the remote repository.
	DiffRemoteAdded
	// DiffRemoteChanged indicates that file has been changed in the remote repository.
	DiffRemoteChanged
	// DiffRemoteDeleted indicates that file has been deleted in the remote repository.
	DiffRemoteDeleted
	// DiffConflict indicates that file has changed in both repositories.
	DiffConflict
)

// DiffResult ...
type DiffResult struct {
	Result int
	Local  *FileInfo
	Remote *FileInfo
}

// Diff ...
func (db *db) Diff(other Boffin) []DiffResult {
	otherFiles := other.GetFiles()

	// sort both file lists so that we can perform merge
	db.Sort()
	other.Sort()

	// allocate for the worst case
	results := make([]DiffResult, 0, len(db.files)+len(otherFiles))

	// calculate results by merging two lists
	i := 0
	j := 0
	for i < len(db.files) || j < len(otherFiles) {
		var localFile *FileInfo
		if i < len(db.files) {
			localFile = db.files[i]
		}

		var remoteFile *FileInfo
		if j < len(otherFiles) {
			remoteFile = otherFiles[j]
		}

		if remoteFile == nil || (localFile != nil && localFile.Path() < remoteFile.Path()) {
			results = append(results, DiffResult{
				Result: DiffLocalAdded,
				Local:  localFile,
			})
			i = i + 1
		} else if localFile == nil || (remoteFile == nil && localFile.Path() > remoteFile.Path()) {
			results = append(results, DiffResult{
				Result: DiffRemoteAdded,
				Remote: remoteFile,
			})
			j = j + 1
		} else {
			result := DiffResult{
				Local:  localFile,
				Remote: remoteFile,
			}
			i = i + 1
			j = j + 1

			if localFile.IsDeleted() && remoteFile.IsDeleted() {
				result.Result = DiffEqual
			} else if localFile.IsDeleted() {
				result.Result = DiffLocalDeleted
			} else if remoteFile.IsDeleted() {
				result.Result = DiffRemoteDeleted
			} else if localFile.Checksum() == remoteFile.Checksum() {
				result.Result = DiffEqual
			} else {
				rightNewer := remoteFile.inheritsFrom(localFile.Checksum())
				leftNewer := localFile.inheritsFrom(remoteFile.Checksum())

				if rightNewer && leftNewer {
					result.Result = DiffConflict
				} else if rightNewer {
					result.Result = DiffRemoteChanged
				} else if leftNewer {
					result.Result = DiffLocalChanged
				} else {
					result.Result = DiffConflict
				}
			}

			results = append(results, result)
		}
	}

	return results
}

// Diff2 ...
func (db *db) Diff2(remote Boffin) []DiffResult {
	localFiles := make(map[string]*FileInfo)
	for _, localFile := range db.files {
		localFile.checked = false
		for _, localEvent := range localFile.History {
			if localEvent.Checksum != "" {
				localFiles[localEvent.Checksum] = localFile
			}
		}
	}

	// allocate for the worst case
	results := make([]DiffResult, 0, len(db.files)+len(remote.GetFiles()))

	// - for each file in the remote repo
	for _, remoteFile := range remote.GetFiles() {
		localFile, localFound := localFiles[remoteFile.Checksum()]
		//   - if not deleted and checksum exists in local repo
		if !remoteFile.IsDeleted() && localFound {
			//     - mark local file as checked
			if localFile.checked {
				panic("already checked")
			}
			localFile.checked = true

			//     - if match is current file version in local repo
			if remoteFile.IsDeleted() && localFile.IsDeleted() {
				results = append(results, DiffResult{
					Result: DiffEqual,
					Local:  localFile,
					Remote: remoteFile,
				})
				//     - else - if match is older version
			} else if localFile.IsDeleted() {
				results = append(results, DiffResult{
					Result: DiffLocalDeleted,
					Local:  localFile,
					Remote: remoteFile,
				})
			} else if remoteFile.Checksum() == localFile.Checksum() {
				results = append(results, DiffResult{
					Result: DiffEqual,
					Local:  localFile,
					Remote: remoteFile,
				})
				//     - else - if match is older version
			} else {
				results = append(results, DiffResult{
					Result: DiffLocalChanged,
					Local:  localFile,
					Remote: remoteFile,
				})
			}
		} else { // localFound
			//     - for each checksum in file history
			foundLocalMatch := false
			// TODO: comparing first checksum is pointless
			for i := range remoteFile.History {
				remoteEvent := remoteFile.History[len(remoteFile.History)-1-i]
				//       - if checksum exists in local repo
				localFile, localFound := localFiles[remoteEvent.Checksum]
				if localFound {
					//     - mark local file as checked
					if localFile.checked {
						panic("already checked")
					}
					localFile.checked = true
					//         - if match is current file version in local repo
					if remoteEvent.Checksum == localFile.Checksum() {
						if remoteFile.IsDeleted() {
							results = append(results, DiffResult{
								Result: DiffRemoteDeleted,
								Local:  localFile,
								Remote: remoteFile,
							})
						} else {
							results = append(results, DiffResult{
								Result: DiffRemoteChanged,
								Local:  localFile,
								Remote: remoteFile,
							})
						}
					} else if remoteFile.IsDeleted() && localFile.IsDeleted() {
						//         - else if both files deleted
						// TODO: merge histories?
						results = append(results, DiffResult{
							Result: DiffRemoteChanged,
							Local:  localFile,
							Remote: remoteFile,
						})
					} else { // both local and remote file have changes after the matched checksums
						//         - else - if match is older version and they are not both deleted
						results = append(results, DiffResult{
							Result: DiffConflict,
							Local:  localFile,
							Remote: remoteFile,
						})
					}

					foundLocalMatch = true
					break // stop on first match
				} // localFound
			} // for remote history

			// if none of the historic hashes found in local repo
			if !foundLocalMatch {
				//     - else - checksum not matched
				if remoteFile.IsDeleted() {
					// TODO: keep history?
					results = append(results, DiffResult{
						Result: DiffEqual,
						Remote: remoteFile,
					})
				} else {
					results = append(results, DiffResult{
						Result: DiffRemoteAdded,
						Remote: remoteFile,
					})
				}
			}
		}
	}

	// - for each file in the local repo
	//   - if not checked
	//     - if deleted
	//       - 'LocalDeleted'
	//     - else
	//       - 'LocalAdded'
	for _, localFile := range db.files {
		if !localFile.checked {
			if localFile.IsDeleted() {
				results = append(results, DiffResult{
					Result: DiffLocalDeleted,
					Local:  localFile,
				})
			} else {
				results = append(results, DiffResult{
					Result: DiffLocalAdded,
					Local:  localFile,
				})
			}
		}
	}

	// sort results according to local path names
	sort.Slice(results, func(i, j int) bool {
		var iFile string
		iResult := results[i]
		if iResult.Local != nil {
			iFile = iResult.Local.Path()
		} else {
			iFile = iResult.Remote.Path()
		}

		var jFile string
		jResult := results[j]
		if jResult.Local != nil {
			jFile = jResult.Local.Path()
		} else {
			jFile = jResult.Remote.Path()
		}

		return iFile < jFile
	})

	return results
}

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
	BaseDir string      `json:"base-dir"`
	Files   []*FileInfo `json:"files"`
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
		dbDir:   dbDir,
		baseDir: rawJSON.V1.BaseDir,
		files:   rawJSON.V1.Files,
	}

	if filepath.IsAbs(db.baseDir) {
		db.absBaseDir, err = cleanPath(db.baseDir)
	} else {
		db.absBaseDir, err = cleanPath(filepath.Join(dbDir, db.baseDir))
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
