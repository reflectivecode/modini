package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestReadModifySplitsOnDelimiterAndNewlines(t *testing.T) {
	got := readModify("[S];A=1\nB=2|C=3", ";")
	want := []string{"[S]", "A=1", "B=2|C=3"}

	if !slices.Equal(got, want) {
		t.Errorf("readModify() = %q, want %q", got, want)
	}
}

func TestReadModifyCustomDelimiter(t *testing.T) {
	got := readModify("[S]|A=1;2|B=3", "|")
	want := []string{"[S]", "A=1;2", "B=3"}

	if !slices.Equal(got, want) {
		t.Errorf("readModify() = %q, want %q", got, want)
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
	got := file.render()
	want := []string{"[S]", "A=1"}

	if !slices.Equal(got, want) {
		t.Errorf("render() = %q, want %q", got, want)
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
	got := file.render()
	want := []string{"[Main]", "key=2"}

	if !slices.Equal(got, want) {
		t.Errorf("render() = %q, want %q", got, want)
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
			t.Errorf("%v: readFile() error = nil, want an error", name)
		}
	}
}

func TestReadFileMissing(t *testing.T) {
	if _, err := readFile(filepath.Join(t.TempDir(), "missing.ini")); err == nil {
		t.Error("readFile() error = nil, want an error")
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

	lines, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}

	file := newFile()
	file.mergeFile(lines)
	file.applyModify(readModify("[Section1];Prop2=frog;Prop3=bear;PropList+=shark;PropList-=trout;[Section2];New=unicorn", ";"))
	got := file.render()
	want := []string{
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

	if !slices.Equal(got, want) {
		t.Errorf("render() =\n%q\nwant\n%q", got, want)
	}
}

func TestWriteFileReplacesContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.ini")
	if err := os.WriteFile(path, []byte("old contents that are longer"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := writeFile(path, "new"); err != nil {
		t.Fatal(err)
	}

	assertFileContents(t, path, "new")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0600); runtime.GOOS != "windows" && got != want {
		t.Errorf("permissions = %v, want %v", got, want)
	}
}

func TestWriteFileCreatesNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.ini")

	if err := writeFile(path, "new"); err != nil {
		t.Fatal(err)
	}

	assertFileContents(t, path, "new")
}

func TestWriteFileMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "out.ini")

	if err := writeFile(path, "new"); err == nil {
		t.Error("writeFile() error = nil, want an error")
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
		t.Errorf("link mode = %v, want a symlink", info.Mode())
	}

	assertFileContents(t, target, "new")
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

	assertFileContents(t, link, "new")
}

func TestUnescapeNewline(t *testing.T) {
	tests := map[string]string{
		`\r\n`: "\r\n",
		`\n`:   "\n",
		`\r`:   "\r",
		"\r\n": "\r\n",
		"\n":   "\n",
	}

	for input, want := range tests {
		if got := unescapeNewline(input); got != want {
			t.Errorf("unescapeNewline(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		inFile, modify, delimiter string
		wantErr                   bool
	}{
		{"in.ini", "", ";", false},
		{"", "[S];A=1", ";", false},
		{"", "", ";", true},
		{"in.ini", "[S];A=1", "", true},
	}

	for _, tt := range tests {
		err := checkArgs(tt.inFile, tt.modify, tt.delimiter)
		if (err != nil) != tt.wantErr {
			t.Errorf("checkArgs(%q, %q, %q) error = %v, want error %v", tt.inFile, tt.modify, tt.delimiter, err, tt.wantErr)
		}
	}
}

func TestBuildVersion(t *testing.T) {
	original := version
	defer func() { version = original }()

	version = "v1.2.3"
	if got, want := buildVersion(), "v1.2.3"; got != want {
		t.Errorf("buildVersion() = %q, want %q", got, want)
	}

	version = ""
	if got := buildVersion(); got == "" {
		t.Error(`buildVersion() = "", want a fallback version`)
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
		t.Errorf("exit code = %v, want 0 (stderr: %q)", code, stderr)
	}
	if want := "[S]\nA=1\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
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
		t.Errorf("exit code = %v, want 0 (stderr: %q)", code, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if want := "skipping line #3: junk"; !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}
	assertFileContents(t, output, "[S]\r\nA=1\r\nB=2\r\n")
}

func TestRunVersion(t *testing.T) {
	code, stdout, _ := runModini("-version")

	if code != 0 {
		t.Errorf("exit code = %v, want 0", code)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Errorf("stdout = %q, want a version", stdout)
	}
}

func TestRunHelp(t *testing.T) {
	code, _, stderr := runModini("-h")

	if code != 0 {
		t.Errorf("exit code = %v, want 0", code)
	}
	if want := "-modify"; !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want usage containing %q", stderr, want)
	}
}

func TestRunInvalidArguments(t *testing.T) {
	tests := [][]string{
		{},
		{"-unknown"},
		{"-modify", "[S];A=1", "-delimit", ""},
	}

	for _, args := range tests {
		if code, _, _ := runModini(args...); code != 1 {
			t.Errorf("run(%q) exit code = %v, want 1", args, code)
		}
	}
}

func TestRunInvalidModifyWritesNothing(t *testing.T) {
	output := filepath.Join(t.TempDir(), "output.ini")

	code, stdout, stderr := runModini("-output", output, "-modify", "[S];A=1;oops")

	if code != 1 {
		t.Errorf("exit code = %v, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if want := `invalid modification: "oops"`; !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}
	assertNoFile(t, output)
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
			t.Errorf("%v: exit code = %v, want 2", input, code)
		}
		if stdout != "" {
			t.Errorf("%v: stdout = %q, want empty", input, stdout)
		}
		assertNoFile(t, output)
	}
}

func TestRunWriteFailure(t *testing.T) {
	output := filepath.Join(t.TempDir(), "missing", "output.ini")
	if code, _, _ := runModini("-modify", "[S];A=1", "-output", output); code != 3 {
		t.Errorf("unwritable file: exit code = %v, want 3", code)
	}

	var stderr strings.Builder
	if code := run([]string{"-modify", "[S];A=1"}, failingWriter{}, &stderr); code != 3 {
		t.Errorf("failed stdout write: exit code = %v, want 3", code)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("disk full")
}

func assertFileContents(t *testing.T, path string, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%v contents = %q, want %q", filepath.Base(path), got, want)
	}
}

func assertNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%v: stat error = %v, want file not to exist", filepath.Base(path), err)
	}
}
