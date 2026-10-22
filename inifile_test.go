package main

import (
	"reflect"
	"testing"
)

func TestRenderOrder(t *testing.T) {
	input := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=this is a test",
		"",
		"[Other.Section]",
		"Thing2=b",
		"Thing1=a",
		"Strange[0]=one",
		"",
		"[My.Section]",
		"Other=yes",
		"a=b",
	}

	expected := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=this is a test",
		"Other=yes",
		"a=b",
		"",
		"[Other.Section]",
		"Thing2=b",
		"Thing1=a",
		"Strange[0]=one",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestSkipsGarbage(t *testing.T) {
	input := []string{
		"garbage1",
		"[My.Section]",
		"Prop=value",
		"garbage2",
		"Prop=this is a test",
		"garbage3",
		"[Other.Section]",
		"Thing2=b",
		"Thing1=a",
		"garbage4",
		"=a",
		"[begin",
		"end]",
	}

	expected := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=this is a test",
		"",
		"[Other.Section]",
		"Thing2=b",
		"Thing1=a",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestMergesCorrectly(t *testing.T) {
	input1 := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=this is a test",
		"",
		"[Other.Section]",
		"Thing2=b",
		"Thing1=a",
		"",
		"[My.Section]",
		"Other=yes",
		"Other=no",
	}

	input2 := []string{
		"[New.Section]",
		"Window=yes",
		"[My.Section]",
		"Prop+=another",
		"Prop-=this is a test",
		"Other=maybe",
		"[Other.Section]",
		"Thing2=c",
		"Thing3=d",
		"",
	}

	expected := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=another",
		"Other=maybe",
		"",
		"[Other.Section]",
		"Thing2=c",
		"Thing1=a",
		"Thing3=d",
		"",
		"[New.Section]",
		"Window=yes",
	}

	file := newFile()
	file.merge(input1, true)
	file.merge(input2, false)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestNoDuplicatesFromInput(t *testing.T) {
	input := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=value",
	}

	expected := []string{
		"[My.Section]",
		"Prop=value",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestNoDuplicatesFromModify(t *testing.T) {
	input1 := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=b",
	}

	input2 := []string{
		"[My.Section]",
		"Prop+=value",
	}

	expected := []string{
		"[My.Section]",
		"Prop=value",
		"Prop=b",
	}

	file := newFile()
	file.merge(input1, true)
	file.merge(input2, false)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestGlobalProperties(t *testing.T) {
	input1 := []string{
		"root=true",
		"[My.Section]",
		"Prop=value",
	}

	input2 := []string{
		"extra=yes",
		"[My.Section]",
		"Prop=changed",
	}

	expected := []string{
		"root=true",
		"extra=yes",
		"",
		"[My.Section]",
		"Prop=changed",
	}

	file := newFile()
	file.merge(input1, true)
	file.merge(input2, false)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestModifyOnlyGlobal(t *testing.T) {
	file := newFile()
	file.merge([]string{"Prop=x"}, false)
	actual := file.render()

	expected := []string{"Prop=x"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestEmptyFile(t *testing.T) {
	file := newFile()
	file.merge([]string{}, true)
	actual := file.render()

	if len(actual) != 0 {
		t.Errorf(`Expected no output. Actual: %v`, actual)
	}
}

func TestSetReplacesAllValues(t *testing.T) {
	input1 := []string{
		"[S]",
		"List=a",
		"List=b",
		"List=c",
	}

	input2 := []string{
		"[S]",
		"List=z",
		"List-=missing",
	}

	expected := []string{
		"[S]",
		"List=z",
	}

	file := newFile()
	file.merge(input1, true)
	file.merge(input2, false)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestRemoveLastValueDropsProperty(t *testing.T) {
	input1 := []string{
		"[S]",
		"A=1",
		"B=2",
	}

	input2 := []string{
		"[S]",
		"A-=1",
	}

	expected := []string{
		"[S]",
		"B=2",
	}

	file := newFile()
	file.merge(input1, true)
	file.merge(input2, false)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestOperatorsIgnoredInInputFile(t *testing.T) {
	input := []string{
		"[S]",
		"A=1",
		"A+=2",
		"A-=1",
	}

	expected := []string{
		"[S]",
		"A=1",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestValuesKeepSpecialCharacters(t *testing.T) {
	input := []string{
		"[S]",
		"Format=100% %s %d",
		"Url=http://example.com/?a=b&c=d",
		"Trailing=value   ",
	}

	expected := []string{
		"[S]",
		"Format=100% %s %d",
		"Url=http://example.com/?a=b&c=d",
		"Trailing=value",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestWindowsLineEndings(t *testing.T) {
	input := []string{
		"[S]\r",
		"A=1\r",
		"\r",
	}

	expected := []string{
		"[S]",
		"A=1",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestSectionHeaderWithIndentAndComment(t *testing.T) {
	input := []string{
		"[A]",
		"x=1",
		"[B] ; database settings",
		"y=2",
		"  [C]",
		"z=3",
		"[D]   # other",
		"w=4",
	}

	expected := []string{
		"[A]",
		"x=1",
		"",
		"[B]",
		"y=2",
		"",
		"[C]",
		"z=3",
		"",
		"[D]",
		"w=4",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestWhitespaceAroundOperators(t *testing.T) {
	input1 := []string{
		"[S]",
		"port = 80",
		"list = a",
		"list= b",
		"name =value with spaces",
	}

	input2 := []string{
		"[S]",
		"port = 8080",
		"list -= a",
		"list += c",
	}

	expected := []string{
		"[S]",
		"port=8080",
		"list=b",
		"list=c",
		"name=value with spaces",
	}

	file := newFile()
	file.merge(input1, true)
	file.merge(input2, false)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}

func TestIndentedAndPunctuatedKeys(t *testing.T) {
	input := []string{
		"[S]",
		"    name=foo",
		"\tTab=bar",
		".NET=4",
		"@version=2",
		"$var=x",
		"; comment=not a key",
		"# comment=not a key",
	}

	expected := []string{
		"[S]",
		"name=foo",
		"Tab=bar",
		".NET=4",
		"@version=2",
		"$var=x",
	}

	file := newFile()
	file.merge(input, true)
	actual := file.render()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Output not as expected. Actual: %v`, actual)
	}
}
