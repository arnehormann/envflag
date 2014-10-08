package envflag

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/confactor/envflag/value"
)

// Scan takes a pointer to a struct and recursively traverses struct fields,
// pointers and interface values to build a representation of all
// modifiable values in the data.
// If a memory destination is encountered more than once, only the first occurence
// is contained in Module.
func Scan(structptr interface{}) (Module, error) {
	return scanStructPtr(
		&scanguard{known: make(map[interface{}]struct{})},
		structptr,
	)
}

// ScanWarn is a version of Scan that returns warnings as errors.
//
// If no error occured but values are encountered more than once or
// struct fields are skipped, the error type is *ScanWarnings.
func ScanWarn(structptr interface{}) (Module, error) {
	tracer := &scantracer{
		pointers: make(map[interface{}][]string),
	}
	m, err := scanStructPtr(tracer, structptr)
	if err != nil {
		return nil, err
	}
	warning := tracer.warning()
	if warning != nil {
		return m, warning
	}
	return m, nil
}

var (
	errPtrNil       = errors.New("structptr is nil")
	errNoStructPtr  = errors.New("structptr does not reference a struct")
	errBadStructPtr = errors.New("structptr could not be scanned")
)

type scanner interface {
	// enter must be called when a field is scanned.
	// Each call of enter must be succeeded with leave.
	enter(name string)

	// leave must be called when scanning of a field is done.
	// Each call of leave must be preceded with enter.
	leave(name string)

	// register stores scanned pointers to track duplicate
	// references to memory locations.
	// It reports whether ptr was registered before.
	register(ptr interface{}) bool

	// skip indicates a field is skipped.
	skip()
}

type scanguard struct {
	known map[interface{}]struct{}
}

// could embed path instead
func (s *scanguard) enter(name string) {}
func (s *scanguard) leave(name string) {}

func (s *scanguard) register(ptr interface{}) bool {
	if _, found := s.known[ptr]; found {
		return true
	}
	s.known[ptr] = struct{}{}
	return false
}

func (s *scanguard) skip() {}

func scanStructPtr(scan scanner, ptr interface{}) (*module, error) {
	if ptr == nil {
		return nil, errPtrNil
	}
	value := reflect.ValueOf(ptr)
	if value.Kind() != reflect.Ptr {
		return nil, errNoStructPtr
	}
	// register inital valueto avoid cycles
	scan.register(ptr)
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return nil, errNoStructPtr
	}
	mod := &module{}
	mod.scanChildren(scan, value)
	return mod, nil
}

// scanChildren adds fields of a struct or elements of an array or a slice to mod.
func (mod *module) scanChildren(scan scanner, src reflect.Value) (ok bool) {
	switch src.Kind() {
	case reflect.Array, reflect.Slice:
		for i, max := 0, src.Len(); i < max; i++ {
			name := strconv.FormatInt(int64(i), 10)
			scan.enter(name)
			f := &field{name: name}
			ok := mod.scanValue(scan, f, src.Index(i))
			if !ok {
				scan.skip()
			}
			scan.leave(name)
		}
		return true
	case reflect.Struct:
		t := src.Type()
		for i, max := 0, src.NumField(); i < max; i++ {
			ft := t.Field(i)
			scan.enter(ft.Name)
			f := &field{name: ft.Name, tag: ft.Tag}
			ok := mod.scanValue(scan, f, src.Field(i))
			if !ok {
				scan.skip()
			}
			scan.leave(ft.Name)
		}
		return true
	}
	return false
}

// scanValue adds a single value into a parameter or a module and adds it to mod.
func (mod *module) scanValue(scan scanner, field *field, src reflect.Value) (ok bool) {
	// find pointer to innermost memory destination
	registered := false
	for {
		switch src.Kind() {
		case reflect.Ptr:
			if scan.register(src.Interface()) {
				// pointer is known
				return false
			}
			registered = true
			fallthrough
		case reflect.Interface:
			if src.IsNil() {
				// pointer or interface value is nil
				return false
			}
			src = src.Elem()
			continue
		}
		break
	}
	if !src.CanAddr() {
		// no simple value; struct, array or slice wrapped in interface{}?
		submod := &module{field: *field}
		if submod.scanChildren(scan, src) {
			mod.module = append(mod.module, submod)
			return true
		}
		return false
	}
	addr := src.Addr()
	if !addr.CanInterface() {
		return false
	}
	ptr := addr.Interface()
	if !registered && scan.register(ptr) {
		// pointer is known
		return false
	}
	// check whether src can be used as a parameter
	val, ok := ptr.(value.Value)
	if !ok {
		// not a Value; pointer to valid builtin type?
		val, ok = value.ValueOf(ptr)
	}
	if ok {
		// usable Getter; src is a parameter
		param := &parameter{
			field: *field,
			Value: val,
		}
		mod.param = append(mod.param, param)
		return true
	}
	// not a parameter; struct, array or slice?
	submod := &module{field: *field}
	if submod.scanChildren(scan, src) {
		mod.module = append(mod.module, submod)
		return true
	}
	// unknown type
	return false
}
