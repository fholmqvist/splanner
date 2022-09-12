/*
	DAILY PLANNER (WIP)

	Copyright (C) 2022  Fredrik Holmqvist

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

package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	PATH        = "testing/"
	DATE_FORMAT = "2006_01_02"
)

func main() {
	filename := latestFilename()
	if !fileIsFromToday(filename) {
		filename = dateToFilename(time.Now())
	}

	if !fileExists(PATH + filename) {
		createFile(PATH, filename)
	}

	_, err := exec.Command("xdg-open", "testing/"+filename).Output()
	if err != nil {
		panic(err)
	}
}

func latestFilename() string {
	var dates []string

	err := filepath.Walk(PATH, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if i.IsDir() {
			return nil
		}

		name := i.Name()
		_, err = time.Parse(DATE_FORMAT, name[:10])
		if err != nil {
			return err
		}

		dates = append(dates, name)

		return nil
	})

	if err != nil {
		panic(err)
	}

	latest := dates[len(dates)-1]

	return latest
}

func fileIsFromToday(filename string) bool {
	return filename == dateToFilename(time.Now())
}

func dateToFilename(t time.Time) string {
	return t.Format(DATE_FORMAT) + ".md"
}

// Excluding path.
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}
	return true
}

func createFile(path, filename string) {
	file, err := os.Create(path + filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("# " + filename[:len(filename)-3])
}
