# modini
Modify ini files from the command line

[![Build Status](https://travis-ci.org/reflectivecode/modini.svg?branch=master)](https://travis-ci.org/reflectivecode/modini)

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

## Build
```
go build
go test
```

## Options
```
usage: modini --input=INPUT [<flags>]

Flags:
  -delimit string
        (optional) Split the modify string on this value. (default ";")
  -input string
        (required) Path to input ini file.
  -modify string
        (optional) Modifications to make. ex: [section1];prop1=value1;prop2=value2;[section2];prop1=value3.
  -output string
        (optional) Path to output ini file.
  -version
        output the version
```
