package golembic

import (
	"strings"
)

// QuoteIdentifier quotes an identifier, such as a table name, for usage
// in a query.
//
// This implementation is vendored in here to avoid the side effects of
// importing `github.com/lib/pq`.
//
// See:
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L1564-L1581
// - https://www.sqlite.org/lang_keywords.html
// - https://github.com/ronsavage/SQL/blob/a67e7eaefae89ed761fa4dcbc5431ec9a235a6c8/sql-99.bnf#L412
func QuoteIdentifier(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}

// QuoteLiteral quotes a literal, such as `2023-01-05 15:00:00Z`, for usage
// in a query.
//
// This implementation is vendored in here to avoid the side effects of
// importing `github.com/lib/pq`.
//
// See:
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L1583-L1614
// - https://www.sqlite.org/lang_keywords.html
// - https://github.com/ronsavage/SQL/blob/a67e7eaefae89ed761fa4dcbc5431ec9a235a6c8/sql-99.bnf#L758-L761
// - https://github.com/ronsavage/SQL/blob/a67e7eaefae89ed761fa4dcbc5431ec9a235a6c8/sql-99.bnf#L290
func QuoteLiteral(literal string) string {
	literal = strings.Replace(literal, `'`, `''`, -1)
	if strings.Contains(literal, `\`) {
		literal = strings.Replace(literal, `\`, `\\`, -1)
		literal = ` E'` + literal + `'`
	} else {
		literal = `'` + literal + `'`
	}
	return literal
}
