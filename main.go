package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
)

// version is set by the release build with -ldflags "-X main.version=...".
var version string

// getVersion falls back to the version Go records for "go install" and
// for builds from a tagged git checkout.
func getVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	inFile := flag.String("input", "", "Path to input ini file.")
	outFile := flag.String("output", "", "Path to output ini file.")
	delimiter := flag.String("delimit", ";", "Split the modify string on this value.")
	modify := flag.String("modify", "", "Modifications to make. ex: [section1];prop1=value1;prop2=value2;[section2];prop1=value3.")
	newline := flag.String("newline", "\n", `Output uses this line ending. Escapes \r and \n are interpreted, ex: \r\n.`)
	showVersion := flag.Bool("version", false, "output the version")

	flag.Parse()

	if *showVersion {
		fmt.Println(getVersion())
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
		if _, err := fmt.Print(output); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
			os.Exit(3)
		}
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
	// Reading UTF-16 as UTF-8 would mangle the file, so refuse it. This also
	// catches UTF-32 LE, whose BOM starts with the same bytes.
	if bytes.HasPrefix(content, []byte{0xFF, 0xFE}) || bytes.HasPrefix(content, []byte{0xFE, 0xFF}) {
		return nil, errors.New("UTF-16 files are not supported, convert the file to UTF-8")
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

// writeFile overwrites in place, which keeps the file's owner, permissions,
// hard links and bind mounts intact.
func writeFile(filename string, contents string) error {
	return os.WriteFile(filename, []byte(contents), 0644)
}
