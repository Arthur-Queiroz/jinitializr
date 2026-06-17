package catalog

import "testing"

func TestRouterLookup(t *testing.T) {
	c := New()

	if _, ok := c.Router("stdlib"); !ok {
		t.Error("stdlib router should exist")
	}
	chi, ok := c.Router("chi")
	if !ok {
		t.Fatal("chi router should exist")
	}
	if chi.Module == "" || chi.Version == "" {
		t.Errorf("chi router missing module/version: %+v", chi)
	}
	if _, ok := c.Router("nope"); ok {
		t.Error("unknown router should not be found")
	}
}

func TestDependencyLookup(t *testing.T) {
	c := New()

	pgx, ok := c.Dependency("pgx")
	if !ok {
		t.Fatal("pgx dependency should exist")
	}
	if len(pgx.Files) == 0 {
		t.Error("pgx should declare at least one file")
	}
	if _, ok := c.Dependency("nope"); ok {
		t.Error("unknown dependency should not be found")
	}
}

func TestDefaultRouterIsStdlib(t *testing.T) {
	c := New()
	var defaults int
	for _, r := range c.Routers {
		if r.Default {
			defaults++
			if r.ID != "stdlib" {
				t.Errorf("default router is %q, want stdlib", r.ID)
			}
		}
	}
	if defaults != 1 {
		t.Errorf("got %d default routers, want exactly 1", defaults)
	}
}
