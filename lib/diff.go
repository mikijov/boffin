package lib

import (
	"sort"
	"strings"
)

type DiffAction interface {
	Unchanged(localFile, remoteFile *FileInfo)
	MetaDataChanged(localFile, remoteFile *FileInfo)
	Moved(localFile, remoteFile *FileInfo)
	LocalOnly(localFile *FileInfo)
	LocalOld(localFile *FileInfo)
	RemoteOnly(remoteFile *FileInfo)
	RemoteOld(remoteFile *FileInfo)
	LocalDeleted(localFile, remoteFile *FileInfo)
	RemoteDeleted(localFile, remoteFile *FileInfo)
	LocalChanged(localFile, remoteFile *FileInfo)
	RemoteChanged(localFile, remoteFile *FileInfo)
	ConflictHash(localFile, remoteFile []*FileInfo)
	ConflictPath(localFile, remoteFile *FileInfo)
}

func Diff(local, remote Boffin, action DiffAction) error {
	localFiles := local.GetFiles()
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
		if file.IsDeleted() {
			action.LocalOld(file)
		} else {
			action.LocalOnly(file)
		}
	}
	for _, file := range remoteFiles {
		if file.IsDeleted() {
			action.RemoteOld(file)
		} else {
			action.RemoteOnly(file)
		}
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
	if len(local) > 0 && len(remote) > 0 {
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
				if j >= len(remote) {
					break
				}
			} else {
				// if paths match, are not deleted and checksums match, mark them equal
				if !local[i].IsDeleted() && !remote[j].IsDeleted() && local[i].Checksum() == remote[j].Checksum() {
					if local[i].Time() != remote[j].Time() {
						action.MetaDataChanged(local[i], remote[j])
					} else {
						action.Unchanged(local[i], remote[j])
					}
				} else {
					newLocal = append(newLocal, local[i])
					newRemote = append(newRemote, remote[j])
				}

				i++
				j++
				if i >= len(local) || j >= len(remote) {
					break
				}
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
	localByHash := FilesToHashMap(local)
	remoteByHash := FilesToHashMap(remote)

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
	remoteByHash := FilesToHashMap(remote)

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
				action.ConflictHash(localFiles, remoteFiles)
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

	localByHash := FilesToHashMap(local)
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
				action.ConflictHash(localFiles, remoteFiles)
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
			if len(localFileIndices) == 1 && len(remoteFileIndices) == 1 {
				localFileIndex := localFileIndices[0]
				remoteFileIndex := remoteFileIndices[0]
				if local[localFileIndex].IsDeleted() && remote[remoteFileIndex].IsDeleted() {
					action.Unchanged(local[localFileIndex], remote[remoteFileIndex])
					local[localFileIndex] = nil
					remote[remoteFileIndex] = nil
					continue
				}
			}

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

			action.ConflictHash(localFiles, remoteFiles)
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
			action.ConflictPath(localFile, remoteFile)
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
func FilesToHashMap(files []*FileInfo) map[string][]*FileInfo {
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
		for _, event := range file.History {
			if event.Checksum != "" {
				fi, found := fileMap[event.Checksum]
				// does the checksum exist in the list
				if found {
					found = false
					// ensure file is added only once
					for _, otherFile := range fi {
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
