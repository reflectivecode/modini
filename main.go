package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var version string

func main() {
	inFile := flag.String("input", "", "Path to input ini file.")
	outFile := flag.String("output", "", "Path to output ini file.")
	delimiter := flag.String("delimit", ";", "Split the modify string on this value.")
	modify := flag.String("modify", "", "Modifications to make. ex: [section1];prop1=value1;prop2=value2;[section2];prop1=value3.")
	newline := flag.String("newline", "\n", "Output uses this line ending.")
	showVersion := flag.Bool("version", false, "output the version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if *inFile == "" && *modify == "" {
		fmt.Fprintf(os.Stderr, "arguments -input and/or -modify are required\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	file := newFile()

	if *inFile != "" {
		lines, err := readFile(*inFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
			os.Exit(2)
		}
		file.merge(lines, true)
	}

	if *modify != "" {
		file.merge(readModify(*modify, *delimiter), false)
	}

	output := strings.Join(file.render(), *newline) + *newline

	if *outFile == "" {
		fmt.Print(output)
	} else {
		writeFile(*outFile, output)
	}
}

func readFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

func readModify(modify string, delimiter string) []string {
	lines := []string{}
	for _, line := range strings.Split(modify, "\n") {
		lines = append(lines, strings.Split(line, delimiter)...)
	}
	return lines
}

func writeFile(filename string, contents string) {
	err := os.WriteFile(filename, []byte(contents), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing file: %v\n", err)
		os.Exit(3)
	}
}
