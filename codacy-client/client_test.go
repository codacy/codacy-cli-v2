package codacyclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codacy/cli-v2/domain"

	"github.com/stretchr/testify/assert"
)

func TestGetRequest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "ok"}`))
	}))
	defer ts.Close()

	initFlags := domain.InitFlags{ApiToken: "dummy"}
	resp, err := getRequest(ts.URL, initFlags)
	assert.NoError(t, err)
	assert.Contains(t, string(resp), "ok")
}

func TestGetRequest_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	initFlags := domain.InitFlags{ApiToken: "dummy"}
	_, err := getRequest(ts.URL, initFlags)
	assert.Error(t, err)
}

func TestHandlePaginationGeneric(t *testing.T) {
	type testItem struct{ Value int }
	pages := [][]testItem{
		{{Value: 1}, {Value: 2}},
		{{Value: 3}},
	}
	calls := 0
	fetchPage := func(url string) ([]testItem, string, error) {
		if calls < len(pages) {
			page := pages[calls]
			calls++
			if calls < len(pages) {
				return page, "next", nil
			}
			return page, "", nil
		}
		return nil, "", nil
	}
	results, err := handlePaginationGeneric[testItem]("base", "", fetchPage)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestGetDefaultToolPatternsConfig_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data":       []interface{}{},
			"pagination": map[string]interface{}{"cursor": ""},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// TODO: Refactor GetDefaultToolPatternsConfig to accept a baseURL for easier testing
	// oldBase := CodacyApiBase
	// CodacyApiBase = ts.URL
	// defer func() { CodacyApiBase = oldBase }()

	// Placeholder: test cannot be run until function is refactored for testability
	_ = ts // avoid unused warning
	// initFlags := domain.InitFlags{ApiToken: "dummy"}
	// patterns, err := GetDefaultToolPatternsConfig(initFlags, "tool-uuid")
	// assert.NoError(t, err)
	// assert.Empty(t, patterns)
}
