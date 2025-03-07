package mql

import (
	"strings"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/uptrace/pkg/metrics/mql/ast"
)

func ParseQueryError(query string) (*ParsedQuery, error) {
	parsedQuery := ParseQuery(query)
	for _, part := range parsedQuery.Parts {
		if part.Error.Wrapped != nil {
			return nil, part.Error.Wrapped
		}
	}
	return parsedQuery, nil
}

func ParseQuery(query string) *ParsedQuery {
	parts := make([]*QueryPart, 0)

	for _, query := range SplitQuery(query) {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		part := &QueryPart{
			Query: query,
		}
		parts = append(parts, part)

		v, err := ast.Parse(query)
		if err != nil {
			part.Error.Wrapped = err
		} else {
			part.AST = v
		}
	}

	return &ParsedQuery{
		Parts: parts,
	}
}

type ParsedQuery struct {
	Parts   []*QueryPart  `json:"parts"`
	Columns []*ColumnInfo `json:"columns"`
}

type QueryPart struct {
	Query    string    `json:"query"`
	Error    JSONError `json:"error,omitempty"`
	Disabled bool      `json:"disabled,omitempty"`

	AST any `json:"-"`
}

type ColumnInfo struct{}

type JSONError struct {
	Wrapped error
}

func (e JSONError) MarshalJSON() ([]byte, error) {
	if e.Wrapped == nil {
		return []byte(`""`), nil
	}
	return json.Marshal(e.Wrapped.Error())
}

func SplitQuery(query string) []string {
	return strings.Split(query, " | ")
}

func JoinQuery(parts []string) string {
	return strings.Join(parts, " | ")
}
