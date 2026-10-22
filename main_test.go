package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestReadModifySplitsOnDelimiterAndNewlines(t *testing.T) {
	actual := readModify("[S];A=1\nB=2|C=3", ";")

	expected := []string{"[S]", "A=1", "B=2|C=3"}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestReadModifyCustomDelimiter(t *testing.T) {
	actual := readModify("[S]|A=1;2|B=3", "|")

	expected := []string{"[S]", "A=1;2", "B=3"}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestReadFileHandlesCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.ini")
	if err := os.WriteFile(path, []byte("[S]\r\nA=1\r\n"), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}

	file := newFile()
	file.mergeFile(lines)
	actual := file.render()

	expected := []string{"[S]", "A=1"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestReadFileStripsBOM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.ini")
	content := strings.Join([]string{utf8BOM + "[Main]", "key=1", ""}, "\n")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}

	file := newFile()
	file.mergeFile(lines)
	file.applyModify(readModify("[Main];key=2", ";"))
	actual := file.render()

	expected := []string{"[Main]", "key=2"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestReadFileRejectsUTF16(t *testing.T) {
	boms := map[string][]byte{
		"little endian": {0xFF, 0xFE},
		"big endian":    {0xFE, 0xFF},
	}

	for name, bom := range boms {
		path := filepath.Join(t.TempDir(), "input.ini")
		content := append(bom, []byte("[S]")...)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}

		if _, err := readFile(path); err == nil {
			t.Errorf(`%v: expected an error for a UTF-16 file`, name)
		}
	}
}

func TestReadFileMissing(t *testing.T) {
	_, err := readFile(filepath.Join(t.TempDir(), "missing.ini"))
	if err == nil {
		t.Error("expected an error for a missing file")
	}
}

// TestReadmeExample keeps the example in README.md honest.
func TestReadmeExample(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.ini")
	input := strings.Join([]string{
		"[Section1]",
		"Prop1=dog",
		"Prop2=cat",
		"PropList=tuna",
		"PropList=trout",
		"PropList=flounder",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	file := newFile()

	lines, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file.mergeFile(lines)

	file.applyModify(readModify("[Section1];Prop2=frog;Prop3=bear;PropList+=shark;PropList-=trout;[Section2];New=unicorn", ";"))

	expected := []string{
		"[Section1]",
		"Prop1=dog",
		"Prop2=frog",
		"PropList=tuna",
		"PropList=flounder",
		"PropList=shark",
		"Prop3=bear",
		"",
		"[Section2]",
		"New=unicorn",
	}
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestWriteFileReplacesContents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.ini")
	if err := os.WriteFile(path, []byte("old contents that are longer"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := writeFile(path, "new"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf(`Output not as expected. Actual: %q`, content)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf(`Expected permissions to be kept. Actual: %v`, info.Mode().Perm())
	}

}

func TestWriteFileCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.ini")

	if err := writeFile(path, "new"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf(`Output not as expected. Actual: %q`, content)
	}

}

func TestWriteFileMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "out.ini")

	if err := writeFile(path, "new"); err == nil {
		t.Error("expected an error for a missing directory")
	}
}

func TestWriteFileFollowsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.ini")
	link := filepath.Join(dir, "link.ini")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	if err := writeFile(link, "new"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected the symlink to be kept")
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf(`Output not as expected. Actual: %q`, content)
	}
}

func TestUnescapeNewline(t *testing.T) {
	cases := map[string]string{
		`\r\n`: "\r\n",
		`\n`:   "\n",
		`\r`:   "\r",
		"\r\n": "\r\n",
		"\n":   "\n",
	}

	for input, expected := range cases {
		if actual := unescapeNewline(input); actual != expected {
			t.Errorf(`unescapeNewline(%q) = %q, expected %q`, input, actual, expected)
		}
	}
}

func TestCheckArgs(t *testing.T) {
	if err := checkArgs("in.ini", "", ";"); err != nil {
		t.Errorf(`Expected input alone to be valid. Actual: %v`, err)
	}
	if err := checkArgs("", "[S];A=1", ";"); err != nil {
		t.Errorf(`Expected modify alone to be valid. Actual: %v`, err)
	}
	if err := checkArgs("", "", ";"); err == nil {
		t.Error("expected an error when -input and -modify are both missing")
	}
	if err := checkArgs("in.ini", "[S];A=1", ""); err == nil {
		t.Error("expected an error for an empty -delimit")
	}
}

func TestBuildVersion(t *testing.T) {
	original := version
	defer func() { version = original }()

	version = "v1.2.3"
	if actual := buildVersion(); actual != "v1.2.3" {
		t.Errorf(`Expected the release version. Actual: %q`, actual)
	}

	version = ""
	if actual := buildVersion(); actual == "" {
		t.Error("expected a fallback version")
	}
}

func TestWriteFileKeepsHardLinks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.ini")
	link := filepath.Join(dir, "link.ini")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(path, link); err != nil {
		t.Skipf("cannot create hard link: %v", err)
	}

	if err := writeFile(path, "new"); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf(`Expected the hard link to see the new contents. Actual: %q`, content)
	}
}

func runModini(args ...string) (int, string, string) {
	var stdout, stderr strings.Builder
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunWritesToStdout(t *testing.T) {
	code, stdout, stderr := runModini("-modify", "[S];A=1")

	if code != 0 {
		t.Errorf(`Expected exit code 0. Actual: %v, stderr: %v`, code, stderr)
	}
	if stdout != "[S]\nA=1\n" {
		t.Errorf(`Output not as expected. Actual: %q`, stdout)
	}
}

func TestRunWritesToFile(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ini")
	output := filepath.Join(dir, "output.ini")
	if err := os.WriteFile(input, []byte("[S]\nA=1\njunk\n"), 0644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runModini("-input", input, "-output", output, "-modify", "[S];B=2", "-newline", `\r\n`)

	if code != 0 {
		t.Errorf(`Expected exit code 0. Actual: %v, stderr: %v`, code, stderr)
	}
	if stdout != "" {
		t.Errorf(`Expected nothing on stdout. Actual: %q`, stdout)
	}
	if !strings.Contains(stderr, "skipping line #3: junk") {
		t.Errorf(`Expected a warning for the junk line. Actual: %q`, stderr)
	}
	content, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "[S]\r\nA=1\r\nB=2\r\n" {
		t.Errorf(`Output not as expected. Actual: %q`, content)
	}
}

func TestRunVersion(t *testing.T) {
	code, stdout, _ := runModini("-version")

	if code != 0 || strings.TrimSpace(stdout) == "" {
		t.Errorf(`Expected a version and exit code 0. Actual: %v, %q`, code, stdout)
	}
}

func TestRunHelp(t *testing.T) {
	code, _, stderr := runModini("-h")

	if code != 0 || !strings.Contains(stderr, "-modify") {
		t.Errorf(`Expected usage and exit code 0. Actual: %v, %q`, code, stderr)
	}
}

func TestRunInvalidArguments(t *testing.T) {
	cases := [][]string{
		{},
		{"-unknown"},
		{"-modify", "[S];A=1", "-delimit", ""},
	}

	for _, args := range cases {
		if code, _, _ := runModini(args...); code != 1 {
			t.Errorf(`%v: expected exit code 1. Actual: %v`, args, code)
		}
	}
}

func TestRunInvalidModifyWritesNothing(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.ini")

	code, stdout, stderr := runModini("-output", output, "-modify", "[S];A=1;oops")

	if code != 1 {
		t.Errorf(`Expected exit code 1. Actual: %v`, code)
	}
	if stdout != "" {
		t.Errorf(`Expected nothing on stdout. Actual: %q`, stdout)
	}
	if !strings.Contains(stderr, `invalid modification: "oops"`) {
		t.Errorf(`Expected an error naming the bad piece. Actual: %q`, stderr)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Errorf(`Expected no output file. Actual error: %v`, err)
	}
}

func TestRunUnreadableInputWritesNothing(t *testing.T) {
	dir := t.TempDir()
	utf16 := filepath.Join(dir, "utf16.ini")
	if err := os.WriteFile(utf16, []byte{0xFF, 0xFE, '[', 0, 'S', 0, ']', 0}, 0644); err != nil {
		t.Fatal(err)
	}

	for _, input := range []string{filepath.Join(dir, "missing.ini"), utf16} {
		output := filepath.Join(dir, "output.ini")

		code, stdout, _ := runModini("-input", input, "-output", output)

		if code != 2 {
			t.Errorf(`%v: expected exit code 2. Actual: %v`, input, code)
		}
		if stdout != "" {
			t.Errorf(`%v: expected nothing on stdout. Actual: %q`, input, stdout)
		}
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Errorf(`%v: expected no output file. Actual error: %v`, input, err)
		}
	}
}

func TestRunWriteFailure(t *testing.T) {
	output := filepath.Join(t.TempDir(), "missing", "output.ini")
	if code, _, _ := runModini("-modify", "[S];A=1", "-output", output); code != 3 {
		t.Errorf(`Expected exit code 3 for an unwritable file. Actual: %v`, code)
	}

	var stderr strings.Builder
	if code := run([]string{"-modify", "[S];A=1"}, failingWriter{}, &stderr); code != 3 {
		t.Errorf(`Expected exit code 3 for a failed stdout write. Actual: %v`, code)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("disk full")
}
