package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var version string

func main() {
	inFile := flag.String("input", "", "Path to input ini file.")
	outFile := flag.String("output", "", "Path to output ini file.")
	delimiter := flag.String("delimit", ";", "Split the modify string on this value.")
	modify := flag.String("modify", "", "Modifications to make. ex: [section1];prop1=value1;prop2=value2;[section2];prop1=value3.")
	newline := flag.String("newline", "\n", `Output uses this line ending. Escapes \r and \n are interpreted, ex: \r\n.`)
	showVersion := flag.Bool("version", false, "output the version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if err := checkArgs(*inFile, *modify, *delimiter); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
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
		for _, skipped := range file.merge(lines, true) {
			fmt.Fprintf(os.Stderr, "skipping line #%v: %v\n", skipped.number, skipped.text)
		}
	}

	if *modify != "" {
		// Unlike junk in an input file, a bad modification is a mistake by
		// the caller, so fail before writing anything.
		invalid := file.merge(readModify(*modify, *delimiter), false)
		for _, skipped := range invalid {
			fmt.Fprintf(os.Stderr, "invalid modification: %q\n", skipped.text)
		}
		if len(invalid) > 0 {
			os.Exit(1)
		}
	}

	lineEnding := unescapeNewline(*newline)
	output := strings.Join(file.render(), lineEnding) + lineEnding

	if *outFile == "" {
		fmt.Print(output)
	} else {
		if err := writeFile(*outFile, output); err != nil {
			fmt.Fprintf(os.Stderr, "error writing file: %v\n", err)
			os.Exit(3)
		}
	}
}

func checkArgs(inFile string, modify string, delimiter string) error {
	if inFile == "" && modify == "" {
		return errors.New("arguments -input and/or -modify are required")
	}
	if delimiter == "" {
		return errors.New("argument -delimit cannot be empty")
	}
	return nil
}

// utf8BOM is often written at the start of files by Windows editors.
var utf8BOM = string(rune(0xFEFF))

func readFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	text := strings.TrimPrefix(string(content), utf8BOM)
	return strings.Split(text, "\n"), nil
}

func readModify(modify string, delimiter string) []string {
	lines := []string{}
	for _, line := range strings.Split(modify, "\n") {
		lines = append(lines, strings.Split(line, delimiter)...)
	}
	return lines
}

// unescapeNewline lets -newline be given as the text \r\n, since most
// shells pass backslash sequences through literally.
func unescapeNewline(value string) string {
	return strings.NewReplacer(`\r`, "\r", `\n`, "\n").Replace(value)
}

// writeFile writes to a temp file in the same directory and renames it over
// the target, so a failed write never leaves a truncated file behind.
func writeFile(filename string, contents string) (err error) {
	// Replace the file a symlink points to, not the symlink itself.
	if resolved, resolveErr := filepath.EvalSymlinks(filename); resolveErr == nil {
		filename = resolved
	}

	mode := os.FileMode(0644)
	if info, statErr := os.Stat(filename); statErr == nil {
		mode = info.Mode().Perm()
	}

	temp, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			temp.Close()
			os.Remove(temp.Name())
		}
	}()

	if _, err = temp.WriteString(contents); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(temp.Name(), mode); err != nil {
		return err
	}
	return os.Rename(temp.Name(), filename)
}
