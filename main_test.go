package main

import (
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
	file.merge(lines, true)
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
	file.merge(lines, true)
	file.merge(readModify("[Main];key=2", ";"), false)
	actual := file.render()

	expected := []string{"[Main]", "key=2"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
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
	file.merge(lines, true)

	file.merge(readModify("[Section1];Prop2=frog;Prop3=bear;PropList+=shark;PropList-=trout;[Section2];New=unicorn", ";"), false)

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

func TestGetVersion(t *testing.T) {
	original := version
	defer func() { version = original }()

	version = "v1.2.3"
	if actual := getVersion(); actual != "v1.2.3" {
		t.Errorf(`Expected the release version. Actual: %q`, actual)
	}

	version = ""
	if actual := getVersion(); actual == "" {
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
