package lexer_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"selfText/giligili_back/libcommon/ql/lexer"
	"selfText/giligili_back/libcommon/ql/token"
)

func TestLexer_Next(t *testing.T) {
	var tests = []struct {
		s   string
		tok token.Token
		lit string
		pos token.Pos
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, tok: token.EOF},
		//{s: `#`, tok: token.ILLEGAL, lit: `#`},
		// {s: ` `, tok: token.WS, lit: " "},
		// {s: "\t", tok: token.WS, lit: "\t"},
		// {s: "\n", tok: token.WS, lit: "\n"},
		// {s: "\r", tok: token.WS, lit: "\n"},
		// {s: "\r\n", tok: token.WS, lit: "\n"},
		// {s: "\rX", tok: token.WS, lit: "\n"},
		// {s: "\n\r", tok: token.WS, lit: "\n\n"},
		// {s: " \n\t \r\n\t", tok: token.WS, lit: " \n\t \n\t"},
		// {s: " foo", tok: token.WS, lit: " "},

		// Numeric operators
		{s: `+`, tok: token.ADD},
		{s: `-`, tok: token.SUB},
		{s: `*`, tok: token.MUL},
		{s: `/`, tok: token.DIV},
		{s: `%`, tok: token.MOD},

		// Logical operators
		{s: `AND`, tok: token.AND},
		{s: `and`, tok: token.AND},
		{s: `OR`, tok: token.OR},
		{s: `or`, tok: token.OR},
		{s: `NOT`, tok: token.NOT},
		{s: `not`, tok: token.NOT},

		{s: `=`, tok: token.EQ},
		{s: `!=`, tok: token.NEQ},
		{s: `! `, tok: token.ILLEGAL, lit: "!"},
		{s: `<`, tok: token.LT},
		{s: `<=`, tok: token.LTE},
		{s: `>`, tok: token.GT},
		{s: `>=`, tok: token.GTE},

		// Misc tokens
		{s: `(`, tok: token.LPAREN},
		{s: `)`, tok: token.RPAREN},
		{s: `,`, tok: token.COMMA},
		//{s: `;`, tok: token.SEMICOLON},
		//{s: `.`, tok: token.DOT},
		{s: `=~`, tok: token.EQREGEX},
		{s: `!~`, tok: token.NEQREGEX},
		//{s: `:`, tok: token.COLON},
		//{s: `::`, tok: token.DOUBLECOLON},

		// Identifiers
		{s: `Aborted_Connects`, tok: token.IDENT, lit: `Aborted_Connects`},
		{s: `FoOa`, tok: token.IDENT, lit: `FoOa`},
		//{s: `foo123`, tok: token.IDENT, lit: `foo123`},
		// {s: `_foo`, tok: token.IDENT, lit: `_foo`},
		// {s: `Zx12_3U_-`, tok: token.IDENT, lit: `Zx12_3U_`},
		// {s: `"foo"`, tok: token.IDENT, lit: `foo`},
		// {s: `"foo\\bar"`, tok: token.IDENT, lit: `foo\bar`},
		//{s: `"foo\bar"`, tok: token.BADESCAPE, lit: `\b`, pos: token.Pos{Line: 0, Char: 5}},
		//{s: `"foo\"bar\""`, tok: token.IDENT, lit: `foo"bar"`},
		//{s: `test"`, tok: token.BADSTRING, lit: "", pos: token.Pos{Line: 0, Char: 3}},
		//{s: `"test`, tok: token.BADSTRING, lit: `test`},
		// {s: `$host`, tok: token.BOUNDPARAM, lit: `$host`},
		// {s: `$"host param"`, tok: token.BOUNDPARAM, lit: `$host param`},

		{s: `true`, tok: token.TRUE},
		{s: `false`, tok: token.FALSE},

		// Strings
		{s: `'testing 123!'`, tok: token.STRING, lit: `'testing 123!'`},
		//{s: `'foo\nbar'`, tok: token.STRING, lit: "foo\nbar"},
		//{s: `'foo\\bar'`, tok: token.STRING, lit: "foo\\bar"},
		//{s: `'test`, tok: token.BADSTRING, lit: `test`},
		//{s: "'test\nfoo", tok: token.BADSTRING, lit: `test`},
		//{s: `'test\g'`, tok: token.BADESCAPE, lit: `\g`, pos: token.Pos{Line: 0, Char: 6}},

		// Numbers
		{s: `100`, tok: token.INTEGER, lit: `100`},
		//{s: `-100`, tok: token.INTEGER, lit: `-100`},
		{s: `100.23`, tok: token.FLOAT, lit: `100.23`},
		//{s: `-100.23`, tok: token.FLOAT, lit: `-100.23`},
		//{s: `.23`, tok: token.NUMBER, lit: `.23`},
		{s: `.`, tok: token.ILLEGAL, lit: `.`},
		//{s: `10.3s`, tok: token.NUMBER, lit: `10.3`},

		// Durations
		// {s: `10u`, tok: token.DURATIONVAL, lit: `10u`},
		// {s: `10µ`, tok: token.DURATIONVAL, lit: `10µ`},
		// {s: `10ms`, tok: token.DURATIONVAL, lit: `10ms`},
		// {s: `1s`, tok: token.DURATIONVAL, lit: `1s`},
		// {s: `10m`, tok: token.DURATIONVAL, lit: `10m`},
		// {s: `10h`, tok: token.DURATIONVAL, lit: `10h`},
		// {s: `10d`, tok: token.DURATIONVAL, lit: `10d`},
		// {s: `10w`, tok: token.DURATIONVAL, lit: `10w`},
		// {s: `10x`, tok: token.DURATIONVAL, lit: `10x`}, // non-duration unit, but scanned as a duration value

		// Keywords
		{s: `ALL`, tok: token.ALL},
		{s: `ALTER`, tok: token.ALTER},
		{s: `AS`, tok: token.AS},
		{s: `ASC`, tok: token.ASC},
		{s: `BEGIN`, tok: token.BEGIN},
		{s: `BY`, tok: token.BY},
		{s: `CREATE`, tok: token.CREATE},
		{s: `CONTINUOUS`, tok: token.CONTINUOUS},
		{s: `DATABASE`, tok: token.DATABASE},
		{s: `DATABASES`, tok: token.DATABASES},
		{s: `DEFAULT`, tok: token.DEFAULT},
		{s: `DELETE`, tok: token.DELETE},
		{s: `DESC`, tok: token.DESC},
		{s: `DROP`, tok: token.DROP},
		{s: `DURATION`, tok: token.DURATION},
		{s: `END`, tok: token.END},
		{s: `EVERY`, tok: token.EVERY},
		{s: `EXPLAIN`, tok: token.EXPLAIN},
		{s: `FIELD`, tok: token.FIELD},
		{s: `FROM`, tok: token.FROM},
		{s: `GRANT`, tok: token.GRANT},
		{s: `GROUP`, tok: token.GROUP},
		{s: `GROUPS`, tok: token.GROUPS},
		{s: `INSERT`, tok: token.INSERT},
		{s: `INTO`, tok: token.INTO},
		{s: `KEY`, tok: token.KEY},
		{s: `KEYS`, tok: token.KEYS},
		{s: `KILL`, tok: token.KILL},
		{s: `LIMIT`, tok: token.LIMIT},
		{s: `SHOW`, tok: token.SHOW},
		{s: `SHARD`, tok: token.SHARD},
		{s: `SHARDS`, tok: token.SHARDS},
		{s: `TABLE`, tok: token.TABLE},
		{s: `TABLES`, tok: token.TABLES},
		{s: `OFFSET`, tok: token.OFFSET},
		{s: `ON`, tok: token.ON},
		{s: `ORDER`, tok: token.ORDER},
		{s: `PASSWORD`, tok: token.PASSWORD},
		{s: `POLICY`, tok: token.POLICY},
		{s: `POLICIES`, tok: token.POLICIES},
		{s: `PRIVILEGES`, tok: token.PRIVILEGES},
		{s: `QUERIES`, tok: token.QUERIES},
		{s: `QUERY`, tok: token.QUERY},
		{s: `READ`, tok: token.READ},
		{s: `REPLICATION`, tok: token.REPLICATION},
		{s: `RESAMPLE`, tok: token.RESAMPLE},
		{s: `RETENTION`, tok: token.RETENTION},
		{s: `REVOKE`, tok: token.REVOKE},
		{s: `SELECT`, tok: token.SELECT},
		{s: `SERIES`, tok: token.SERIES},
		{s: `TAG`, tok: token.TAG},
		{s: `TO`, tok: token.TO},
		{s: `USER`, tok: token.USER},
		{s: `USERS`, tok: token.USERS},
		{s: `VALUES`, tok: token.VALUES},
		{s: `WHERE`, tok: token.WHERE},
		{s: `WITH`, tok: token.WITH},
		{s: `WRITE`, tok: token.WRITE},
		{s: `explain`, tok: token.EXPLAIN}, // case insensitive
		{s: `seLECT`, tok: token.SELECT},   // case insensitive
		{s: `DISTINCT`, tok: token.DISTINCT},
		{s: `disTINCT`, tok: token.DISTINCT},
		{s: `disTInCT`, tok: token.DISTINCT},
		{s: `distinct`, tok: token.DISTINCT},
	}

	for i, tc := range tests {
		l := lexer.NewLexer(strings.NewReader(tc.s))
		tok, lit, pos := l.Next()
		if tc.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tc.s, tc.tok, tok, lit)
		} else if tc.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tc.s, tc.lit, lit)
		} else if tc.pos.Line != pos.Line || tc.pos.Char != pos.Char {
			t.Errorf("%d. %q pos mismatch: exp=%#v got=%#v", i, tc.s, tc.pos, pos)
		}
	}
}

func TestLexer_String(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err string
	}{
		{in: `SELECT usr,sys FROM cpu,mem where total<>''`, out: `SELECTusr,sysFROMcpu,memWHEREtotal<>''`},
		//{in: `SELECT value FROM cpu`, out: `SELECTvalueFROMcpu`},
		//{in: "SELECT 用户,sys FROM cpu", out: `SELECT用户,sysFROMcpu`},
		{in: `SELECT usr,sys FROM cpu`, out: `SELECTusr,sysFROMcpu`},
		{in: `SELECT usr,sys FROM cpu,mem`, out: `SELECTusr,sysFROMcpu,mem`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1`, out: `SELECTusr,sysFROMcpu,memWHEREtotal=1`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1 and sys=10`, out: `SELECTusr,sysFROMcpu,memWHEREtotal=1ANDsys=10`},
		{in: `SELECT usr,sys FROM cpu,mem where (total=1 and sys=10) or idle=1.0`, out: `SELECTusr,sysFROMcpu,memWHERE(total=1ANDsys=10)ORidle=1.0`},

		{in: `SELECT usr,sys FROM cpu,mem where 1=total`, out: `SELECTusr,sysFROMcpu,memWHERE1=total`},
		{in: `SELECT usr,sys FROM cpu,mem where total='1'`, out: `SELECTusr,sysFROMcpu,memWHEREtotal='1'`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'=total`, out: `SELECTusr,sysFROMcpu,memWHERE'1'=total`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'<>total`, out: `SELECTusr,sysFROMcpu,memWHERE'1'<>total`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'!~total`, out: `SELECTusr,sysFROMcpu,memWHERE'1'!~total`},
		{in: `SELECT usr,sys FROM cpu,mem where total=~'1$'`, out: `SELECTusr,sysFROMcpu,memWHEREtotal=~'1$'`},
		{in: `SELECT distinct usr,distinct sys FROM cpu,mem where total=~'1$'`, out: `SELECTDISTINCTusr,DISTINCTsysFROMcpu,memWHEREtotal=~'1$'`},
	}

	for i, tc := range tests {
		l := lexer.NewLexer(strings.NewReader(tc.in))
		var buf bytes.Buffer
		tok, lit, _ := l.Next()
		for tok != token.EOF {
			if lit != "" {
				fmt.Print(lit)
				buf.WriteString(lit)
			} else {
				fmt.Print(tok.String())
				buf.WriteString(tok.String())
			}
			tok, lit, _ = l.Next()
		}
		got := buf.String()
		fmt.Println(got)
		if tc.out != buf.String() {
			t.Errorf("%d. %s: exp=%s, got=%s", i, tc.in, tc.out, got)
		}
		fmt.Println("")
	}
}
