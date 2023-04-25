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

const LICENSE = `

Copyright (C) 2023  Fredrik Holmqvist

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

var (
	PATH = findOrCreatePath()
)

const (
	TITLE              = "SPLANNER v1.0"
	DATE_FORMAT        = "2006-01-02"
	FILE_FLAGS         = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	PERMISSIONS        = 0700
	SETTINGS_PATH      = ".settings"
	DEFAULT_FOLDER_KEY = "default_folder"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h":
			fallthrough
		case "--help":
			fmt.Println(TITLE + `

usage:
	-h, --help	this menu
	-l, --license	prints the license (GPLv3)
	-d, --default	sets the default folder path
	-c, --current	prints default folder path`)

		case "-l":
			fallthrough
		case "--license":
			fmt.Println(TITLE + LICENSE)

		case "-d":
			fallthrough
		case "--default":
			if len(os.Args) < 3 {
				fmt.Println("please provide a path:")
				fmt.Println("\tsplanner --default /some/path")
				os.Exit(3)
			}
			err := os.WriteFile(SETTINGS_PATH, []byte(DEFAULT_FOLDER_KEY+os.Args[2]+"/"), fs.FileMode(FILE_FLAGS))
			if err != nil {
				fmt.Println(err)
				os.Exit(3)
			}
			fmt.Println("successfully set default folder to " + os.Args[2])

		case "-c":
			fallthrough
		case "--current":
			fmt.Printf("%v=%v\n", DEFAULT_FOLDER_KEY, PATH)

		default:
			fmt.Printf("unrecognized command: %v\n", os.Args[1])
			os.Exit(3)
		}

		return
	}

	curr, prev := lastTwoFiles()
	if curr != dateToFilename(time.Now()) {
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

	if _, err := exec.Command("xdg-open", PATH+curr).Output(); err != nil {
		panic(err)
	}
}

func findOrCreatePath() string {
	var path string
	if fileExists(SETTINGS_PATH) {
		bb, err := os.ReadFile(SETTINGS_PATH)
		if err != nil {
			panic(err)
		}

		bbs := bytes.Split(bb, []byte("\n"))
		if len(bbs) < 1 {
			panic(fmt.Errorf("settings is empty"))
		}
		if !bytes.Contains(bbs[0], []byte(DEFAULT_FOLDER_KEY+"=")) {
			panic(fmt.Errorf("first setting isn't folder"))
		}

		path = string(bytes.Split(bbs[0], []byte("="))[1])
	} else {
		bb, err := exec.Command("bash", "-c", "echo $USER").Output()
		if err != nil {
			panic(err)
		}

		user := strings.Trim(string(bb), "\n")

		path = fmt.Sprintf("/home/%s/splanner/", user)

		mkdir := fmt.Sprintf("mkdir -p -m 755 %v", path)
		_, err = exec.Command("bash", "-c", mkdir).Output()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(
			SETTINGS_PATH,
			[]byte(fmt.Sprintf("%v=%v", DEFAULT_FOLDER_KEY, path)),
			fs.FileMode(FILE_FLAGS),
		)
		if err != nil {
			panic(err)
		}

		_, err = exec.Command("bash", "-c", "chmod 775 "+SETTINGS_PATH).Output()
		if err != nil {
			panic(err)
		}
	}

	if !fileExists(path) {
		mkdir := fmt.Sprintf("mkdir -p -m 755 %v", path)
		_, err := exec.Command("bash", "-c", mkdir).Output()
		if err != nil {
			panic(err)
		}
	}

	return path
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
		if len(name) != 13 || !strings.Contains(name, ".md") {
			return nil
		}

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

	lines := bytes.Split(bb, []byte("\n"))
	var remaining []byte
	for i, line := range lines {
		var todoIdx int

		switch STATE {
		case LOOKING:
			todoIdx = hasTodo(line)
			if todoIdx < 0 {
				continue
			}

			STATE = BEGIN
			fallthrough

		case BEGIN:
			if len(line) == 0 {
				continue
			}

			if !taskCompleted(line, todoIdx) {
				remaining = append(remaining, line...)
				STATE = BODY
			} else {
				STATE = LOOKING
			}

		case BODY:
			if len(line) == 0 {
				// This and next line is empty, look for new todo.
				if len(lines) < i+1 && len(lines[i+1]) == 0 {
					STATE = LOOKING
				}

				// Only this line is empty, look for more body.
				continue
			}

			// Line is new todo.
			if hasTodo(line) != -1 {
				remaining = append(remaining, '\n')
				STATE = BEGIN
				i--
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
