package walk

import (
	"reflect"
	"strconv"
)

// Key uniquely identifies an edge from one value to another in a graph of values.
//
// For arrays, slices and struct fields, the key is the element index.
// In structs, it can also be the field name. For a map, it is the map key.
// Every value can serve as a key to the element of a pointer or an interface.
type Key interface{}

// Elem is the canonical key of a value referenced by a pointer or an interface.
var Elem = struct{}{}

// Pointer is the memory address of a value.
type Pointer interface{}

// Crawler visits a data graph starting with a pointer as the root node.
//
// It can descend into struct fields, elements of arrays, slices and maps
// and into elements referenced by a pointer or an interface.
//
// If data is changed while it is crawled, the crawler becomes invalid
// unless the node that was changed is left and reentered.
type Crawler struct {
	path []edge
	val  reflect.Value
	size int
}

// edge leads to a node.
type edge struct {
	src reflect.Value
	key Key
}

// mapKey is a map Key, provided as a fastpath for mapIterator.
type mapKey struct {
	reflect.Value
}

// NewCrawler retrieves a crawler for ptr.
//
// It will fail if ptr is not a non nil pointer.
func NewCrawler(ptr Pointer) (c *Crawler, ok bool) {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil, false
	}
	val = val.Elem()
	return &Crawler{
		val:  val,
		size: size(val),
	}, true
}

// size determines the number of descendant nodes.
func size(val reflect.Value) int {
	if !val.CanInterface() {
		return 0
	}
	switch val.Kind() {
	case reflect.Array,
		reflect.Slice,
		reflect.Map:
		return val.Len()
	case reflect.Struct:
		return val.NumField()
	case reflect.Interface,
		reflect.Ptr:
		if !val.IsNil() {
			return 1
		}
	}
	return 0
}

// Size retrieves the number of enterable edges at the current node.
//
// Changing the graph invalidates Size until the node is left and entered again.
func (c *Crawler) Size() int {
	return c.size
}

// Depth retrieves the number of nodes entered and not yet left.
func (c *Crawler) Depth() int {
	return len(c.path)
}

// Ordered reports whether numbers in [0, Size) can be used as keys in Enter.
func (c *Crawler) Ordered() bool {
	switch c.val.Kind() {
	case reflect.Array,
		reflect.Slice,
		reflect.Struct,
		reflect.Interface,
		reflect.Ptr:
		return true
	}
	return false
}

// Enter descends into a node and reports whether it succeeded.
func (c *Crawler) Enter(key Key) bool {
	var val reflect.Value
	switch k := key.(type) {
	// type checks are useless if c.val is a map; handled in the !IsValid check
	case int:
		if k < 0 || k >= c.size {
			// out of bounds access in reflect panics; Enter must not
			return false
		}
		switch c.val.Kind() {
		case reflect.Array,
			reflect.Slice:
			val = c.val.Index(k)
		case reflect.Struct:
			key = c.val.Type().Field(k)
			val = c.val.Field(k)
		}
	case string:
		switch c.val.Kind() {
		case reflect.Struct:
			sf, ok := c.val.Type().FieldByName(k)
			if !ok {
				return false
			}
			key, val = sf, c.val.FieldByIndex(sf.Index)
		}
	}
	if !val.IsValid() {
		switch c.val.Kind() {
		case reflect.Interface,
			reflect.Ptr:
			key = Elem
			val = c.val.Elem()
		case reflect.Map:
			k, ok := key.(mapKey)
			if !ok {
				kv := reflect.ValueOf(key)
				if !kv.Type().AssignableTo(c.val.Type().Key()) {
					// wrong key type for map
					return false
				}
				k = mapKey{kv}
			}
			key = k
			val = c.val.MapIndex(k.Value)
		}
		if !val.IsValid() {
			// key not present, pointer or interface is nil
			return false
		}
	}
	c.path = append(
		c.path,
		edge{
			src: c.val,
			key: key,
		},
	)
	c.val, c.size = val, size(val)
	return true
}

// Leave reverts the last successful Enter.
func (c *Crawler) Leave() {
	if end := len(c.path) - 1; end >= 0 {
		src := c.path[end].src
		c.path = c.path[:end]
		c.val, c.size = src, size(src)
	}
}

// Pointer retrieves a pointer to the current node.
func (c *Crawler) Pointer() (ptr Pointer, ok bool) {
	if v := c.val; v.CanAddr() {
		v = v.Addr()
		if v.CanInterface() {
			return v.Interface(), true
		}
	}
	return nil, false
}

// Tag retrieves the struct tag of the current node.
//
// It retrieves the full tag when key is an emtpy string.
func (c *Crawler) Tag(key string) string {
	if len(c.path) == 0 {
		return ""
	}
	field, ok := c.path[len(c.path)-1].key.(reflect.StructField)
	if ok && key != "" {
		return field.Tag.Get(key)
	}
	return string(field.Tag)
}

// Key retrieves the key used to leave the node at the given depth.
//
// The retrieved key must not be modified.
//
// Key panics if no key to a node exists at that depth.
func (c *Crawler) Key(depth int) Key {
	p := c.path[depth]
	mk, ismap := p.key.(mapKey)
	if !ismap {
		return p.key
	}
	// TODO: is !mk.CanInterface() possible?
	return mk.Interface()
}

// Type retrieves the type of the node at the given depth.
//
// Type panics if no node exists at that depth.
func (c *Crawler) Type(depth int) string {
	if depth == len(c.path) {
		return c.val.Type().String()
	}
	return c.path[depth].src.Type().String()
}

// appendEscaped appends str to desc and escapes "\" as "\\" and "/" as "\/".
func appendEscaped(dest []byte, str string) []byte {
	o := 0
	for i := 0; i < len(str); i++ {
		switch b := str[i]; b {
		case '\\', '/':
			dest = append(dest, str[o:i]...)
			o = i + 1
			dest = append(dest, '\\', b)
		}
	}
	return append(dest, str[o:len(str)]...)
}

type stringer interface {
	String() string
}

// AppendPath appends the path from root to the current node.
//
// The path is delimited by slashes ("/"). "\" and "/" are escaped as "\\" and "\/".
// Unprintable path segments, e.g. pointer dereferences, are delimited but not added.
//
// A full path contains the names of all embedded struct fields along a path,
// regardless of their usage in Enter.
func (c *Crawler) AppendPath(dest []byte, full bool) []byte {
	for i := 0; i < len(c.path); i++ {
		if i > 0 {
			dest = append(dest, '/')
		}
		switch key := c.Key(i).(type) {
		case int:
			dest = strconv.AppendInt(dest, int64(key), 10)
		case reflect.StructField:
			if full {
				base := c.path[i].src.Type().Field(key.Index[0])
				dest = append(dest, base.Name...)
				for _, f := range key.Index[1:] {
					base = base.Type.Field(f)
					dest = append(dest, '/')
					dest = append(dest, base.Name...)
				}
			} else {
				dest = append(dest, key.Name...)
			}
		case string:
			dest = appendEscaped(dest, key)
		case stringer:
			dest = appendEscaped(dest, key.String())
		}
	}
	return dest
}

// ErrorInto provides details about a failure of a call to Into.
type ErrorInto struct {
	// Index of the path segment where Into failed.
	Index int
	// Number of nodes Into could enter, including pointers and interfaces.
	Entered int
}

func (e *ErrorInto) Error() string {
	return "error following path, " +
		"invalid key at index " +
		strconv.FormatInt(int64(e.Index), 10) +
		", " +
		"failed after " +
		strconv.FormatInt(int64(e.Entered), 10) +
		" nodes"
}

// ReturnTo returns to the node at the given depth.
//
// It panics if the crawler is at a lower depth already or if dpeth is negative.
func (c *Crawler) ReturnTo(depth int) {
	cd := c.Depth()
	if depth < 0 || cd < depth {
		// callers can only end up here with wrong state tracking.
		// panic indicates this loudly and support this convenient idiom:
		// defer c.ReturnTo(c.Depth())
		panic("bad depth: " + strconv.FormatInt(int64(depth), 10))
	}
	for cd -= depth; cd > 0; cd-- {
		c.Leave()
	}
}

// Follow enters pointers and interfaces and returns wether it reached a value.
//
// Use ReturnTo with the prior depth to return the crawler to the original node.
func (c *Crawler) Follow() (ok bool) {
	for {
		switch c.val.Kind() {
		case reflect.Interface,
			reflect.Ptr:
			if !c.Enter(Elem) {
				return false
			}
		}
		return true
	}
}

// Into follows a path to a node.
//
// Pointers and interfaces leading to each key are automatically entered
// and must not be part of the path.
// Use ReturnTo with the prior depth to return the crawler to the original node.
//
// If Into fails, the crawler returns to the original node.
func (c *Crawler) Into(path ...Key) *ErrorInto {
	prevdepth := c.Depth()
	for i, key := range path {
		if !c.Follow() || !c.Enter(key) {
			reldepth := c.Depth() - prevdepth
			c.ReturnTo(prevdepth)
			return &ErrorInto{
				Index:   i,
				Entered: reldepth,
			}
		}
	}
	return nil
}
