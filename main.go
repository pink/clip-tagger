package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("clip-tagger v0.1.0")
	if len(os.Args) < 2 {
		fmt.Println("Usage: clip-tagger <directory>")
		os.Exit(1)
	}
}
