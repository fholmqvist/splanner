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
	"bytes"
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
	curr, prev := lastTwoFiles()
	if !fileIsFromToday(curr) {
		prev = curr
		curr = dateToFilename(time.Now())
	}

	if !fileExists(PATH + curr) {
		createFile(PATH, curr, unfinishedTodos(PATH+prev))
	}

	_, err := exec.Command("xdg-open", "testing/"+curr).Output()
	if err != nil {
		panic(err)
	}
}

func lastTwoFiles() (string, string) {
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

	if len(dates) == 0 {
		return "", ""
	}

	if len(dates) == 1 {
		return dates[0], ""
	}

	return dates[len(dates)-1], dates[len(dates)-2]
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

func createFile(path, filename string, todos []byte) {
	file, err := os.Create(path + filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("# " + filename[:len(filename)-3] + "\n\n")

	for _, todo := range bytes.Split(todos, []byte("\n")) {
		file.Write(todo)
		file.Write([]byte("\n"))
	}
}

func unfinishedTodos(filepath string) []byte {
	bb, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	var remaining []byte
	for _, line := range bytes.Split(bb, []byte("\n")) {
		i := hasTodo(line)
		if i < 0 {
			continue
		}

		if !taskCompleted(line, i) {
			remaining = append(remaining, line...)
		}
	}

	return remaining
}

func hasTodo(line []byte) int {
	var inTodo bool
	var open int
	for i := 0; i < len(line); i++ {
		if line[i] == '[' {
			open = i
			inTodo = true
			continue
		}
		if inTodo && line[i] == ']' {
			return open
		}
	}

	return -1
}

func taskCompleted(line []byte, i int) bool {
	if line[i] == '[' &&
		bytes.ToLower(line)[i+1] == 'x' &&
		line[i+2] == ']' {
		return true
	}

	return false
}
