# modini
Modify ini files from the command line

[![CI](https://github.com/reflectivecode/modini/actions/workflows/ci.yml/badge.svg)](https://github.com/reflectivecode/modini/actions/workflows/ci.yml)

## Example
```
modini -input input.ini -output output.ini -modify "[Section1];Prop2=frog;Prop3=bear;PropList+=shark;PropList-=trout;[Section2];New=unicorn"

=== input.ini ===
[Section1]
Prop1=dog
Prop2=cat
PropList=tuna
PropList=trout
PropList=flounder

=== output.ini ===
[Section1]
Prop1=dog
Prop2=frog
PropList=tuna
PropList=flounder
PropList=shark
Prop3=bear

[Section2]
New=unicorn
```

## Modify syntax
The modify string is split on the delimiter (`;` by default) and on newlines. Each piece is one of:

| Syntax | Effect |
| --- | --- |
| `[Section]` | Following modifications apply to this section (created if missing) |
| `Prop=value` | Replace all values of `Prop` with `value` |
| `Prop+=value` | Add `value` to `Prop`, unless it is already present |
| `Prop-=value` | Remove `value` from `Prop` |

Properties that come before any `[Section]` apply to the top of the file, outside any section.

If your values contain `;`, choose a different delimiter with `-delimit`.

## Notes
- Sections and properties keep their original order, and new ones are added at the end.
- If a section appears more than once in the input, its properties are merged into one section.
- Duplicate `Prop=value` lines are collapsed into one.
- Whitespace around `=`, `+=` and `-=` is removed, so `Prop = value` is written as `Prop=value`.
- Comments and any lines that can't be parsed are dropped. A warning for each is printed to stderr.
- `-input` and `-output` can be the same file. The output is written to a temporary file and then renamed into place, so a failed write leaves the original untouched.
- If `-output` is omitted, the result is written to stdout.

## Options
At least one of `-input` or `-modify` is required.
```
  -delimit string
        Split the modify string on this value. (default ";")
  -input string
        Path to input ini file.
  -modify string
        Modifications to make. ex: [section1];prop1=value1;prop2=value2;[section2];prop1=value3.
  -newline string
        Output uses this line ending. (default "\n")
  -output string
        Path to output ini file.
  -version
        output the version
```

## Download
Prebuilt binaries for Linux, macOS and Windows are attached to each [GitHub release](https://github.com/reflectivecode/modini/releases), along with a `SHA256SUMS` file.

## Build
Requires Go 1.26 or newer.
```
go build
go test ./...
```
