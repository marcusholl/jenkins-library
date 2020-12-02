package main

import (
	"fmt"
	pipergit "github.com/SAP/jenkins-library/pkg/git"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	fmt.Printf("Hello Marcus\n")
	r, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("ERR PlainOpen: %v\n", err)
		return
	}
	cIter, err := pipergit.LogRange(r, "github/master", "HEAD")
	if err != nil {
		fmt.Printf("ERR Log: %v\n", err)
		return
	}
	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
	if err != nil {
		fmt.Printf("ERR Iter: %v\n", err)
		return
	}

}
