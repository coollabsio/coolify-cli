package wireguard

import (
	"reflect"
	"sort"
	"testing"
)

func TestBuilderHostSet_EnableBuilderAppliesToAll(t *testing.T) {
	d := &DesiredMesh{
		Hosts:         []string{"a", "b", "c"},
		EnableBuilder: true,
	}
	got := d.BuilderHostSet()
	want := map[string]bool{"a": true, "b": true, "c": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBuilderHostSet_ExplicitListWinsOverEnable(t *testing.T) {
	d := &DesiredMesh{
		Hosts:         []string{"a", "b", "c"},
		EnableBuilder: true,
		BuilderHosts:  []string{"b"},
	}
	got := d.BuilderHostSet()
	want := map[string]bool{"b": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBuilderHostSet_FiltersToServersOnly(t *testing.T) {
	// A --builder-hosts entry not present in --servers is dropped.
	d := &DesiredMesh{
		Hosts:        []string{"a", "b"},
		BuilderHosts: []string{"a", "z"},
	}
	got := d.BuilderHostSet()
	want := map[string]bool{"a": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBuilderHostSet_DefaultDisabled(t *testing.T) {
	d := &DesiredMesh{Hosts: []string{"a"}}
	if len(d.BuilderHostSet()) != 0 {
		t.Fatalf("want empty set, got %v", d.BuilderHostSet())
	}
	if d.HasBuilderCap("a") {
		t.Fatalf("HasBuilderCap should be false by default")
	}
}

func TestBuilderHostSet_EnableBuilderFalse_NoBuilderHosts(t *testing.T) {
	d := &DesiredMesh{
		Hosts:         []string{"a"},
		EnableBuilder: false,
	}
	if len(d.BuilderHostSet()) != 0 {
		t.Fatalf("want empty set, got %v", d.BuilderHostSet())
	}
}

func TestBuilderHostSet_Stable(t *testing.T) {
	// Test that calling twice produces the same set (sanity — no side effects).
	d := &DesiredMesh{
		Hosts:        []string{"a", "b"},
		BuilderHosts: []string{"a"},
	}
	a := keys(d.BuilderHostSet())
	b := keys(d.BuilderHostSet())
	sort.Strings(a)
	sort.Strings(b)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("unstable: %v vs %v", a, b)
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
