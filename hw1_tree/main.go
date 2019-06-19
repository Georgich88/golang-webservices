package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out *os.File, path string, printFiles bool) error {

	return dirTreePrint(out, path, printFiles, 0, false)

}

func dirTreePrint(out *os.File, path string, printFiles bool, level int, continuePrint bool) error {

	directory, err := os.Open(path)
	if err != nil {
		return err
	}

	files, err := directory.Readdir(-1)
	directory.Close()
	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	size := len(files)
	for i, file := range files {

		if file.IsDir() || printFiles {
			if level > 0 {
				fmt.Fprint(out, strings.Repeat("\t", level-1))
				if continuePrint {
					fmt.Fprint(out, "│")
				}
				fmt.Fprint(out, "\t")
			}
			if i == size-1 {
				fmt.Fprint(out, "└───")
			} else {
				fmt.Fprint(out, "├───")
			}
			fmt.Fprintln(out, file.Name())

		}
		if file.IsDir() {
			dirPath := path + string(os.PathSeparator) + file.Name()
			err := dirTreePrint(out, dirPath, printFiles, level+1, i != size-1)
			if err != nil {
				return err
			}
		}
	}

	return nil

}
