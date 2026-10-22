package main

import (
	"slices"
)

type iniProperty struct {
	name   string
	values []string
}

func newProperty(name string) *iniProperty {
	return &iniProperty{
		name:   name,
		values: []string{},
	}
}

func (p *iniProperty) setValue(value string) {
	p.values = []string{value}
}

func (p *iniProperty) addValue(value string) {
	if !slices.Contains(p.values, value) {
		p.values = append(p.values, value)
	}
}

func (p *iniProperty) removeValue(value string) {
	p.values = slices.DeleteFunc(p.values, func(v string) bool {
		return v == value
	})
}
