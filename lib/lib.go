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

// import "io"
import "os"
import "fmt"
import "path/filepath"

// import "crypto/sha256"
// import "encoding/json"
import "time"

// import "encoding/base64"

const defaultDbDir = ".aether"

// FileMeta ...
type FileMeta struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Time     time.Time `json:"time"`
	Checksum []string  `json:"checksum"`
}

// comment
const (
	Equal = iota
	LeftChanged
	RightChanged
	Conflict
)

// Compare ...
func Compare(left, right FileMeta) {
}

// FileDB ...
type FileDB interface {
	DbDir() string
	BaseDir() string
	Save() error
	Update() error
	IsChanged() bool
}

type fileDb struct {
	dbDir   string
	baseDir string
	changed bool
}

// CreateFileDB ...
func CreateFileDB(dbDir string, baseDir string) (FileDB, error) {
	dbDir, err := filepath.Abs(dbDir)
	if err != nil {
		return nil, err
	}
	dbDir = filepath.Clean(dbDir)
	if _, err := os.Stat(dbDir); err != nil {
		return nil, fmt.Errorf("'%s' already exists", dbDir)
	}

	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return nil, err
	}

	return &fileDb{
		dbDir:   dbDir,
		baseDir: dbDir,
		changed: true,
	}, nil
}

// DbDir ...
func (db *fileDb) DbDir() string {
	return db.dbDir
}

// BaseDir ...
func (db *fileDb) BaseDir() string {
	return db.baseDir
}

// Save ...
func (db *fileDb) Save() error {
	return fmt.Errorf("not implemented")
}

// Update ...
func (db *fileDb) Update() error {
	return fmt.Errorf("not implemented")
}

// IsChanged ...
func (db *fileDb) IsChanged() bool {
	return db.changed
}

func cleanPath(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// func ensureDir(dir string) string, error {
// 	dir, err := cleanPath(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	info, err := os.Stat(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !info.IsDir() {
// 		return nil, fmt.Errorf("'%s' is not a directory", dir)
// 	}
// 	return dir, nil
// }

// ConstuctDbDirPath ...
func ConstuctDbDirPath(dir string) string {
	return filepath.Join(dir, defaultDbDir)
}

// InitDbDir ...
func InitDbDir(dbDir, baseDir string) (FileDB, error) {
	dbDir, err := cleanPath(dbDir)
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

	baseDir, err = cleanPath(baseDir)
	if err != nil {
		return nil, err
	}
	if relDir, err := filepath.Rel(dbDir, baseDir); err == nil {
		// if we can deduce relative path, use it instead of absolute one
		baseDir = relDir
	}

	return &fileDb{
		dbDir:   dbDir,
		baseDir: baseDir,
	}, nil
}

// func GetDbDir(dbDir, baseDir string) string, error {
// 	if dbDir != "" {
// 		return ensureDir(dbDir)
// 	}
//
// 	for dbDir != nil {
// 		dbDir, err := filepath.Join(baseDir, defaultDbDir)
// 		if err == nil {
// 			return dbDir, nil
// 		}
// 		dbDir, _ = filepath.Split(dbDir)
// 	}
//
// 	return filepath.Join(dir, defaultDbDir), nil
// }
//
// func LoadFileDB(aetherDir string) FileDB, error {
// }

// type filterFunc func(info os.FileInfo) (accepted bool, reason string)
//
// func okFilter(info os.FileInfo) (accepted bool, reason string) {
// 	return true, ""
// }
//
// func getFiles(dir string, accept filterFunc) error {
// 	dir, err := filepath.Abs(dir)
// 	if err != nil {
// 		return err
// 	}
// 	dir = filepath.Clean(dir)
// 	info, err := os.Stat(dir)
// 	if err != nil {
// 		return err
// 	}
// 	if !info.IsDir() {
// 		return fmt.Errorf("'%s' is not a directory", dir)
// 	}
//
// 	encoder := json.NewEncoder(os.Stdout)
// 	if encoder == nil {
// 		return fmt.Errorf("failed to create json encoder")
// 	}
// 	encoder.Encode('[')
//
// 	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			fmt.Printf("%s: error getting file info: %s\n", path, err)
// 			return nil
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}
//
// 		if accept != nil {
// 			if ok, reason := accept(info); !ok {
// 				fmt.Printf("%s: skipping: %s\n", path, reason)
// 				return nil
// 			}
// 		}
//
// 		root := path[:len(dir)]
// 		if dir != root {
// 			// TODO: just checking if the beginning of the path matches expectation
// 			fmt.Printf("Root mismatch '%s' != '%s'\n", dir, root)
// 		}
//
// 		file, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer file.Close()
//
// 		hash := sha256.New()
// 		if _, err := io.Copy(hash, file); err != nil {
// 			return err
// 		}
//
// 		path = path[len(dir)+1:]
//
// 		meta := &FileMeta{
// 			Name:     path,
// 			Size:     info.Size(),
// 			Time:     info.ModTime().UTC(),
// 			Checksum: []string{base64.StdEncoding.EncodeToString(hash.Sum(nil))},
// 		}
//
// 		encoder.Encode(meta)
// 		// // buf, err := json.MarshalIndent(c, "", "  ")
// 		// buf, err := json.Marshal(meta)
// 		// if err != nil {
// 		// 	return err
// 		// }
// 		//
// 		// fmt.Printf("%s\n", string(buf))
//
// 		return nil
// 	})
//
// 	if err != nil {
// 		return err
// 	}
//
// 	encoder.Encode(']')
//
// 	return nil
// }

// // Scan ...
// func Scan(dir string) error {
//
// 	db := LoadFileDB(dir)
// }
