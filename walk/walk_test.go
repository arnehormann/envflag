package walk

import (
	"math"
	"testing"

	"github.com/confactor/envflag/walk/testdata"
)

func TestWalkerNil(t *testing.T) {
	_, ok := WalkUnique(nil)
	if ok {
		t.Fatalf("must not walk nil")
	}
}

func TestWalkerEmpty(t *testing.T) {
	type crawlee struct{}
	src := crawlee{}
	w, ok := WalkUnique(&src)
	if !ok {
		t.Fatalf("must walk pointer")
	}
	if want, got := "/", w.Path(); want != got {
		t.Fatalf("want path %q, got %q", want, got)
	}
	if w.Next() || w.Next() {
		t.Fatalf("had next node (%q), expected none", w.Path())
	}
}

func TestWalkerLoop(t *testing.T) {
	type A struct {
		A interface{}
		B map[string]*A
	}
	x := A{}
	var w Walker
	var ok bool

	x.A = &x
	x.B = map[string]*A{"C": &x}

	w, ok = WalkUnique(&x)
	if !ok {
		t.Fatalf("must walk pointer")
	}

	if !w.Next() {
		t.Fatalf("could not walk unknown node in loop: %q", w.Path())
	}
	if !w.Next() {
		t.Fatalf("could not walk unknown node in loop")
	}
	if !w.Next() {
		t.Fatalf("could not walk unknown node in loop")
	}
	if w.Next() {
		t.Fatalf("could walk known node in loop")
	}
}

func TestWalkerWithHole(t *testing.T) {
	x := []Key{
		0,
		nil,
		uint16(2),
		[2]int8{-3, -4},
		[]float64{math.NaN()},
		testdata.Struct("string", false, false),
	}
	var w Walker
	var ok bool

	w, ok = WalkUnique(&x)
	if !ok {
		t.Fatalf("must walk pointer")
	}
	nodes := []struct{ path, typ string }{
		{"/", "[]walk.Key"},
		{"/0", "walk.Key"},
		{"/0/", "int"},
		{"/1", "walk.Key"},
		{"/2", "walk.Key"},
		{"/2/", "uint16"},
		{"/3", "walk.Key"},
		{"/3/", "[2]int8"},
		{"/3//0", "int8"},
		{"/3//1", "int8"},
		{"/4", "walk.Key"},
		{"/4/", "[]float64"},
		{"/4//0", "float64"},
		{"/5", "walk.Key"},
		{"/5/", "*testdata.dataU"},
		{"/5//", "testdata.dataU"},
		{"/5///v", "interface {}"},
	}
	for i, n := range nodes {
		if w.Path() != n.path {
			t.Fatalf("wanted %s, got %s", n.path, w.Path())
		}
		if w.Type() != n.typ {
			t.Errorf("wanted %s, got %s", n.typ, w.Type())
		}
		next := w.Next()
		if i == len(nodes)-1 && next {
			t.Fatalf("Next was true after the last element: [%d] %s", i, w.Path())
		}
		if i < len(nodes)-1 && !next {
			t.Fatalf("Next was false before the last element: [%d] %s", i, w.Path())
		}
	}
}

func TestWalkerSkipDeleted(t *testing.T) {
	// though deletion of walked values is handled, it is illegal api usage.

	x := map[string]struct{}{"a": struct{}{}, "b": struct{}{}}
	w, ok := WalkUnique(&x)
	if !ok {
		t.Fatalf("must walk pointer")
	}

	if !w.Next() {
		t.Fatalf("walker must have an iterable value")
	}
	delete(x, "b")
	if w.Next() {
		t.Fatalf("after deletion, walker must have no more values left")
	}
}

func TestWalkerUnenterable(t *testing.T) {

	x := map[string]struct{}{"a": struct{}{}, "b": struct{}{}}
	w, ok := WalkUnique(&x)
	if !ok {
		t.Fatalf("must walk pointer")
	}

	if !w.Next() {
		t.Fatalf("walker must have an iterable value")
	}
	delete(x, "b")
	if w.Next() {
		t.Fatalf("after deletion, walker must have no more values left")
	}
}
