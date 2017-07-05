package pool

import (
	"runtime"
	"runtime/debug"
	"testing"
)

func TestPool(t *testing.T) {
	var p Pool
	if p.Get() != nil {
		t.Fatal("expected empty")
	}
	p.Put("a")
	p.Put("b")
	if g := p.Get(); g != "b" {
		t.Fatalf("got %#v; want b", g)
	}
	if g := p.Get(); g != "a" {
		t.Fatalf("got %#v; want a", g)
	}
	if g := p.Get(); g != nil {
		t.Fatalf("got %#v; want nil", g)
	}

	p.Put("c")
	debug.SetGCPercent(100) // to allow following GC to actually run
	runtime.GC()
	if g := p.Get(); g == nil {
		t.Fatalf("got nil; want %#v after GC", g)
	}
}

func TestPoolNew(t *testing.T) {
	i := 0
	p := Pool{
		New: func() interface{} {
			i++
			return i
		},
	}
	if v := p.Get(); v != 1 {
		t.Fatalf("got %v; want 1", v)
	}
	if v := p.Get(); v != 2 {
		t.Fatalf("got %v; want 2", v)
	}
	p.Put(42)
	if v := p.Get(); v != 42 {
		t.Fatalf("got %v; want 42", v)
	}
	if v := p.Get(); v != 3 {
		t.Fatalf("got %v; want 3", v)
	}
}

func TestPoolStress(t *testing.T) {
	const P = 10
	N := int(1e6)
	if testing.Short() {
		N /= 100
	}
	var p Pool
	done := make(chan bool)
	for i := 0; i < P; i++ {
		go func() {
			var v interface{} = 0
			for j := 0; j < N; j++ {
				if v == nil {
					v = 0
				}
				p.Put(v)
				v = p.Get()
				if v != nil && v.(int) != 0 {
					t.Errorf("expect 0, got %v", v)
					break
				}
			}
			done <- true
		}()
	}
	for i := 0; i < P; i++ {
		<-done
	}
}
