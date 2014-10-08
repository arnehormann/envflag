package envflag

import (
	"reflect"

	"github.com/confactor/envflag/value"
)

type Field interface {
	Name() string

	// Tag retrieves a tag.
	// The key "" retrieves all tags.
	Tag(key string) string
}

type Value interface {
	Set(string) error
	String() string
}

// Parameter is a configurable value.
type Parameter interface {
	Field
	Value
}

// Module is a collection of modules and parameters.
type Module interface {
	Field
	Module(name string) (Module, bool)
	Modules() []Module
	Parameter(name string) (Parameter, bool)
	Parameters() []Parameter
}

type boolValue interface {
	value.Value
	IsBoolFlag() bool
}

type field struct {
	name string
	tag  reflect.StructTag
}

// parameter is a configurable value.
type parameter struct {
	field
	value.Value
}

// module is a collection of configurable values and other modules.
type module struct {
	field
	module []*module
	param  []*parameter
}

type path struct {
	path []byte
}

func (p *path) enter(name string) {
	if len(p.path) > 0 {
		p.path = append(p.path, '/')
	}
	p.path = append(p.path, name...)
}

func (p *path) leave(name string) {
	// no check for name equality, this is internal only.
	if prev := len(p.path) - (len(name) + 1); prev > 0 {
		p.path = p.path[:prev]
	} else {
		p.path = p.path[:0]
	}
}

func (p *path) String() string {
	return string(p.path)
}

func (f *field) Name() string {
	return f.name
}

func (f *field) Tag(key string) string {
	if key == "" {
		return string(f.tag)
	}
	return f.tag.Get(key)
}

func (m *module) Module(name string) (Module, bool) {
	for _, m := range m.module {
		if m.name == name {
			return m, true
		}
	}
	return nil, false
}

func (m *module) Modules() []Module {
	ms := make([]Module, len(m.module))
	for i := range m.module {
		ms[i] = m.module[i]
	}
	return ms
}

func (m *module) Parameter(name string) (Parameter, bool) {
	for _, p := range m.param {
		if p.name == name {
			return p, true
		}
	}
	return nil, false
}

func (m *module) Parameters() []Parameter {
	ps := make([]Parameter, len(m.param))
	for i := range m.param {
		ps[i] = m.param[i]
	}
	return ps
}

func (m *module) String() string {
	buf := [2048]byte{}
	return string(m.appendIndented(buf[:], ""))
}

func (m *module) appendIndented(buf []byte, path string) []byte {
	path = path + m.name + "/"
	buf = append(buf, (path + "\n")...)
	for i := range m.module {
		buf = m.module[i].appendIndented(buf, path)
	}
	for i := range m.param {
		buf = append(buf, (path + m.param[i].name + "\n")...)
	}
	return buf
}
