package codacyclient

import (
	"encoding/json"
	"fmt"
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
	resp, err := getRequest(ts.URL, initFlags.ApiToken)
	assert.NoError(t, err)
	assert.Contains(t, string(resp), "ok")
}

func TestGetRequest_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	initFlags := domain.InitFlags{ApiToken: "dummy"}
	_, err := getRequest(ts.URL, initFlags.ApiToken)
	assert.Error(t, err)
}

func TestGetPageAndGetAllPages(t *testing.T) {
	type testItem struct{ Value int }
	serverPages := []struct {
		data   []testItem
		cursor string
	}{
		{[]testItem{{Value: 1}, {Value: 2}}, "next"},
		{[]testItem{{Value: 3}}, ""},
	}
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"data":       serverPages[calls].data,
			"pagination": map[string]interface{}{"cursor": serverPages[calls].cursor},
		}
		calls++
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	initFlags := domain.InitFlags{ApiToken: "dummy"}

	parser := func(body []byte) ([]testItem, string, error) {
		var objmap map[string]json.RawMessage
		if err := json.Unmarshal(body, &objmap); err != nil {
			return nil, "", err
		}
		var items []testItem
		if err := json.Unmarshal(objmap["data"], &items); err != nil {
			return nil, "", err
		}
		var pagination struct {
			Cursor string `json:"cursor"`
		}
		if objmap["pagination"] != nil {
			_ = json.Unmarshal(objmap["pagination"], &pagination)
		}
		return items, pagination.Cursor, nil
	}

	// Test GetPage
	calls = 0
	items, cursor, err := GetPage[testItem](ts.URL, initFlags, parser)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "next", cursor)

	// Test getAllPages
	calls = 0
	allItems, err := getAllPages[testItem](ts.URL, initFlags, parser)
	assert.NoError(t, err)
	assert.Len(t, allItems, 3)
}

func TestGetToolPatternsConfig_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data":       []interface{}{},
			"pagination": map[string]interface{}{"cursor": ""},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	CodacyApiBase = ts.URL

	initFlags := domain.InitFlags{ApiToken: "dummy"}
	patterns, err := GetToolPatternsConfigWithCodacyAPIBase(CodacyApiBase, initFlags, "tool-uuid", true)
	assert.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestGetToolPatternsConfig_WithNonRecommended(t *testing.T) {

	config := []domain.PatternDefinition{
		{
			Id:      "internal_id_1",
			Enabled: true,
		},
		{
			Id:      "internal_id_2",
			Enabled: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resp := map[string]interface{}{
			"data":       config,
			"pagination": map[string]interface{}{"cursor": ""},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	expected := []domain.PatternConfiguration{
		{
			Enabled: true,
			PatternDefinition: domain.PatternDefinition{
				Id:      "internal_id_1",
				Enabled: true,
			},
		},
	}

	CodacyApiBase = ts.URL

	initFlags := domain.InitFlags{ApiToken: "dummy"}
	patterns, err := GetToolPatternsConfigWithCodacyAPIBase(CodacyApiBase, initFlags, "tool-uuid", true)

	fmt.Println(len(patterns))

	assert.NoError(t, err)
	assert.Equal(t, expected, patterns)
}
