package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadModifySplitsOnDelimiterAndNewlines(t *testing.T) {
	channel := make(chan string)
	go readModify(channel, "[S];A=1\nB=2|C=3", ";")

	expected := []string{"[S]", "A=1", "B=2|C=3"}
	actual := channelToSlice(channel)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestReadModifyCustomDelimiter(t *testing.T) {
	channel := make(chan string)
	go readModify(channel, "[S]|A=1;2|B=3", "|")

	expected := []string{"[S]", "A=1;2", "B=3"}
	actual := channelToSlice(channel)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestReadFileHandlesCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.ini")
	if err := os.WriteFile(path, []byte("[S]\r\nA=1\r\n"), 0644); err != nil {
		t.Fatal(err)
	}

	channel := make(chan string)
	go readFile(channel, path)

	file := newFile()
	file.merge(channel, true)
	actual := file.render()

	expected := []string{"[S]", "A=1"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
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

	fileChannel := make(chan string)
	go readFile(fileChannel, path)
	file.merge(fileChannel, true)

	modifyChannel := make(chan string)
	go readModify(modifyChannel, "[Section1];Prop2=frog;Prop3=bear;PropList+=shark;PropList-=trout;[Section2];New=unicorn", ";")
	file.merge(modifyChannel, false)

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

func channelToSlice(channel <-chan string) []string {
	result := []string{}
	for line := range channel {
		result = append(result, line)
	}
	return result
}
