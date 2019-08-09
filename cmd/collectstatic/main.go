package main

import (
	"flag"
	"fmt"
	"github.com/catcombo/go-staticfiles"
	"os"
)

type dirSliceValue []string

func (s *dirSliceValue) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *dirSliceValue) Set(value string) error {
	if _, err := os.Stat(value); err != nil {
		return err
	}

	*s = append(*s, value)
	return nil
}

func main() {
	var outputDir string
	var inputDirs []string

	flag.StringVar(&outputDir, "output", "", "Output directory (required)")
	flag.Var((*dirSliceValue)(&inputDirs), "input", "Input directory(ies)")
	flag.Parse()

	if outputDir == "" {
		fmt.Println("Output directory required")
		flag.Usage()
		os.Exit(2)
	}

	storage := staticfiles.NewStorage(outputDir)
	storage.SetVerboseOutput(true)

	for _, dir := range inputDirs {
		storage.AddInputDir(dir)
	}

	err := storage.CollectStatic()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
