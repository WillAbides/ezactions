package main

import (
	"path/filepath"

	"github.com/alecthomas/kong"
	testoutput "github.com/willabides/ezactions/internal"
)

var cli struct {
	Passthrough bool   `kong:"help='write test output to stdout'"`
	RootPath    string `kong:"default='.',help='root path for test packages'"`
	RootPkg     string `kong:"required,help='the package at root-path'"`
}

func main() {
	ctx := kong.Parse(&cli)
	rootPath, err := filepath.Abs(cli.RootPath)
	ctx.FatalIfErrorf(err)
}
