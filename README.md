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
| `[Section]` | Following modifications apply to this section (created when a value is set or added) |
| `Prop=value` | Replace all values of `Prop` with `value` |
| `Prop+=value` | Add `value` to `Prop`, unless it is already present |
| `Prop-=value` | Remove `value` from `Prop` |

Properties that come before any `[Section]` apply to the top of the file, outside any section.

Empty pieces, such as one left by a trailing `;`, are ignored. If your values contain `;`, choose a different delimiter with `-delimit`.

## Notes
- Sections and properties keep their original order, and new ones are added at the end.
- If a section appears more than once in the input, its properties are merged into one section.
- Duplicate `Prop=value` lines are collapsed into one.
- Whitespace around `=` is removed, so `Prop = value` is written as `Prop=value`. In the modify string, whitespace around `+=` and `-=` is removed too.
- In the input file, `+=` and `-=` have no special meaning. `CFLAGS+=-O2` is the key `CFLAGS+` with the value `-O2`.
- Comments and any lines in the input file that can't be parsed are dropped. A warning for each is printed to stderr.
- A piece of the modify string that can't be parsed is an error. Nothing is written and the exit code is 1.
- `-input` and `-output` can be the same file.
- If `-output` is omitted, the result is written to stdout.

## Limitations
- Comments are not preserved, and the output is reformatted: one blank line between sections and no whitespace around `=`.
- Text after a value is part of the value. `x=1 ; note` has the value `1 ; note`. Only section headers have a trailing comment removed.
- Section and key names are case-sensitive, and spaces inside brackets count. `[a]`, `[A]` and `[ a ]` are three different sections.
- Quotes are not interpreted. `x="a b"` has the value `"a b"`, including the quotes.
- In `-modify`, `Prop+=value` and `Prop-=value` are always operators. To set a key whose name ends in `+` or `-`, put a space before the `=`, as in `Prop+ =value`.
- The output file is overwritten in place. If the write fails partway, for example because the disk is full, the file can be left empty or incomplete. When `-input` and `-output` are the same file, the original is lost in that case. If that matters, write to a different file and move it over the original once modini succeeds.

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
        Output uses this line ending. Escapes \r and \n are interpreted, ex: \r\n. (default "\n")
  -output string
        Path to output ini file.
  -version
        output the version
```

## Exit codes
| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Invalid arguments, or a piece of the modify string that can't be parsed. Nothing is written |
| 2 | The input file can't be read. Nothing is written |
| 3 | The output can't be written |

## Download
Prebuilt binaries are attached to each [GitHub release](https://github.com/reflectivecode/modini/releases), along with a `SHA256SUMS` file. Releases tagged with a `-`, such as `v1.0.0-rc1`, are marked as pre-releases.

| File | Platform |
| --- | --- |
| `modini-amd64` | Linux x86-64 |
| `modini-arm64` | Linux ARM64 |
| `modini-arm` | Linux ARM (32-bit) |
| `modini-386` | Linux x86 (32-bit) |
| `modini-darwin` | macOS Intel |
| `modini-darwin-arm64` | macOS Apple Silicon |
| `modini-amd64.exe` | Windows x86-64 |
| `modini-arm64.exe` | Windows ARM64 |
| `modini-386.exe` | Windows x86 (32-bit) |

To always get the newest stable release, use `https://github.com/reflectivecode/modini/releases/latest/download/<file>`.

## Build
Requires Go 1.26 or newer.
```
go build
go test ./...
```
