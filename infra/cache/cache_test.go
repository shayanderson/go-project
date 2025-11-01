package cache

import "testing"

func TestCache(t *testing.T) {
	type item struct{ ID, name string }
	c := New[item, string]()
	if c == nil {
		t.Fatal("expected cache to be created")
	}
	if c.metrics == nil {
		t.Fatal("expected metrics to be initialized")
	}
	if c.store == nil {
		t.Fatal("expected store to be initialized")
	}
	if len(c.store) != 0 {
		t.Fatalf("expected store to be empty, got %d", len(c.store))
	}

	k := "k1"
	v := item{ID: k, name: "item1"}
	c.Put(k, v)
	if size := c.Size(); size != 1 {
		t.Fatalf("expected size to be 1, got %d", size)
	}

	got, ok := c.Get(k)
	if !ok {
		t.Fatalf("expected to find key %s", k)
	}
	if got != v {
		t.Fatalf("expected value %+v, got %+v", v, got)
	}

	m := c.Metrics()
	if m.Hits != 1 {
		t.Fatalf("expected hits to be 1, got %d", m.Hits)
	}
	if m.Misses != 0 {
		t.Fatalf("expected misses to be 0, got %d", m.Misses)
	}

	c.Delete(k)
	if size := c.Size(); size != 0 {
		t.Fatalf("expected size to be 0 after delete, got %d", size)
	}

	_, ok = c.Get(k)
	if ok {
		t.Fatalf("expected key %s to be deleted", k)
	}
	m = c.Metrics()
	if m.Hits != 1 {
		t.Fatalf("expected hits to be 1, got %d", m.Hits)
	}
	if m.Misses != 1 {
		t.Fatalf("expected misses to be 1, got %d", m.Misses)
	}

	c.Put(k, v)
	c.Put("k2", item{ID: "k2", name: "item2"})
	if size := c.Size(); size != 2 {
		t.Fatalf("expected size to be 2, got %d", size)
	}
	c.Clear()
	if size := c.Size(); size != 0 {
		t.Fatalf("expected size to be 0 after clear, got %d", size)
	}
}
