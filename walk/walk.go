package walk

import "reflect"

// Walker walks nodes in a graph of values.
type Walker interface {

	// Next enters the next node.
	Next() bool

	// Tag retrieves a tag of the current struct field.
	// It retrieves the full tag when called with an empty string.
	Tag(tag string) string

	// Pointer retrieves a pointer to the current node.
	Pointer() (ptr Pointer, ok bool)

	// Depth retrieves the path length to the root node.
	Depth() int

	// Type retrieves the current node type.
	Type() string

	// Path retrieves the full slash-delimited path to the current node.
	Path() string
}

type level struct {
	Iterator
	depth int
}

type uniqueWalker struct {
	*Crawler
	steps []level
	seen  map[reflect.Value]struct{}
}

var walkedNode = struct{}{}

// WalkUnique retrieves a walker that visits each node once.
// The nodes are walked depth first.
//
// WalkUnique does not follow loops in a graph.
func WalkUnique(ptr Pointer) (Walker, bool) {
	c, ok := NewCrawler(ptr)
	if !ok {
		return nil, false
	}
	return &uniqueWalker{
		Crawler: c,
		steps: []level{
			{c.Iterator(), 0},
		},
		seen: map[reflect.Value]struct{}{
			c.val: walkedNode,
		},
	}, true
}

func (w *uniqueWalker) Next() bool {
	if len(w.steps) == 0 {
		return false
	}
	last := len(w.steps) - 1
	it := w.steps[last]
	for {
		if !it.HasNext() {
			w.steps = w.steps[:last]
			if len(w.steps) == 0 {
				return false
			}
			last = len(w.steps) - 1
			it = w.steps[last]
			w.Crawler.ReturnTo(it.depth)
			continue
		}
		if !it.EnterNext() {
			continue
		}
		if _, crawled := w.seen[w.val]; crawled {
			continue
		}
		w.seen[w.val] = walkedNode
		it = level{w.Iterator(), w.Depth()}
		w.steps = append(w.steps, it)
		last = len(w.steps) - 1
		return true
	}
}

func (w *uniqueWalker) Type() string {
	return w.Crawler.Type(w.Depth())
}

func (w *uniqueWalker) Path() string {
	return string(w.AppendPath([]byte{'/'}, true))
}
