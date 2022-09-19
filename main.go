/*
	SPLANNER

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
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	PATH = setPath()
)

const (
	DATE_FORMAT = "2006_01_02"
	FILE_FLAGS  = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	PERMISSIONS = 0700
)

func main() {
	createPathIfEmpty()

	curr, prev := lastTwoFiles()
	if !fileIsFromToday(curr) {
		prev = curr
		curr = dateToFilename(time.Now())
	}

	if !fileExists(PATH + curr) {
		var unfinished []byte
		if prev != "" {
			unfinished = unfinishedTodos(PATH + prev)
		}

		createFile(PATH, curr, unfinished)
	}

	openInEditor(PATH + curr)
}

func setPath() string {
	bb, err := exec.Command("bash", "-c", "echo $USER").Output()
	if err != nil {
		panic(err)
	}

	user := strings.Trim(string(bb), "\n")

	return fmt.Sprintf("/home/%s/splanner/", user)
}

func createPathIfEmpty() {
	if fileExists(PATH) {
		return
	}

	mkdir := fmt.Sprintf("mkdir -p -m 755 %v", PATH)

	_, err := exec.Command("bash", "-c", mkdir).Output()
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

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			panic(err)
		}
	}

	return true
}

func createFile(path, filename string, todos []byte) {
	file, err := os.OpenFile(path+filename, FILE_FLAGS, PERMISSIONS)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("# " + filename[:len(filename)-3] + "\n")

	if len(todos) > 0 {
		file.WriteString("\n")

		for _, todo := range bytes.Split(todos, []byte("\n")) {
			file.Write(todo)
			file.Write([]byte("\n"))
		}
	}
}

func unfinishedTodos(filepath string) []byte {
	if !fileExists(filepath) {
		return []byte{}
	}

	bb, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	const (
		LOOKING = 0
		BEGIN   = 1
		BODY    = 2
	)
	STATE := LOOKING

	var remaining []byte
	for _, line := range bytes.Split(bb, []byte("\n")) {
		var i int

		switch STATE {
		case LOOKING:
			i = hasTodo(line)
			if i < 0 {
				continue
			}

			STATE = BEGIN
			fallthrough

		case BEGIN:
			if len(line) == 0 {
				continue
			}

			if !taskCompleted(line, i) {
				remaining = append(remaining, line...)
				STATE = BODY
			}

		case BODY:
			if len(line) == 0 {
				STATE = LOOKING
				continue
			}

			remaining = append(remaining, '\n')
			remaining = append(remaining, line...)
		}
	}

	return remaining
}

func hasTodo(line []byte) int {
	var (
		inTodo bool
		open   int
	)

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

func openInEditor(filepath string) {
	if _, err := exec.Command("xdg-open", filepath).Output(); err != nil {
		panic(err)
	}
}
