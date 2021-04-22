package token

import (
	"strconv"
	"strings"
)

// Token is a lexical token of the NervQL.
type Token int

// Token types.
const (
	// ILLEGAL Token, EOF, WS are Special tokens.
	ILLEGAL Token = iota
	EOF
	WS
	COMMENT

	literalBeg
	// IDENT and the following are   literal tokens.
	IDENT       // main
	QIDENT      //QIDENT
	BOUNDPARAM  // $param
	FLOAT       // 12345.67
	INTEGER     // 12345
	DURATIONVAL // 13h
	STRING      // "abc"
	//BADSTRING   // "abc
	//BADESCAPE // \q

	//REGEX // Regular expressions
	//BADREGEX    // `.*
	literalEnd

	operatorBeg
	// ADD and the following are Operators
	ADD // +
	SUB // -
	MUL // *
	DIV // /
	MOD // %
	//BITWISE_AND // &
	//BITWISE_OR  // |
	//BITWISE_XOR // ^

	EQ       // =
	NEQ      // !=
	EQREGEX  // =~
	NEQREGEX // !~
	LT       // <
	LTE      // <=
	GT       // >
	GTE      // >=
	operatorEnd

	LPAREN   // (
	RPAREN   // )
	COMMA    // ,
	QUESTION //?
	//COLON  // :
	//DOUBLECOLON // ::
	//SEMICOLON // ;
	//DOT // .

	keywordBeg
	// ALL and the following are Keywords
	BINARY
	TRUE  // true
	FALSE // false
	AND   // AND
	OR    // OR
	NOT   // NOT
	ALL
	ALTER
	ANALYZE
	ANY
	AS
	ASC
	BEGIN
	BY
	CARDINALITY
	CREATE
	CONTINUOUS
	DATABASE
	DATABASES
	DEFAULT
	DELETE
	DESC
	DESTINATIONS
	DIAGNOSTICS
	DISTINCT
	DROP
	DURATION
	END
	EVERY
	EXPLAIN
	FIELD
	FOR
	FROM
	GRANT
	GRANTS
	GROUP
	GROUPS
	IN
	INF
	INSERT
	INTO
	KEY
	KEYS
	KILL
	LIMIT
	TABLE
	TABLES
	//NAME
	OFFSET
	ON
	ORDER
	PASSWORD
	POLICY
	POLICIES
	PRIVILEGES
	QUERIES
	QUERY
	READ
	REPLICATION
	RESAMPLE
	RETENTION
	REVOKE
	SELECT
	SERIES
	SET
	SHOW
	SHARD
	SHARDS
	SLIMIT
	SOFFSET
	STATS
	SUBSCRIPTION
	SUBSCRIPTIONS
	TAG
	TO
	USER
	USERS
	VALUES
	WHERE
	WITH
	WRITE
	LIKE
	PRECISE
	keywordEnd
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",

	IDENT:       "IDENT",
	QIDENT:      "QIDENT",
	FLOAT:       "FLOAT",
	INTEGER:     "INTEGER",
	DURATIONVAL: "DURATIONVAL",
	STRING:      "STRING",
	//BADSTRING:   "BADSTRING",
	//BADESCAPE: "BADESCAPE",

	//REGEX: "REGEX",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	MOD: "%",
	//BITWISE_AND: "&",
	//BITWISE_OR:  "|",
	//BITWISE_XOR: "^",

	EQ:       "=",
	NEQ:      "<>", // or "!=""
	EQREGEX:  "=~",
	NEQREGEX: "!~",
	LT:       "<",
	LTE:      "<=",
	GT:       ">",
	GTE:      ">=",

	LPAREN:   "(",
	RPAREN:   ")",
	COMMA:    ",",
	QUESTION: "?",
	//COLON:       ":",
	//DOUBLECOLON: "::",
	//SEMICOLON: ";",
	//DOT: ".",
	BINARY:       "BINARY",
	TRUE:         "TRUE",
	FALSE:        "FALSE",
	AND:          "AND",
	OR:           "OR",
	NOT:          "NOT",
	ALL:          "ALL",
	ALTER:        "ALTER",
	ANALYZE:      "ANALYZE",
	ANY:          "ANY",
	AS:           "AS",
	ASC:          "ASC",
	BEGIN:        "BEGIN",
	BY:           "BY",
	CARDINALITY:  "CARDINALITY",
	CREATE:       "CREATE",
	CONTINUOUS:   "CONTINUOUS",
	DATABASE:     "DATABASE",
	DATABASES:    "DATABASES",
	DEFAULT:      "DEFAULT",
	DELETE:       "DELETE",
	DESC:         "DESC",
	DESTINATIONS: "DESTINATIONS",
	DIAGNOSTICS:  "DIAGNOSTICS",
	DISTINCT:     "DISTINCT",
	DROP:         "DROP",
	DURATION:     "DURATION",
	END:          "END",
	EVERY:        "EVERY",
	EXPLAIN:      "EXPLAIN",
	//FIELD:         "FIELD",
	FOR:    "FOR",
	FROM:   "FROM",
	GRANT:  "GRANT",
	GRANTS: "GRANTS",
	GROUP:  "GROUP",
	GROUPS: "GROUPS",
	IN:     "IN",
	INF:    "INF",
	INSERT: "INSERT",
	INTO:   "INTO",
	KEY:    "KEY",
	KEYS:   "KEYS",
	KILL:   "KILL",
	LIMIT:  "LIMIT",
	TABLE:  "TABLE",
	TABLES: "TABLES",
	//NAME:          "NAME",
	OFFSET:        "OFFSET",
	ON:            "ON",
	ORDER:         "ORDER",
	PASSWORD:      "PASSWORD",
	POLICY:        "POLICY",
	POLICIES:      "POLICIES",
	PRIVILEGES:    "PRIVILEGES",
	QUERIES:       "QUERIES",
	QUERY:         "QUERY",
	READ:          "READ",
	REPLICATION:   "REPLICATION",
	RESAMPLE:      "RESAMPLE",
	RETENTION:     "RETENTION",
	REVOKE:        "REVOKE",
	SELECT:        "SELECT",
	SERIES:        "SERIES",
	SET:           "SET",
	SHOW:          "SHOW",
	SHARD:         "SHARD",
	SHARDS:        "SHARDS",
	SLIMIT:        "SLIMIT",
	SOFFSET:       "SOFFSET",
	STATS:         "STATS",
	SUBSCRIPTION:  "SUBSCRIPTION",
	SUBSCRIPTIONS: "SUBSCRIPTIONS",
	TAG:           "TAG",
	TO:            "TO",
	//USER:          "USER",
	USERS:   "USERS",
	VALUES:  "VALUES",
	WHERE:   "WHERE",
	WITH:    "WITH",
	WRITE:   "WRITE",
	LIKE:    "LIKE",
	PRECISE: "PRECISE",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
func (t Token) String() string {
	s := ""
	if 0 <= t && t < Token(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

// IsOperator
func (t Token) IsOperator() bool {
	b := t >= operatorBeg && t <= operatorEnd || t == AND || t == OR || t == NOT || t == LIKE || t == IN
	return b
}

// Precedence returns the operator precedence of the binary operator token.
func (t Token) Precedence() int {
	switch t {
	case OR:
		return 1
	case AND:
		return 2
	case EQ, NEQ, LT, LTE, GT, GTE, EQREGEX, NEQREGEX, LIKE, IN:
		return 3
	case ADD, SUB:
		return 4
	case MUL, DIV, MOD:
		return 5
	}
	return 0
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for t := keywordBeg + 1; t < keywordEnd; t++ {
		keywords[strings.ToLower(tokens[t])] = t
	}
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword)./
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[strings.ToLower(ident)]; isKeyword {
		return tok
	}
	return IDENT
}
