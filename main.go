package main

import (
	"flag"
	"fmt"
	. "github.com/fergus-oakley/bump/pkg"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"log"
	"os"
)

var (
	dir   string
	help  bool
	minor bool
	major bool
)

func main() {
	flag.BoolVar(&help, "help", false, "")
	flag.StringVar(&dir, "dir", "", "root directory of the repository you want to bump the version for. By default uses present working directory.")
	flag.BoolVar(&minor, "minor", false, "increments the minor release version")
	flag.BoolVar(&major, "major", false, "increments the major release version")
	flag.Parse()

	if help {
		fmt.Printf("Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Printf("Description: command line tool to allow the current remote git tag version (format 'v0.0.0') for a given repo to be incremented. increment can be to the minor or major release if respective flags are passed, but by default will increment the bug fix release. \n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		panic(errors.Wrap(err, "error: unable to open repository"))
	}

	if err := BumpVersion(repo, major, minor); err != nil {
		log.Fatal(err, "Error: failed to bump version tag and push to remote repository")
	}
	fmt.Println("tag bumped and pushed to remote successfully.")
}
