package main

import (
	"fmt"
	"os"
	"sort"
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

	zeroLevelInfo := []bool{}
	return dirTreePrint(out, path, printFiles, zeroLevelInfo)

}

func dirTreePrint(out *os.File, path string, printFiles bool, levelInfo []bool) error {

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
			if len(levelInfo) > 0 {

				for _, level := range levelInfo {
					if level {
						fmt.Fprint(out, "│")
						fmt.Fprint(out, "\t")
					} else {
						fmt.Fprint(out, "\t")
					}
				}

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
			err := dirTreePrint(out, dirPath, printFiles, append(levelInfo, i != size-1))
			if err != nil {
				return err
			}
		}
	}

	return nil

}
