package main

import (
	"regexp"
)

type iniFile struct {
	sectionOrder []string
	sectionMap   map[string]*iniSection
}

func newFile() *iniFile {
	return &iniFile{
		sectionOrder: []string{},
		sectionMap:   make(map[string]*iniSection),
	}
}

// ensureSection returns the named section, creating it if needed.
func (f *iniFile) ensureSection(name string) *iniSection {
	if s, ok := f.sectionMap[name]; ok {
		return s
	}

	s := newSection(name)
	f.sectionOrder = append(f.sectionOrder, name)
	f.sectionMap[name] = s
	return s
}

// findProperty returns nil if the section or property doesn't exist.
func (f *iniFile) findProperty(sectionName string, propertyName string) *iniProperty {
	s, ok := f.sectionMap[sectionName]
	if !ok {
		return nil
	}
	return s.propertyMap[propertyName]
}

func (f *iniFile) allSections() []*iniSection {
	sections := []*iniSection{}
	for _, name := range f.sectionOrder {
		sections = append(sections, f.sectionMap[name])
	}
	return sections
}

var emptyRegex = regexp.MustCompile(`^\s*$`)
var sectionRegex = regexp.MustCompile(`^\s*\[([^\[\]]+)\]\s*([;#].*)?$`)

// Operators only have meaning in -modify. In an input file, a line like
// "a+=b" is the key "a+" with the value "b".
var filePropertyRegex = regexp.MustCompile(`^\s*([^\s;#\[=][^=]*?)\s*(=)\s*(.*?)\s*$`)
var modifyPropertyRegex = regexp.MustCompile(`^\s*([^\s;#\[=][^=]*?)\s*(=|\+=|\-=)\s*(.*?)\s*$`)

// globalSection holds properties that appear before any [section] header.
// The section regex requires a non-empty name, so this can't collide.
const globalSection = ""

type skippedLine struct {
	number int
	text   string
}

// parsedLine is either a section header or a property in section.
type parsedLine struct {
	section  string
	isHeader bool
	key      string
	operator string
	value    string
}

// parseLines classifies each line and returns the lines it couldn't parse
// separately. Blank lines are dropped.
func parseLines(lines []string, propertyRegex *regexp.Regexp) ([]parsedLine, []skippedLine) {
	parsed := []parsedLine{}
	skipped := []skippedLine{}
	section := globalSection

	for i, line := range lines {
		if emptyRegex.MatchString(line) {
			continue
		}

		if match := sectionRegex.FindStringSubmatch(line); match != nil {
			section = match[1]
			parsed = append(parsed, parsedLine{section: section, isHeader: true})
			continue
		}

		if match := propertyRegex.FindStringSubmatch(line); match != nil {
			parsed = append(parsed, parsedLine{section: section, key: match[1], operator: match[2], value: match[3]})
			continue
		}

		skipped = append(skipped, skippedLine{number: i + 1, text: line})
	}
	return parsed, skipped
}

// mergeFile adds the contents of an input file and returns the lines it
// couldn't parse.
func (f *iniFile) mergeFile(lines []string) []skippedLine {
	parsed, skipped := parseLines(lines, filePropertyRegex)
	for _, line := range parsed {
		s := f.ensureSection(line.section)
		if !line.isHeader {
			s.ensureProperty(line.key).addValue(line.value)
		}
	}
	return skipped
}

// applyModify applies modifications and returns the ones it couldn't
// parse. Sections and properties are only created when a value is set or
// added, so a removal never adds anything to the output.
func (f *iniFile) applyModify(lines []string) []skippedLine {
	parsed, skipped := parseLines(lines, modifyPropertyRegex)
	for _, line := range parsed {
		switch line.operator {
		case "=":
			f.ensureSection(line.section).ensureProperty(line.key).setValue(line.value)
		case "+=":
			f.ensureSection(line.section).ensureProperty(line.key).addValue(line.value)
		case "-=":
			if p := f.findProperty(line.section, line.key); p != nil {
				p.removeValue(line.value)
			}
		}
	}
	return skipped
}

func (f *iniFile) render() []string {
	result := []string{}

	if global, ok := f.sectionMap[globalSection]; ok {
		result = append(result, renderProperties(global)...)
	}

	for _, s := range f.allSections() {
		if s.name == globalSection {
			continue
		}

		if len(result) > 0 {
			result = append(result, "")
		}

		result = append(result, "["+s.name+"]")
		result = append(result, renderProperties(s)...)
	}
	return result
}

func renderProperties(s *iniSection) []string {
	result := []string{}
	for _, p := range s.allProperties() {
		for _, value := range p.values {
			result = append(result, p.name+"="+value)
		}
	}
	return result
}
