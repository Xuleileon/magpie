package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// searchListOut is what `magpie search` prints.
func searchListOut(t *testing.T) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() { b, _ := io.ReadAll(r); done <- string(b) }()
	err = searchCmd(nil)
	w.Close()
	os.Stdout = old
	out := <-done
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// magpie search add exa <key> type=lite asks Exa's low-cost tier; a new
// key alone keeps it, type= or type=auto puts Exa's default back, and a
// type for another API, or one Exa answers with research, is refused.
func TestSearchCLIType(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	exa := func() provider.SearchAPI {
		for _, a := range provider.StoredSearchAPIs() {
			if a.Vendor == "exa" {
				return a
			}
		}
		t.Fatal("no exa")
		return provider.SearchAPI{}
	}
	for _, bad := range [][]string{{"add", "tavily", "tvly-1", "type=lite"}, {"add", "exa", "e-1", "type=deep"}, {"add", "exa", "e-1", "type=lite", "extra"}} {
		if err := searchCmd(bad); err == nil {
			t.Errorf("%v taken", bad)
		}
	}
	if err := searchCmd([]string{"add", "exa", "type=lite", "e-1"}); err != nil {
		t.Fatal(err)
	}
	if a := exa(); a.Type != "lite" || a.Key != "e-1" {
		t.Fatalf("type=lite: %+v", a)
	}
	if out := searchListOut(t); !strings.Contains(out, "lite") {
		t.Errorf("magpie search doesn't say lite:\n%s", out)
	}
	if err := searchCmd([]string{"add", "exa", "e-2"}); err != nil {
		t.Fatal(err)
	}
	if a := exa(); a.Type != "lite" || a.Key != "e-2" {
		t.Fatalf("a new key took the type away: %+v", a)
	}
	for _, back := range []string{"type=", "type=auto"} {
		if err := searchCmd([]string{"add", "exa", "type=lite"}); err != nil {
			t.Fatal(err)
		}
		if err := searchCmd([]string{"add", "exa", back}); err != nil {
			t.Fatal(err)
		}
		if a := exa(); a.Type != "" || a.Key != "e-2" {
			t.Fatalf("%s: %+v", back, a)
		}
	}
	if out := searchListOut(t); strings.Contains(out, "lite") {
		t.Errorf("magpie search says lite after auto:\n%s", out)
	}
}
