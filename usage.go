package main

import "fmt"

func usage() {
	fmt.Print(`
	Go-Types
	List all declared types within a given Go file.
	Visit http://www.github.com/natdm/go-typelist for more detailed example useage.

    Usage:
		go-typelist -v (yields version)
        go-typelist file/path/gofile.go

`)
}
