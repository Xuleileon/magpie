package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// Exa is asked with no type unless the user set one, so its default
// (auto) stays what a search spends; with type lite the body is the same
// besides "type": "lite", every other field one Lite takes
// (docs.exa.ai/reference/search), since Lite turns away a request with
// any other.
func TestExaSearchType(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var mu sync.Mutex
	var bodies []map[string]any
	exa := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" || r.Header.Get("X-Api-Key") != "exa-k" {
			http.Error(w, `{"error":"unexpected request"}`, 400)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		if err := json.Unmarshal(b, &body); err != nil {
			http.Error(w, `{"error":"not JSON"}`, 400)
			return
		}
		mu.Lock()
		bodies = append(bodies, body)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"requestId":"r1","results":[{"id":"https://go.dev/dl/","title":"Go downloads","url":"https://go.dev/dl/","text":"Go 1.27.1 was released."}],"costDollars":{"total":0.004}}`)
	}))
	defer exa.Close()
	s := New()
	liteTakes := []string{"query", "objective", "userIdentifier", "metadata", "category", "numResults", "contents",
		"userLocation", "includeDomains", "excludeDomains", "startPublishedDate", "endPublishedDate", "moderation", "compliance", "type"}
	for _, typ := range []string{"", "lite"} {
		pages, err := s.askSearchAPI(context.Background(), provider.SearchAPI{Vendor: "exa", Key: "exa-k", URL: exa.URL, Type: typ}, "latest go")
		if err != nil {
			t.Fatalf("type %q: %v", typ, err)
		}
		if len(pages) != 1 || pages[0].URL != "https://go.dev/dl/" || pages[0].Text != "Go 1.27.1 was released." {
			t.Fatalf("type %q: %+v", typ, pages)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 {
		t.Fatalf("Exa asked %d times", len(bodies))
	}
	auto, lite := bodies[0], bodies[1]
	if _, ok := auto["type"]; ok {
		t.Errorf("no type set, yet sent %v", auto["type"])
	}
	if lite["type"] != "lite" {
		t.Errorf("type lite sent as %v", lite["type"])
	}
	for k := range lite {
		if !slices.Contains(liteTakes, k) {
			t.Errorf("Lite turns away %q", k)
		}
	}
	delete(lite, "type")
	if !reflect.DeepEqual(auto, lite) {
		t.Errorf("the rest of the body changed with the type:\n auto %v\n lite %v", auto, lite)
	}
	if want := map[string]any{"text": map[string]any{"maxCharacters": float64(pageText)}}; !reflect.DeepEqual(auto["contents"], want) {
		t.Errorf("contents %v", auto["contents"])
	}
	if auto["query"] != "latest go" {
		t.Errorf("query %v", auto["query"])
	}
}
