package walk

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/confactor/envflag/walk/testdata"
)

func errf(t *testing.T, format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		file = file[strings.LastIndex(file, "/")+1:]
		format = fmt.Sprintf("ERR@[%s:%d] ", file, line) + format
	}
	t.Fatalf(format, args...)
}

type TestA struct{}

type TestB struct {
	TestA
}

type empty struct{}

func (e empty) String() string {
	return "empty"
}

func TestCrawlerCreation(t *testing.T) {
	type crawlee struct{}
	src := crawlee{}
	if _, ok := NewCrawler(&src); !ok {
		t.Fatalf("must crawl pointer")
	}
	if _, ok := NewCrawler(nil); ok {
		t.Fatalf("must not crawl nil")
	}
	if _, ok := NewCrawler(src); ok {
		t.Fatalf("must not crawl non pointer")
	}
	if _, ok := NewCrawler((*crawlee)(nil)); ok {
		t.Fatalf("must not crawl nil pointer")
	}
}

func TestCrawlerProperties(t *testing.T) {
	type S0 struct{}
	type A0 [0]S0
	type S1 struct{ S0 }
	type A1 [1]S0
	checkProps := func(ptr Pointer, size int, ordered bool) {
		c, ok := NewCrawler(ptr)
		if !ok {
			errf(t, "could not crawl %T", ptr)
		}
		if s := c.Size(); s != size {
			errf(t, "want size %d, got %d for %T", size, s, ptr)
		}
		if ordered != c.Ordered() {
			errf(t, "want ordered to be %v", ordered)
		}
	}
	// structs
	s0 := S0{}
	s1 := S1{}
	checkProps(&s0, 0, true)
	checkProps(&s1, 1, true)
	// arrays
	a0 := [0]S0{}
	a1 := [1]S0{}
	a2 := [2]S0{}
	checkProps(&a0, 0, true)
	checkProps(&a1, 1, true)
	checkProps(&a2, 2, true)
	// slices
	sl0 := []S0{}
	sl1 := []S0{{}}
	sl2 := []S0{{}, {}}
	checkProps(&sl0, 0, true)
	checkProps(&sl1, 1, true)
	checkProps(&sl2, 2, true)
	// pointers
	p0 := (*S0)(nil)
	p1 := (*S0)(&s0)
	checkProps(&p0, 0, true)
	checkProps(&p1, 1, true)
	// interfaces
	i0 := Pointer(nil)
	i1 := Pointer(p0)
	checkProps(&i0, 0, true)
	checkProps(&i1, 1, true)
	// maps
	m := make(map[string]string)
	checkProps(&m, 0, false)
	m[""] = ""
	checkProps(&m, 1, false)
	m[" "] = ""
	checkProps(&m, 2, false)
	// anything else, demonstrated by byte and string
	b := byte(0)
	checkProps(&b, 0, false)
	str := "abc"
	checkProps(&str, 0, false)
}

func TestCrawlerDepth(t *testing.T) {
	x := struct {
		D1 struct {
			D2 struct {
				I int
			}
		}
	}{}
	var c *Crawler
	var ok bool
	if c, ok = NewCrawler(&x); !ok {
		t.Fatalf("could not crawl %T", &x)
	}
	d := 0
	path := []byte{'/'}
	for ; d < 3; d++ {
		if c.Depth() != d {
			t.Fatalf("want depth %d at %s, got %d",
				d, string(c.AppendPath(path[:1], false)), c.Depth(),
			)
		}
		if !c.Ordered() || !c.Enter(0) {
			t.Fatalf("could not enter subnode at %d", d)
		}
	}
	c.Leave()
	d--
	if c.Depth() != d {
		t.Fatalf("want depth %d at %s, got %d",
			d, string(c.AppendPath(path[:1], false)), c.Depth(),
		)
	}
}

func TestCrawlerBasicTypes(t *testing.T) {
	x := struct {
		B    bool       `idx:"0" type:"bool" name:"B"`
		I8   int8       `idx:"1" type:"int8" name:"I8"`
		I16  int16      `idx:"2" type:"int16" name:"I16"`
		I32  int32      `idx:"3" type:"int32" name:"I32"`
		I64  int64      `idx:"4" type:"int64" name:"I64"`
		I    int        `idx:"5" type:"int" name:"I"`
		U8   uint8      `idx:"6" type:"uint8" name:"U8"`
		U16  uint16     `idx:"7" type:"uint16" name:"U16"`
		U32  uint32     `idx:"8" type:"uint32" name:"U32"`
		U64  uint64     `idx:"9" type:"uint64" name:"U64"`
		F32  float32    `idx:"10" type:"float32" name:"F32"`
		F64  float64    `idx:"11" type:"float64" name:"F64"`
		C64  complex64  `idx:"12" type:"complex64" name:"C64"`
		C128 complex128 `idx:"13" type:"complex128" name:"C128"`
	}{}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	size := int64(c.Size())
	if size != 14 {
		t.Errorf("number of fields does not match")
	}
	if c.Tag("") != "" {
		t.Errorf("does not have a tag")
	}
	for i := int64(0); i < size; i++ {
		if ok := c.Enter(int(i)); !ok {
			t.Errorf("could not enter field #%d", i)
		}
		if c.Tag("") == "" {
			t.Errorf("unexpected empty tag for field #%d", i)
		}
		idx := c.Tag("idx")
		if j, err := strconv.ParseInt(idx, 10, 64); err != nil || i != j {
			t.Errorf("wrong tag idx: %s, should be %d", idx, i)
		}
		name := c.Tag("name")
		if name == "" || name != c.Tag("name") {
			t.Errorf("wrong name or tag 'name'")
		}
		if _, ok := c.Pointer(); !ok {
			t.Errorf("field #%d must be addressable", i)
		}
		if size := c.Size(); size != 0 {
			t.Errorf("field #%d must have size 0, is %d", i, size)
		}
		c.Leave()
	}
}

func TestCrawlerOutOfBounds(t *testing.T) {
	checkBadKey := func(t *testing.T, ptr Pointer, key Key) {
		c, ok := NewCrawler(ptr)
		if !ok {
			errf(t, "could not crawl %T", ptr)
		}
		if c.Enter(key) {
			errf(t, "must not enter unavailable key %s", key)
		}
	}
	x := struct {
		F struct{}
	}{}
	xptr := &x
	checkBadKey(t, &x, 1)
	checkBadKey(t, &x, "f")
	checkBadKey(t, &xptr, 1)
}

/*
func TestCrawlerStructKey(t *testing.T) {
	x := struct {
		F struct{}
	}{}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	var keys [2]reflect.StructField
	for i, key := range []Key{0, "F"} {
		if !c.Enter(key) {
			t.Fatalf("could not enter %v", key)
		}
		if keys[i], ok = c.Key(0).(reflect.StructField); !ok {
			t.Fatalf("key for %s is not a struct field", key)
		}
		c.Leave()
	}
	if !reflect.DeepEqual(keys[0], keys[1]) {
		t.Fatalf("keys %v and %v are not equal", keys[0], keys[1])
	}
}
*/

func TestCrawlerEnter(t *testing.T) {
	x := struct {
		A0 [0]struct{}
		A1 [1]struct{}
		A2 [2]struct{}
		S  []struct{}
		I  interface{}
		M  map[string]string
		X  int
		x  int
	}{}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	checkSize := func(size int, path ...Key) {
		for d, key := range path {
			if !c.Enter(key) {
				errf(t, "could not enter %#v", path[:d])
			}
		}
		if c.Size() != size {
			errf(t, "want %#v size %d, got %d", path, size, c.Size())
		}
		for d := 0; d < len(path); d++ {
			c.Leave()
		}
	}
	// root (struct)
	checkSize(8)
	// array
	checkSize(0, "A0")
	checkSize(1, "A1")
	checkSize(2, "A2")
	// slice
	checkSize(0, "S")
	x.S = x.A2[:]
	checkSize(2, "S")
	// map
	checkSize(0, "M")
	x.M = make(map[string]string)
	checkSize(0, "M")
	x.M["m"] = ""
	checkSize(1, "M")
	// interface
	checkSize(0, "I")
	x.I = x.A2
	checkSize(1, "I")
	// inner values
	checkSize(0, "M", "m")
	checkSize(0, "A2", 1)
	checkSize(0, "S", 1)
	checkSize(2, "I", Elem)
	if c.Enter("X") && c.Enter(nil) {
		t.Fatalf("must not enter X")
	}
	c.Leave()
	if !c.Enter("x") {
		t.Fatalf("must not enter x")
	}
	if _, ok := c.Pointer(); ok {
		t.Fatalf("Pointer on unexported var must fail")
	}
	c.Leave()
	// test map key
	path := []Key{"M", "m"}
	depth := c.Depth()
	if err := c.Into(path...); err != nil {
		errf(t, "could not enter %#v, %s", path, err)
	}
	if _, ok := c.Pointer(); ok {
		t.Fatalf("Pointer on map value must fail")
	}
	c.ReturnTo(depth)
}

func TestCrawlerReturnToPanics(t *testing.T) {
	x := struct{ A int }{}
	c, _ := NewCrawler(&x)
	// negative value
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Fatal("expected panic for negative argument to ReturnTo")
			}
		}()
		c.ReturnTo(-1)
	}()
	func() {
		defer func() {
			if p := recover(); p == nil {
				t.Fatal("expected panic for depth > current depth in ReturnTo")
			}
		}()
		c.ReturnTo(1)
	}()
}

func TestCrawlerInto(t *testing.T) {
	type testcase struct {
		path    []Key
		failIdx int
		depth   int
		errs    bool
	}
	runTest := func(c *Crawler, test *testcase, i int) {
		defer c.ReturnTo(c.Depth())
		err := c.Into(test.path...)
		if err == nil {
			if test.errs {
				errf(t, "case %d: expected an error", i)
			}
			if test.failIdx > 0 {
				errf(t, "case %d: should not enter %#v but fail at %d",
					i, test.path, test.failIdx,
				)
			}
		}
		if err != nil {
			if !test.errs {
				errf(t, "case %d: expected no error, got: %s",
					i, err,
				)
			}
			if err.Index != test.failIdx {
				errf(t, "case %d: could not enter %#v; %s",
					i, test.path, err,
				)
			}
		}
		want, got := test.depth, c.Depth()
		if want != got {
			errf(t, "case %d: expected depth %d, got %d",
				i, want, got,
			)
		}
	}

	x := struct {
		Ptr       *struct{ TestA }
		Interface interface{}
		Map       map[int]struct{ TestA }
	}{}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}

	tests := []testcase{
		{path: []Key{"Ptr"}, depth: 1},
		{path: []Key{"Interface"}, depth: 1},
		{path: []Key{"Map"}, depth: 1},
		{path: []Key{"Ptr", "TestA"}, failIdx: 1, errs: true},
		{path: []Key{"Interface", "TestA"}, failIdx: 1, errs: true},
		{path: []Key{"Map", 0, "TestA"}, failIdx: 1, errs: true},
	}
	for i := range tests {
		runTest(c, &tests[i], i)
	}

	v := struct{ TestA }{}
	x.Ptr = &v
	x.Interface = v
	x.Map = map[int]struct{ TestA }{0: v}

	tests = []testcase{
		{path: []Key{"Ptr", "TestA"}, depth: 3},
		{path: []Key{"Interface", "TestA"}, depth: 3},
		{path: []Key{"Map", 0, "TestA"}, depth: 3},
	}
	for i := range tests {
		runTest(c, &tests[i], i)
	}
}

func TestCrawlerIntoError(t *testing.T) {
	x := struct{ Not int }{}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	err := c.Into("Not", "valid", "here")
	if err == nil {
		t.Fatalf("expected an error")
	}
	if err.Entered != 1 || err.Index != 1 {
		t.Fatalf("expected different error metrics")
	}
	if errstr := err.Error(); errstr == "" {
		t.Fatalf("error was an empty string")
	}
	c.ReturnTo(0)
}

func TestCrawlerCircular(t *testing.T) {
	const maxdepth = 1000
	var x interface{}
	x = &x
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	for i := 0; i < maxdepth; i++ {
		if i != c.Depth() {
			t.Fatalf("want depth %d, got %d", i, c.Depth())
		}
		if !c.Enter(0) {
			t.Fatalf("could not enter at depth %d", i)
		}
	}
	for i := maxdepth; i >= 0; i-- {
		if i != c.Depth() {
			t.Fatalf("want depth %d, got %d", i, c.Depth())
		}
		c.Leave()
	}
}

func TestCrawlerAppendPath(t *testing.T) {
	type F1 struct{ F10 struct{} }
	type f3 struct{ f30 struct{} }

	x := struct {
		F0 [1]int
		F1
		F2 map[interface{}]string
		f3
	}{
		F2: map[interface{}]string{"string": "", empty{}: ""},
	}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	for _, test := range []struct {
		want string
		path []Key
	}{
		{"F0", []Key{0}},
		{"F0", []Key{"F0"}},
		{"F0/0", []Key{0, 0}},
		{"F0/0", []Key{"F0", 0}},
		{"F1", []Key{"F1"}},
		{"F1/F10", []Key{1, 0}},
		{"F1/F10", []Key{"F1", 0}},
		{"F1/F10", []Key{1, "F10"}},
		{"F1/F10", []Key{"F1", "F10"}},
		{"F10", []Key{"F10"}}, // embedded
		{"F2", []Key{"F2"}},
		{"F2/string", []Key{"F2", "string"}},
		{"F2/empty", []Key{"F2", empty{}}},
		{"f3", []Key{"f3"}},
		{"f3/f30", []Key{"f3", "f30"}},
	} {
		for d, key := range test.path {
			if !c.Enter(key) {
				errf(t, "could not enter %#v", test.path[:d])
			}
		}
		got := string(c.AppendPath(nil, false))
		if test.want != got {
			t.Fatalf("want path %s, got %s", test.want, got)
		}
		c.ReturnTo(0)
	}
	// check embedded struct fields
	c.Into("F10")
	if got, want := "F1/F10", string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %s, got %s", want, got)
	}
	c.ReturnTo(0)
	c.Into("f30")
	if want, got := "f3/f30", string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %s, got %s", want, got)
	}
	c.ReturnTo(0)
}

func TestCrawlerPathEscaping(t *testing.T) {
	x := map[string]map[string]string{
		`/`:  map[string]string{`/`: ""},
		`\`:  map[string]string{`/\`: "", `\/`: ""},
		`//`: map[string]string{`//`: ""},
		`\\`: map[string]string{`\\`: ""},
	}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}

	c.Into(`/`, `/`)
	if want, got := `\//\/`, string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %q, got %q", want, got)
	}
	c.ReturnTo(0)

	c.Into(`\`, `/\`)
	if want, got := `\\/\/\\`, string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %q, got %q", want, got)
	}
	c.ReturnTo(0)

	c.Into(`\`, `\/`)
	if want, got := `\\/\\\/`, string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %q, got %q", want, got)
	}
	c.ReturnTo(0)

	c.Into(`//`, `//`)
	if want, got := `\/\//\/\/`, string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %q, got %q", want, got)
	}
	c.ReturnTo(0)

	c.Into(`\\`, `\\`)
	if want, got := `\\\\/\\\\`, string(c.AppendPath(nil, true)); want != got {
		t.Fatalf("want path %q, got %q", want, got)
	}
	c.ReturnTo(0)
}

func TestCrawlerExternal(t *testing.T) {
	type testcase struct {
		base   Pointer
		key    Key
		want   interface{}
		canGet bool
	}
	newCase := func(val interface{}, exportedStruct, exportedField bool) *testcase {
		field := "v"
		if exportedField {
			field = "V"
		}
		return &testcase{
			base:   testdata.Struct(val, exportedStruct, exportedField),
			key:    field,
			want:   val,
			canGet: exportedField,
		}
	}
	cases := []*testcase{
		newCase(1, true, true),
		newCase(2, true, false),
		newCase(3, false, true),
		newCase(4, false, false),
	}

	for i := range cases {
		tc := cases[i]
		c, ok := NewCrawler(tc.base)
		if !ok {
			t.Fatalf("could not crawl %T", tc.base)
		}

		if !c.Enter(tc.key) {
			t.Fatalf("could not enter %s", tc.key)
		}
		for c.Enter(struct{}{}) {
		}
		//if !c.Enter(0) && !c.Enter(0) {
		//	t.Fatalf("could not get value inside interface{} in %T.%s", tc.base, tc.key)
		//}
		if ptr, ok := c.Pointer(); ok {
			if !tc.canGet {
				t.Fatalf("Must not be able to get a pointer to %T.%s", tc.base, tc.key)
			}
			pi, ok := ptr.(*int)
			if !ok {
				t.Fatalf("wrong type in %T.%s, want %#v, got %#v", tc.base, tc.key, tc.want, ptr)
			}
			val := *pi
			if val != tc.want {
				t.Fatalf("wrong value in %T.%s, want %#v, got %#v", tc.base, tc.key, tc.want, val)
			}
		} else if tc.canGet {
			//t.Fatalf("Could not get a pointer to %T.%s; %s", tc.base, tc.key, c.val)
		}
		c.Leave()
		c.Leave()
		c.Leave()
	}
}

func TestCrawlerType(t *testing.T) {
	const want = "walk.TestA"
	x := &TestB{}
	c, ok := NewCrawler(x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}

	if !c.Enter(0) || c.Depth() != 1 {
		t.Fatalf("wrong depth")
	}
	if want, got := "walk.TestB", c.Type(0); want != got {
		t.Fatalf("wrong type name, want %q, got %q", want, got)
	}
	if want, got := "walk.TestA", c.Type(1); want != got {
		t.Fatalf("wrong type name, want %q, got %q", want, got)
	}
}

func TestCrawlerDontPanicOnBadMapKey(t *testing.T) {
	x := map[string]string{"A": "B"}
	c, ok := NewCrawler(&x)
	if !ok {
		t.Fatalf("could not crawl %T", x)
	}
	if c.Enter(0) {
		// getting here or panicing fails the test
		t.Fatalf("invalid map key type")
	}
}
