package main

import (
	"fmt"
	"os"

	"jeskin.net/gitssg/repo"

	"flag"

	"jeskin.net/gitssg/index"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: gitssg [-t template] cmd [repodir]

Commands:
	index
	repo
`)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
	}

	switch args[0] {
	default:
		usage()

	case "repo":
		if len(args) != 2 {
			usage()
		}
		repo.Run(args[1])

	case "index":
		if len(args) < 2 {
			usage()
		}
		index.Run(args[1:])
	}
}
