package pool

import "sync"

type Pool struct {
	head *entry
	sync.Mutex
	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New         func() interface{}
	freeEntries *entry
}

type entry struct {
	v    interface{}
	next *entry
}

func (p *Pool) getFreeEntry() *entry {
	e := p.freeEntries
	if e == nil {
		e = new(entry)
	} else {
		p.freeEntries = e.next
	}

	return e
}

func (p *Pool) putFreeEntry(e *entry) {
	e.v = nil
	e.next = p.freeEntries
	p.freeEntries = e
}

// Get selects an arbitrary item from the Pool, removes it from the Pool.
// If Get would otherwise return nil and p.New is non-nil, Get returns the result of calling p.New.
func (p *Pool) Get() interface{} {
	p.Lock()
	e := p.head
	if e == nil {
		p.Unlock()
		if p.New == nil {
			return nil
		}
		return p.New()
	}

	p.head = e.next

	v := e.v
	p.putFreeEntry(e)
	p.Unlock()
	return v
}

// Put adds v to the pool.
func (p *Pool) Put(v interface{}) {
	p.Lock()
	e := p.getFreeEntry()
	e.v = v
	e.next = p.head
	p.head = e
	p.Unlock()
}
