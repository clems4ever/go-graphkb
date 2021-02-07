package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	Query string
	Error string
}

var testCases = []TestCase{
	{
		Query: "MATCH",
		Error: "Parsing errors detected: line 1:5 - no viable alternative at input 'MATCH'",
	},
	{
		Query: "MATCH (n) RETURN ",
		Error: "Parsing errors detected: line 1:17 - no viable alternative at input ' '",
	},
	{
		Query: "MATCH (n) RETURN c, n,",
		Error: "Parsing errors detected: line 1:22 - mismatched input '<EOF>' expecting {'(', '[', '+', '-', '{', '$', ALL, NOT, NULL, COUNT, ANY, NONE, SINGLE, TRUE, FALSE, EXISTS, CASE, StringLiteral, HexInteger, DecimalInteger, OctalInteger, HexLetter, ExponentDecimalReal, RegularDecimalReal, FILTER, EXTRACT, UnescapedSymbolicName, EscapedSymbolicName, SP}",
	},
	{
		Query: "MATCH (n)-[r]-> RETURN c",
		Error: "Parsing errors detected: line 1:16 - no viable alternative at input 'MATCH (n)-[r]-> RETURN'",
	},
	{
		Query: "MATCH (n)->[r]-(n) RETURN c",
		Error: "Parsing errors detected: line 1:10 - no viable alternative at input 'MATCH (n)->'",
	},
}

func TestQuery(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.Query, func(t *testing.T) {
			q, err := TransformCypher(tc.Query)

			require.Nil(t, q)
			require.EqualError(t, err, tc.Error)
		})
	}
}
