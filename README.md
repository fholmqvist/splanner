# SPLANNER

A [suckless](https://suckless.org/) daily planner written in Go.

Creates a new markdown file for the day, appending any unfinished tasks from
yesterday. 

Opens the file with `xdg-open`.

## FORMAT

```
[ ] Unfinished task

[x] Finished task

[ ] Multiple
    lines
```

## REQUIREMENTS

* [Go](https://www.go.dev/)
* [Make](https://www.gnu.org/software/make/)

## INSTALL

```
sudo make install
```

If you don't have Go as root, try:

```
make build
sudo make copy
```
