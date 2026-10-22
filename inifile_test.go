package main

import (
	"slices"
	"testing"
)

func TestMergeAndModify(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		modify      []string
		want        []string
		wantSkipped []skippedLine // lines skipped from input
		wantInvalid []skippedLine // pieces rejected from modify
	}{
		{
			name: "render order",
			input: []string{
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
			},
			want: []string{
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
			},
		},
		{
			name: "skips garbage",
			input: []string{
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
			},
			want: []string{
				"[My.Section]",
				"Prop=value",
				"Prop=this is a test",
				"",
				"[Other.Section]",
				"Thing2=b",
				"Thing1=a",
			},
			wantSkipped: []skippedLine{
				{1, "garbage1"},
				{4, "garbage2"},
				{6, "garbage3"},
				{10, "garbage4"},
				{11, "=a"},
				{12, "[begin"},
				{13, "end]"},
			},
		},
		{
			name:   "invalid modify is reported",
			modify: []string{"[S]", "A=1", "Bvalue", "[Broken", "", "C"},
			want:   []string{"[S]", "A=1"},
			wantInvalid: []skippedLine{
				{3, "Bvalue"},
				{4, "[Broken"},
				{6, "C"},
			},
		},
		{
			name: "merges correctly",
			input: []string{
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
			},
			modify: []string{
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
			},
			want: []string{
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
			},
		},
		{
			name:  "no duplicates from input",
			input: []string{"[My.Section]", "Prop=value", "Prop=value"},
			want:  []string{"[My.Section]", "Prop=value"},
		},
		{
			name:   "no duplicates from modify",
			input:  []string{"[My.Section]", "Prop=value", "Prop=b"},
			modify: []string{"[My.Section]", "Prop+=value"},
			want:   []string{"[My.Section]", "Prop=value", "Prop=b"},
		},
		{
			name:   "global properties",
			input:  []string{"root=true", "[My.Section]", "Prop=value"},
			modify: []string{"extra=yes", "[My.Section]", "Prop=changed"},
			want:   []string{"root=true", "extra=yes", "", "[My.Section]", "Prop=changed"},
		},
		{
			name:   "modify only global",
			modify: []string{"Prop=x"},
			want:   []string{"Prop=x"},
		},
		{
			name:  "empty file",
			input: []string{},
			want:  []string{},
		},
		{
			name:   "set replaces all values",
			input:  []string{"[S]", "List=a", "List=b", "List=c"},
			modify: []string{"[S]", "List=z", "List-=missing"},
			want:   []string{"[S]", "List=z"},
		},
		{
			name:   "remove last value drops property",
			input:  []string{"[S]", "A=1", "B=2"},
			modify: []string{"[S]", "A-=1"},
			want:   []string{"[S]", "B=2"},
		},
		{
			name:  "operators are literal in input file",
			input: []string{"[S]", "A=1", "A+=2", "A-=1", "CFLAGS += -O2"},
			want:  []string{"[S]", "A=1", "A+=2", "A-=1", "CFLAGS +=-O2"},
		},
		{
			name:   "modify can target a literal key ending in +",
			input:  []string{"[S]", "A=1", "A+=2"},
			modify: []string{"[S]", "A+ = 3"},
			want:   []string{"[S]", "A=1", "A+=3"},
		},
		{
			name: "values keep special characters",
			input: []string{
				"[S]",
				"Format=100% %s %d",
				"Url=http://example.com/?a=b&c=d",
				"Trailing=value   ",
			},
			want: []string{
				"[S]",
				"Format=100% %s %d",
				"Url=http://example.com/?a=b&c=d",
				"Trailing=value",
			},
		},
		{
			name:  "windows line endings",
			input: []string{"[S]\r", "A=1\r", "\r"},
			want:  []string{"[S]", "A=1"},
		},
		{
			name: "section header with indent and comment",
			input: []string{
				"[A]",
				"x=1",
				"[B] ; database settings",
				"y=2",
				"  [C]",
				"z=3",
				"[D]   # other",
				"w=4",
			},
			want: []string{
				"[A]", "x=1", "",
				"[B]", "y=2", "",
				"[C]", "z=3", "",
				"[D]", "w=4",
			},
		},
		{
			name: "whitespace around operators",
			input: []string{
				"[S]",
				"port = 80",
				"list = a",
				"list= b",
				"name =value with spaces",
			},
			modify: []string{"[S]", "port = 8080", "list -= a", "list += c"},
			want:   []string{"[S]", "port=8080", "list=b", "list=c", "name=value with spaces"},
		},
		{
			name: "indented and punctuated keys",
			input: []string{
				"[S]",
				"    name=foo",
				"\tTab=bar",
				".NET=4",
				"@version=2",
				"$var=x",
				"; comment=not a key",
				"# comment=not a key",
			},
			want: []string{"[S]", "name=foo", "Tab=bar", ".NET=4", "@version=2", "$var=x"},
			wantSkipped: []skippedLine{
				{7, "; comment=not a key"},
				{8, "# comment=not a key"},
			},
		},
		{
			name:   "remove does not create",
			input:  []string{"[S]", "A=1"},
			modify: []string{"[Legacy]", "old-=x", "[S]", "B-=2", "A-=9", "global-=y"},
			want:   []string{"[S]", "A=1"},
		},
		{
			name:   "set after remove creates property",
			input:  []string{"[S]", "A=1"},
			modify: []string{"[S]", "B-=2", "B=2"},
			want:   []string{"[S]", "A=1", "B=2"},
		},
		{
			name:  "empty section from input is kept",
			input: []string{"[Empty]", "[S]", "A=1"},
			want:  []string{"[Empty]", "", "[S]", "A=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := newFile()

			skipped := file.mergeFile(tt.input)
			if !slices.Equal(skipped, tt.wantSkipped) {
				t.Errorf("mergeFile() skipped %v, want %v", skipped, tt.wantSkipped)
			}

			invalid := file.applyModify(tt.modify)
			if !slices.Equal(invalid, tt.wantInvalid) {
				t.Errorf("applyModify() rejected %v, want %v", invalid, tt.wantInvalid)
			}

			if got := file.render(); !slices.Equal(got, tt.want) {
				t.Errorf("render() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
