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
		channel := make(chan string)
		go readFile(channel, *inFile)
		file.merge(channel, true)
	}

	if *modify != "" {
		channel := make(chan string)
		go readModify(channel, *modify, *delimiter)
		file.merge(channel, false)
	}

	output := strings.Join(file.render(), *newline) + *newline

	if *outFile == "" {
		fmt.Print(output)
	} else {
		writeFile(*outFile, output)
	}
}

func readFile(channel chan<- string, filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(2)
	}
	for _, line := range strings.Split(string(content), "\n") {
		channel <- line
	}
	close(channel)
}

func readModify(channel chan<- string, modify string, delimiter string) {
	for _, line := range strings.Split(modify, "\n") {
		for _, line := range strings.Split(line, delimiter) {
			channel <- line
		}
	}
	close(channel)
}

func writeFile(filename string, contents string) {
	err := os.WriteFile(filename, []byte(contents), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing file: %v\n", err)
		os.Exit(3)
	}
}
