package walk

import "testing"

func testIterator(t *testing.T, ptr interface{}, size int) {
	c, ok := NewCrawler(ptr)
	if !ok {
		errf(t, "could not crawl %T", ptr)
	}

	rit := c.Iterator()
	i, max := 0, c.Size()
	if max != size {
		errf(t, "expected size %d, got %d", size, max)
	}
	exit := false
	for i <= max {
		if !rit.HasNext() {
			if rit.EnterNext() {
				errf(t, "EnterNext() succeeded though HasNext() returned false")
			}
			if exit {
				return
			}
			if i != max {
				errf(t, "wrong number of iterations: %d instead of %d", i, max)
			}
			rit.Reset()
			exit = true
			i = 0
			continue
		}
		if !rit.EnterNext() {
			errf(t, "could not enter step %d", i)
		}
		rit.Leave()
		i++
	}
}

func TestIteratorInt(t *testing.T) {
	var x int
	testIterator(t, &x, 0)
}

func TestIteratorPointer(t *testing.T) {
	x0 := (*struct{})(nil)
	testIterator(t, &x0, 0)
	x1 := &x0
	testIterator(t, &x1, 1)
}

func TestIteratorInterface(t *testing.T) {
	var x0, x1 interface{}
	testIterator(t, &x0, 0)
	x1 = &x0
	testIterator(t, &x1, 1)
}

func TestIteratorsSlice(t *testing.T) {
	x0 := []string{}
	testIterator(t, &x0, len(x0))
	x1 := []string{"a", "b", "c"}
	testIterator(t, &x1, len(x1))
}

func TestIteratorArray(t *testing.T) {
	x0 := [0]string{}
	testIterator(t, &x0, len(x0))
	x1 := [3]string{"a", "b", "c"}
	testIterator(t, &x1, len(x1))
}

func TestIteratorsMap(t *testing.T) {
	x0 := map[string]int{}
	testIterator(t, &x0, len(x0))
	x1 := map[string]int{
		"a": 0,
		"b": 1,
		"c": 2,
	}
	testIterator(t, &x1, len(x1))
}

func TestIteratorStruct(t *testing.T) {
	x0 := struct{}{}
	testIterator(t, &x0, 0)
	x1 := struct{ A, B, C int }{}
	testIterator(t, &x1, 3)
}

func TestIteratorInvalid(t *testing.T) {
	c := &Crawler{}
	rit := c.Iterator()
	if rit.HasNext() || rit.EnterNext() {
		t.Fatalf("must not iterate empty interface{}")
	}
	// must not panic
	rit.Leave()
}
