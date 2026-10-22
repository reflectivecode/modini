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

	assertOnlyFiles(t, dir, "out.ini")
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

	assertOnlyFiles(t, dir, "out.ini")
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

func assertOnlyFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	actual := []string{}
	for _, entry := range entries {
		actual = append(actual, entry.Name())
	}
	if !reflect.DeepEqual(actual, names) {
		t.Errorf(`Unexpected files in directory: %v`, actual)
	}
}
