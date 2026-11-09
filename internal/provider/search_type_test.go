package provider

import (
	"testing"
)

func storedSearch(t *testing.T, vendor string) SearchAPI {
	t.Helper()
	for _, a := range StoredSearchAPIs() {
		if a.Vendor == vendor {
			return a
		}
	}
	t.Fatalf("no %s", vendor)
	return SearchAPI{}
}

// Exa can be asked with type lite, its low-cost tier, and with nothing
// else besides its default: another vendor, or a type Exa answers with
// research rather than pages, is refused. The type the user set stays
// through a new key, the Settings window's Add (which sends no type) and
// a backup's restore; auto puts the default back.
func TestSearchAPIType(t *testing.T) {
	claudeHome(t)
	for _, bad := range []SearchAPI{
		{Vendor: "tavily", Key: "tvly-a", Type: "lite"},
		{Vendor: "exa", Key: "exa-a", Type: "deep"},
		{Vendor: "exa", Key: "exa-a", Type: "fast"},
	} {
		if err := SetSearchAPI(bad); err == nil {
			t.Errorf("%s took type %q", bad.Vendor, bad.Type)
		}
	}
	if as := StoredSearchAPIs(); len(as) != 0 {
		t.Fatalf("a refused type was saved: %+v", as)
	}
	if err := SetSearchAPI(SearchAPI{Vendor: "exa", Key: "exa-a"}); err != nil {
		t.Fatal(err)
	}
	if a := storedSearch(t, "exa"); a.Type != "" {
		t.Fatalf("exa got a type nobody set: %q", a.Type)
	}
	if err := SetSearchAPI(SearchAPI{Vendor: "exa", Type: " Lite "}); err != nil {
		t.Fatal(err)
	}
	if a := storedSearch(t, "exa"); a.Type != "lite" || a.Key != "exa-a" {
		t.Fatalf("type lite with no key: %+v", a)
	}
	// a new key alone, from the CLI or from Settings' Add, keeps the type
	if err := SetSearchAPI(SearchAPI{Vendor: "exa", Key: "exa-b", URL: "", Pick: ""}); err != nil {
		t.Fatal(err)
	}
	if a := storedSearch(t, "exa"); a.Type != "lite" || a.Key != "exa-b" {
		t.Fatalf("a new key took the type away: %+v", a)
	}
	saved := StoredSearchAPIs()
	if err := SetSearchAPI(SearchAPI{Vendor: "exa", Type: "auto"}); err != nil {
		t.Fatal(err)
	}
	if a := storedSearch(t, "exa"); a.Type != "" || a.Key != "exa-b" {
		t.Fatalf("type auto: %+v", a)
	}
	// a backup taken while it was lite brings lite back
	if err := RestoreSearchAPIs(saved); err != nil {
		t.Fatal(err)
	}
	if a := storedSearch(t, "exa"); a.Type != "lite" {
		t.Fatalf("restore dropped the type: %+v", a)
	}
}
