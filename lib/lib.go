/*
Importer
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

import "io"
import "os"
import "fmt"
import "path/filepath"
import "crypto/sha256"

type filterFunc func(info os.FileInfo) (accepted bool, reason string)

func okFilter(info os.FileInfo) (accepted bool, reason string) {
	return true, ""
}

func getFiles(dir string, accept filterFunc) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir = filepath.Clean(dir)
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory.", dir)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("%s: error getting file info: %s\n", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if accept != nil {
			if ok, reason := accept(info); !ok {
				fmt.Printf("%s: skipping: %s\n", path, reason)
				return nil
			}
		}

		root := path[:len(dir)]
		if dir != root {
			fmt.Printf("Root mismatch '%s' != '%s'\n", dir, root)
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

		path = path[len(dir):]
		fmt.Printf("%x: %s\n", hash.Sum(nil), path)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Scan ...
func Scan() {
	getFiles("/home/miki/tmp", okFilter)
}
