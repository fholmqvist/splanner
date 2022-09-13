# SPLANNER

A [suckless](https://suckless.org/) daily planner written in Go.

Creates a new markdown file for the day, appending any unfinished tasks from
yesterday. Opens the file with `xdg-open`.

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
sudo make bin
```

If you are getting permission denied when running splanner,
try running it with `sudo`.
