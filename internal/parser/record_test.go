package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/desprit-media/traversql-core/internal/parser"
)

func TestWhere(t *testing.T) {
	t.Run("should create where clause for a primary key", func(t *testing.T) {
		testCases := []struct {
			name         string
			columns      []parser.Column
			values       []interface{}
			expectClause string
			expectArgs   []interface{}
		}{
			{
				name:         "Single column",
				columns:      []parser.Column{{Name: "id"}},
				values:       []interface{}{1},
				expectClause: "id = $1",
				expectArgs:   []interface{}{1},
			},
			{
				name:         "Two columns",
				columns:      []parser.Column{{Name: "first_name"}, {Name: "last_name"}},
				values:       []interface{}{"John", "Doe"},
				expectClause: "first_name = $1 AND last_name = $2",
				expectArgs:   []interface{}{"John", "Doe"},
			},
			{
				name:         "Mixed types",
				columns:      []parser.Column{{Name: "id"}, {Name: "active"}, {Name: "name"}},
				values:       []interface{}{42, true, "test"},
				expectClause: "id = $1 AND active = $2 AND name = $3",
				expectArgs:   []interface{}{42, true, "test"},
			},
			{
				name:         "Empty key",
				columns:      []parser.Column{},
				values:       []interface{}{},
				expectClause: "",
				expectArgs:   []interface{}{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				pk, err := parser.NewPrimaryKey(tc.columns, tc.values)
				assert.NoError(t, err)

				whereClause, args := pk.WhereClause()

				assert.Equal(t, tc.expectClause, whereClause)
				assert.Equal(t, tc.expectArgs, args)
			})
		}
	})
}
