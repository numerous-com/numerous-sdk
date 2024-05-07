package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"

	_ "embed"
)

func CreateTestGqlClient(t *testing.T, response string) *gqlclient.Client {
	t.Helper()

	schema := loadSchema(t)
	h := http.Header{}
	h.Add("Content-Type", "application/json")

	ts := TestTransport{
		WithResponse: &http.Response{
			Header:     h,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		},
		Handler: func(r *http.Request) *struct {
			Response *http.Response
			Error    error
		} {
			query, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			doc := parseQuery(t, string(query))
			validateQuery(t, schema, doc)

			return nil
		},
	}

	return gqlclient.New("http://localhost:8080", &http.Client{Transport: &ts})
}

func CreateMockGqlClient(responses ...string) (*gqlclient.Client, *MockTransport) {
	ts := MockTransport{}

	for _, response := range responses {
		AddResponseToMockGqlClient(response, &ts)
	}

	return gqlclient.New("http://localhost:8080", &http.Client{Transport: &ts}), &ts
}

func AddResponseToMockGqlClient(response string, ts *MockTransport) {
	h := http.Header{}
	h.Add("Content-Type", "application/json")

	ts.On("RoundTrip", mock.Anything).Once().Return(
		&http.Response{
			Header:     h,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(response))),
		},
		nil,
	)
}

func loadSchema(t *testing.T) *ast.Schema {
	t.Helper()

	dots := findSchemaFilePath(t)

	content, err := os.ReadFile(dots + "shared/schema.gql")
	require.NoError(t, err)

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "schema.gql",
		Input: string(content),
	})
	require.NoError(t, err)

	return schema
}

var schemaRelative = "/shared/schema.gql"

func findSchemaFilePath(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	dots := ""
	for {
		if _, err := os.Stat(wd + schemaRelative); err == nil {
			break
		}

		dir, _ := filepath.Split(wd)
		wd = filepath.Clean(dir)
		require.NotEmpty(t, wd)
		require.NotEqual(t, "/", wd)

		dots += "../"
	}

	return dots
}

func validateQuery(t *testing.T, schema *ast.Schema, doc *ast.QueryDocument) {
	t.Helper()

	require.NotEqual(t, []ast.Operation{}, doc.Operations)
	listErr := validator.Validate(schema, doc)
	require.Equal(t, []error{}, listErr.Unwrap())
}

func parseQuery(t *testing.T, query string) *ast.QueryDocument {
	t.Helper()

	var queryObj struct {
		Query     string
		Variables map[string]any
	}

	err := json.NewDecoder(strings.NewReader(query)).Decode(&queryObj)
	require.NoError(t, err, "error decoding query")

	doc, err := parser.ParseQuery(&ast.Source{Input: queryObj.Query})
	if err != nil {
		_, ok := err.(*gqlerror.Error)
		require.False(t, ok, "error parsing query")
	}

	return doc
}
