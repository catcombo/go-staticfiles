package main

import (
	"flag"
	"fmt"
	"github.com/catcombo/go-staticfiles"
	"os"
)

type arrayString []string

func (a *arrayString) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *arrayString) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	var outputDir string
	var inputDirs []string
	var ignorePatterns []string

	flag.StringVar(&outputDir, "output", "", "Output directory (required)")
	flag.Var((*arrayString)(&inputDirs), "input", "Input directory(ies)")
	flag.Var((*arrayString)(&ignorePatterns), "ignore", "Ignore files, directories, or paths matching glob-style pattern")
	flag.Parse()

	if outputDir == "" {
		fmt.Println("Output directory required")
		flag.Usage()
		os.Exit(2)
	}

	storage, err := staticfiles.NewStorage(outputDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	storage.Verbose = true

	for _, dir := range inputDirs {
		storage.AddInputDir(dir)
	}

	for _, pattern := range ignorePatterns {
		storage.AddIgnorePattern(pattern)
	}

	err = storage.CollectStatic()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
