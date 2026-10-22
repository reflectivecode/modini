package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
)

// version is set by the release build with -ldflags "-X main.version=...".
var version string

// buildVersion falls back to the version Go records for "go install" and
// for builds from a tagged git checkout.
func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is main without the process-level side effects, so tests can check
// output and exit codes.
func run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("modini", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inFile := flags.String("input", "", "Path to input ini file.")
	outFile := flags.String("output", "", "Path to output ini file.")
	delimiter := flags.String("delimit", ";", "Split the modify string on this value.")
	modify := flags.String("modify", "", "Modifications to make. ex: [section1];prop1=value1;prop2=value2;[section2];prop1=value3.")
	newline := flags.String("newline", "\n", `Output uses this line ending. Escapes \r and \n are interpreted, ex: \r\n.`)
	showVersion := flags.Bool("version", false, "output the version")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if *showVersion {
		fmt.Fprintln(stdout, buildVersion())
		return 0
	}

	if err := checkArgs(*inFile, *modify, *delimiter); err != nil {
		fmt.Fprintf(stderr, "%v\n\n", err)
		flags.PrintDefaults()
		return 1
	}

	file := newFile()

	if *inFile != "" {
		lines, err := readFile(*inFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading file: %v\n", err)
			return 2
		}
		for _, skipped := range file.mergeFile(lines) {
			fmt.Fprintf(stderr, "skipping line #%v: %v\n", skipped.number, skipped.text)
		}
	}

	if *modify != "" {
		// Unlike junk in an input file, a bad modification is a mistake by
		// the caller, so fail before writing anything.
		invalid := file.applyModify(readModify(*modify, *delimiter))
		for _, skipped := range invalid {
			fmt.Fprintf(stderr, "invalid modification: %q\n", skipped.text)
		}
		if len(invalid) > 0 {
			return 1
		}
	}

	lineEnding := unescapeNewline(*newline)
	output := strings.Join(file.render(), lineEnding) + lineEnding

	if *outFile == "" {
		if _, err := io.WriteString(stdout, output); err != nil {
			fmt.Fprintf(stderr, "error writing output: %v\n", err)
			return 3
		}
	} else {
		if err := writeFile(*outFile, output); err != nil {
			fmt.Fprintf(stderr, "error writing file: %v\n", err)
			return 3
		}
	}
	return 0
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
