package test

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

const schemaRelative = "/shared/schema.gql"

func assertQuery(t *testing.T, r *http.Request) {
	t.Helper()

	schema := loadSchema(t)
	query, err := readQuery(r)
	require.NoError(t, err)
	doc := parseQuery(t, string(query.query))
	query.updateDoc(t, &doc)
	validateQuery(t, schema, doc)
}

func loadSchema(t *testing.T) *ast.Schema {
	t.Helper()

	schemaFilePath := findSchemaFilePath(t)

	content, err := os.ReadFile(schemaFilePath)
	require.NoError(t, err)

	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "schema.gql",
		Input: string(content),
	})
	require.NoError(t, err)

	return schema
}

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

	return dots + schemaRelative
}

type multipartUploadMap = map[string][]string

type unparsedGraphqlQuery struct {
	query         []byte
	variables     map[string]any
	fileVariables multipartUploadMap
	fileContents  map[string][]byte
}

func (q unparsedGraphqlQuery) updateDoc(t *testing.T, doc *parsedGraphQLQuery) {
	t.Helper()

	for filePartName, variablePath := range q.fileVariables {
		require.Len(t, variablePath, 1)
		pathParts := strings.Split(variablePath[0], ".")
		varIdx := slices.Index(pathParts, "variables")
		require.Equal(t, 0, varIdx)
		// assume no reference to nested file variables
		doc.Variables[pathParts[1]] = q.fileContents[filePartName]
	}
}

type parsedGraphQLQuery struct {
	OperationName string
	Query         string
	Variables     map[string]any
	doc           *ast.QueryDocument
}

func readQuery(r *http.Request) (unparsedGraphqlQuery, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart") {
		if m, err := readMultipartUpload(r); err != nil {
			return unparsedGraphqlQuery{}, err
		} else {
			return m, nil
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return unparsedGraphqlQuery{}, err
	}

	return unparsedGraphqlQuery{query: body}, nil
}

func readMultipartUpload(r *http.Request) (unparsedGraphqlQuery, error) {
	var operations []byte
	variables := make(map[string]any)
	var fileVariables multipartUploadMap
	fileContents := make(map[string][]byte)

	contentType := r.Header.Get("Content-Type")
	boundarySplit := strings.Split(contentType, "boundary=")
	boundary := boundarySplit[1]
	reader := multipart.NewReader(r.Body, boundary)

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return unparsedGraphqlQuery{}, err
		}

		name := part.FormName()
		switch name {
		case "operations":
			if operations, err = io.ReadAll(part); err != nil {
				return unparsedGraphqlQuery{}, err
			}
		case "map":
			data, err := io.ReadAll(part)
			if err != nil {
				return unparsedGraphqlQuery{}, err
			}

			if err := json.Unmarshal(data, &fileVariables); err != nil {
				return unparsedGraphqlQuery{}, err
			}
		default:
			if data, err := io.ReadAll(part); err != nil {
				return unparsedGraphqlQuery{}, err
			} else {
				fileContents[name] = data
			}
		}
	}

	result := unparsedGraphqlQuery{
		query:         operations,
		variables:     variables,
		fileVariables: fileVariables,
		fileContents:  fileContents,
	}

	return result, nil
}

func validateQuery(t *testing.T, schema *ast.Schema, query parsedGraphQLQuery) {
	t.Helper()

	require.NotEqual(t, []ast.Operation{}, query.doc.Operations)
	listErr := validator.Validate(schema, query.doc)
	require.Equal(t, []error{}, listErr.Unwrap(), "invalid query")

	_, err := validator.VariableValues(schema, query.doc.Operations.ForName(query.OperationName), query.Variables)
	assert.NoError(t, err, "invalid variable values")
}

func parseQuery(t *testing.T, query string) parsedGraphQLQuery {
	t.Helper()

	var parsedQuery parsedGraphQLQuery

	err := json.NewDecoder(strings.NewReader(query)).Decode(&parsedQuery)
	require.NoError(t, err, "error decoding query")

	doc, err := parser.ParseQuery(&ast.Source{Input: parsedQuery.Query})
	if err != nil {
		_, ok := err.(*gqlerror.Error)
		require.False(t, ok, "error parsing query")
	}

	parsedQuery.doc = doc

	return parsedQuery
}
