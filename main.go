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
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

const (
	PATH        = "testing/"
	DATE_FORMAT = "2006_01_02"
)

func main() {
	listLatest()
}

func listLatest() {
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

	fmt.Println(latest)
}
