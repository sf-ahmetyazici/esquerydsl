package esquerydsl

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestWrappedQuery(t *testing.T) {
	item := NestedQueryItem{
		Not: []QueryItem{
			{
				Field: "field",
				Type:  Exists,
				Value: "value",
			},
		},
	}

	body, err := json.Marshal(GetWrappedQuery(item))
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"bool":{"must_not":[{"exists":{"field":"value"}}]}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestBogusQueryType(t *testing.T) {
	_, err := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{{"id": "asc"}},
		And: []QueryItem{
			{
				Field: "some_index_id",
				Value: "some-long-key-id-value",
				Type:  100001,
			},
		},
	})

	var queryTypeErr *QueryTypeErr
	if !errors.As(err, &queryTypeErr) {
		t.Errorf("\nUnexpected error: %v", err)
	}
}

func TestQueryStringEsc(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Field: "user.id",
				Value: "kimchy!",
				Type:  QueryString,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"query_string":{"analyze_wildcard":true,"fields":["user.id"],"query":"kimchy\\!"}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestMultiSearchDoc(t *testing.T) {
	doc, _ := MultiSearchDoc([]QueryDoc{
		{
			Index: "index1",
			And: []QueryItem{
				{
					Field: "user.id",
					Value: "kimchy!",
					Type:  QueryString,
				},
			},
		},
		{
			Index: "index2",
			And: []QueryItem{
				{
					Field: "some_index_id",
					Value: "some-long-key-id-value",
					Type:  Match,
				},
			},
		},
	})

	expected := `{"index":"index1"}
{"query":{"bool":{"must":[{"query_string":{"analyze_wildcard":true,"fields":["user.id"],"query":"kimchy\\!"}}]}}}
{"index":"index2"}
{"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}}}
`
	if string(doc) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(doc))
	}
}

func TestAndQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{{"id": "asc"}},
		And: []QueryItem{
			{
				Field: "some_index_id",
				Value: "some-long-key-id-value",
				Type:  Match,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}},"sort":[{"id":"asc"}]}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestFilterQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Field: "title",
				Value: "Search",
				Type:  Match,
			},
			{
				Field: "content",
				Value: "Elasticsearch",
				Type:  Match,
			},
		},
		Filter: []QueryItem{
			{
				Field: "status",
				Value: "published",
				Type:  Term,
			},
			{
				Field: "publish_date",
				Value: map[string]string{
					"gte": "2015-01-01",
				},
				Type: Range,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"match":{"title":"Search"}},{"match":{"content":"Elasticsearch"}}],"filter":[{"term":{"status":"published"}},{"range":{"publish_date":{"gte":"2015-01-01"}}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestNestedQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Field: "nested_path",
				Value: NestedQueryItem{
					Filter: []QueryItem{WrapQueryItems("filter", QueryItem{
						Field: "id",
						Value: []string{"b4ab2c6e-93e3-40b9-8e66-9379f864186f"},
						Type:  Terms,
					})},
				},
				Type: NestedQuery,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"nested":{"path":["nested_path"],"query":{"bool":{"filter":[{"bool":{"filter":[{"terms":{"id":["b4ab2c6e-93e3-40b9-8e66-9379f864186f"]}}]}}]}}}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}
