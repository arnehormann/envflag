package walk

import "reflect"

// Iterator knows all keys for a Crawler node and can enter them sequentially.
// It is only valid for the node it was retrieved at.
type Iterator interface {
	HasNext() bool
	EnterNext() bool
	Leave()
	Reset()
}

// emptyIterator does nothing and always returns false.
type emptyIterator struct{}

// seqIterator iterates on nodes with ordered descendants
// (slice, array, struct, pointer, interface).
type seqIterator struct {
	*Crawler
	i int
}

// mapIterator iterates on maps.
type mapIterator struct {
	*Crawler
	keys []reflect.Value
	i    int
}

// enforce interface conformity
var (
	_ Iterator = emptyIterator{}
	_ Iterator = &seqIterator{}
	_ Iterator = &mapIterator{}
)

// Iterator provides a way to enter unknown key sets, e.g. keys of a map.
// An iterator must only be used when the crawler is at the node the iterator was retrieved at.
func (c *Crawler) Iterator() Iterator {
	if c.Size() > 0 {
		if c.Ordered() {
			return &seqIterator{
				Crawler: c,
			}
		}
		if c.val.Kind() == reflect.Map {
			return &mapIterator{
				Crawler: c,
				keys:    c.val.MapKeys(),
			}
		}
	}
	return emptyIterator{}
}

func (i emptyIterator) HasNext() bool {
	return false
}

func (i emptyIterator) EnterNext() bool {
	return false
}

func (i emptyIterator) Leave() {
}

func (i emptyIterator) Reset() {
}

func (i *seqIterator) HasNext() bool {
	return i.i < i.Size()
}

func (i *seqIterator) EnterNext() bool {
	ii := i.i
	if ii >= i.Size() {
		return false
	}
	i.i++
	return i.Enter(ii)
}

func (i *seqIterator) Reset() {
	i.i = 0
}

func (i *mapIterator) HasNext() bool {
	return i.i < len(i.keys)
}

func (i *mapIterator) EnterNext() bool {
	ii := i.i
	if ii >= len(i.keys) {
		return false
	}
	i.i++
	return i.Enter(mapKey{i.keys[ii]})
}

func (i *mapIterator) Reset() {
	i.i = 0
}
