package main

type iniSection struct {
	name          string
	propertyOrder []string
	propertyMap   map[string]*iniProperty
}

func newSection(name string) *iniSection {
	return &iniSection{
		name:          name,
		propertyOrder: []string{},
		propertyMap:   make(map[string]*iniProperty),
	}
}

// ensureProperty returns the named property, creating it if needed.
func (s *iniSection) ensureProperty(name string) *iniProperty {
	if p, ok := s.propertyMap[name]; ok {
		return p
	}

	p := newProperty(name)
	s.propertyOrder = append(s.propertyOrder, name)
	s.propertyMap[name] = p
	return p
}

func (s *iniSection) allProperties() []*iniProperty {
	properties := []*iniProperty{}
	for _, name := range s.propertyOrder {
		properties = append(properties, s.propertyMap[name])
	}
	return properties
}
