package main

import (
	"os"
	"path/filepath"
	"reflect"
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
