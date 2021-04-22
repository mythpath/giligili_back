package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"selfText/giligili_back/libcommon/ql/lexer"
	"selfText/giligili_back/libcommon/ql/parser"
)

func TestSelect(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err string
	}{
		{in: "select * from `主机` where  id in (select target_id from `关系名` where source_id = 9) and target_id = 8 limit 10 offset 1  ", out: "SELECT * FROM `主机` WHERE (id IN (SELECT target_id FROM `关系名` WHERE (source_id = 9)) AND (target_id = 8)) LIMIT 10 OFFSET 1"},
		{in: "select * from `主机` where  id in (select target_id from `关系名` where source_id = 9) limit 10 offset 1  ", out: "SELECT * FROM `主机` WHERE id IN (SELECT target_id FROM `关系名` WHERE (source_id = 9)) LIMIT 10 OFFSET 1"},
		{in: "SELECT distinct * FROM cpu,mem where total >=  -1 limit 10 offset 1 ", out: "SELECT DISTINCT * FROM cpu, mem WHERE (total >= -1) LIMIT 10 OFFSET 1"},
		{in: "SELECT distinct `用户`,sys FROM cpu,mem where total >=  -1 ", out: "SELECT DISTINCT `用户`, sys FROM cpu, mem WHERE (total >= -1)"},
		{in: "SELECT `用户`,sys FROM cpu,mem where total >=  -1 ", out: "SELECT `用户`, sys FROM cpu, mem WHERE (total >= -1)"},
		{in: "SELECT `用户`,sys FROM cpu,mem where total >=  ? ", out: "SELECT `用户`, sys FROM cpu, mem WHERE (total >= ?)"},
		{in: `SELECT usr,sys FROM cpu,mem where total >=  ?`, out: `SELECT usr, sys FROM cpu, mem WHERE (total >= ?)`},
		{in: "SELECT * FROM Aborted_Connects group by `region`, idc, appid", out: "SELECT * FROM Aborted_Connects GROUP BY `region`, idc, appid"},
		{in: `SELECT * FROM Aborted_Connects`, out: `SELECT * FROM Aborted_Connects`},
		{in: `SELECT total FROM cpu`, out: `SELECT total FROM cpu`},
		{in: `SELECT usr,sys FROM cpu`, out: `SELECT usr, sys FROM cpu`},
		{in: `SELECT usr,sys FROM cpu,mem where not (total > 1)`, out: `SELECT usr, sys FROM cpu, mem WHERE (NOT (total > 1))`},
		{in: `SELECT usr,sys FROM cpu,mem where total >=   '1'`, out: `SELECT usr, sys FROM cpu, mem WHERE (total >= '1')`},
		{in: `SELECT usr,sys FROM cpu,mem where 1=idle`, out: `SELECT usr, sys FROM cpu, mem WHERE (1 = idle)`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'=idle`, out: `SELECT usr, sys FROM cpu, mem WHERE ('1' = idle)`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'='2'`, out: `SELECT usr, sys FROM cpu, mem WHERE ('1' = '2')`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'=2`, out: `SELECT usr, sys FROM cpu, mem WHERE ('1' = 2)`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'=~2`, out: `SELECT usr, sys FROM cpu, mem WHERE ('1' =~ 2)`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'!~2`, out: `SELECT usr, sys FROM cpu, mem WHERE ('1' !~ 2)`},
		{in: `SELECT usr,sys FROM cpu,mem where '1'=~2 and idc =~ '^bj' and jf !~ 'yz$' and a = 'b'`, out: `SELECT usr, sys FROM cpu, mem WHERE (((('1' =~ 2) AND (idc =~ '^bj')) AND (jf !~ 'yz$')) AND (a = 'b'))`},
		{in: `SELECT usr,sys FROM cpu,mem where 1 ='2'`, out: `SELECT usr, sys FROM cpu, mem WHERE (1 = '2')`},
		{in: `SELECT usr,sys FROM cpu,mem where 1= 2`, out: `SELECT usr, sys FROM cpu, mem WHERE (1 = 2)`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1 and idle=1.2`, out: `SELECT usr, sys FROM cpu, mem WHERE ((total = 1) AND (idle = 1.2))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1+2 and idle=1.2`, out: `SELECT usr, sys FROM cpu, mem WHERE ((total = (1 + 2)) AND (idle = 1.2))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1+2+3 and idle=1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((total = ((1 + 2) + 3)) AND (idle = 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1*2/4+9+3 and idle=1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((total = ((((1 * 2) / 4) + 9) + 3)) AND (idle = 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1*2+3 and idle>1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((total = ((1 * 2) + 3)) AND (idle > 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1*2+3/6 and idle>1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((total = ((1 * 2) + (3 / 6))) AND (idle > 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1*(2+3)/6 and idle>1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((total = ((1 * (2 + 3)) / 6)) AND (idle > 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where 1*(2+3)/6=total and idle>1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((((1 * (2 + 3)) / 6) = total) AND (idle > 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where 1*(2-3)/6=total and idle>1.2 or '5'=sys`, out: `SELECT usr, sys FROM cpu, mem WHERE (((((1 * (2 - 3)) / 6) = total) AND (idle > 1.2)) OR ('5' = sys))`},
		{in: `SELECT usr,sys FROM cpu,mem where -1= 2`, out: `SELECT usr, sys FROM cpu, mem WHERE (-1 = 2)`},
		{in: `SELECT usr,sys FROM cpu,mem where -1= 2-t`, out: `SELECT usr, sys FROM cpu, mem WHERE (-1 = (2 - t))`},
		{in: `SELECT usr,sys FROM cpu,mem where -1= -1.02-t`, out: `SELECT usr, sys FROM cpu, mem WHERE (-1 = (-1.02 - t))`},
		{in: `SELECT usr,sys FROM cpu,mem where total=1 order by sys, usr desc limit 10 offset 1000`, out: `SELECT usr, sys FROM cpu, mem WHERE (total = 1) ORDER BY sys ASC, usr DESC LIMIT 10 OFFSET 1000`},
		{in: "SELECT `usr`,`sys` FROM `cpu`,`mem` where total=1 order by `sys` desc, `usr`", out: "SELECT `usr`, `sys` FROM `cpu`, `mem` WHERE (total = 1) ORDER BY `sys` DESC, `usr` ASC"},
		{in: `SELECT usr,sys FROM cpu,mem where total=1 order by sys desc, usr limit 10 offset 1000`, out: `SELECT usr, sys FROM cpu, mem WHERE (total = 1) ORDER BY sys DESC, usr ASC LIMIT 10 OFFSET 1000`},
		{in: `SELECT total,usr as t2 FROM cpu`, out: `SELECT total, usr AS t2 FROM cpu`},
		{in: `SELECT account(total),usr as t2 FROM cpu`, out: `SELECT account( total ), usr AS t2 FROM cpu`},
		{in: `SELECT map(total,t2,t3),usr as t2 FROM cpu`, out: `SELECT map( total, t2, t3 ), usr AS t2 FROM cpu`},
		{in: `SELECT map(total,t2,t3) as a,usr as t2 FROM cpu where sys=filter(a1,a2)`, out: `SELECT map( total, t2, t3 ) AS a, usr AS t2 FROM cpu WHERE (sys = filter( a1, a2 ))`},
		{in: `SELECT distinct usr,sys FROM cpu`, out: `SELECT DISTINCT usr, sys FROM cpu`},
		{in: `SELECT distinct usr as u,sys FROM cpu`, out: `SELECT DISTINCT usr AS u, sys FROM cpu`},
		{in: `SELECT distinct count(usr) FROM cpu`, out: `SELECT DISTINCT count( usr ) FROM cpu`},
		{in: `SELECT count(distinct usr,idle), sys FROM cpu`, out: `SELECT count( DISTINCT usr, idle ), sys FROM cpu`},
		{in: `SELECT count(distinct usr,idle), sys FROM cpu where usr like ? or idle like ?`, out: `SELECT count( DISTINCT usr, idle ), sys FROM cpu WHERE ((usr LIKE ?) OR (idle LIKE ?))`},
	}
	for _, tc := range tests {
		out, err := parser.NewParser(lexer.NewLexer(strings.NewReader(tc.in))).Select()
		//fmt.Println(out.String())
		if err != nil {
			t.Fatalf("\nin:\t\t%s \nerror:\t%s", tc.in, err)
		} else if tc.out != out.String() {
			t.Fatalf("\nin:\t\t%s \nexpect:\t%s \ngot:\t%s", tc.in, tc.out, out)
		} else {
			fmt.Printf("in:\t\t%s \nexpect:\t%s \ngot:\t%s\n\n", tc.in, tc.out, out)
		}
	}
}

func TestWhereExpr(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err string
	}{
		{in: `func()`, out: `func()`},
		{in: `func('
			')`,
			out: `func( '
			' )`},
		{in: `func(aa,bb)`, out: `func( aa, bb )`},
		{in: `10`, out: `10`},
		{in: `'usr > 10'`, out: `'usr > 10'`},
		{in: `usr > 10`, out: `(usr > 10)`},
		{in: `usr = 10`, out: `(usr = 10)`},
		{in: `usr like ? or idle like ?`, out: `((usr LIKE ?) OR (idle LIKE ?))`},
		{in: `1*(2+3)/6=total and idle>1.2 or '5'=sys`, out: `(((((1 * (2 + 3)) / 6) = total) AND (idle > 1.2)) OR ('5' = sys))`},
		{in: `(name like ? or nick like ? or mail like ? or phone like ?) AND tenant= ?`, out: `(((((name LIKE ?) OR (nick LIKE ?)) OR (mail LIKE ?)) OR (phone LIKE ?)) AND (tenant = ?))`},
		{in: `type = ? and ((name like ? or user like ?)) AND tenant= ?`, out: `(((type = ?) AND ((name LIKE ?) OR (user LIKE ?))) AND (tenant = ?))`},
		{in: `(project_id in (52, 86)) AND ((tenant LIKE ?) OR (created_by LIKE ?) OR (created_at LIKE ?) OR (tenant_type LIKE ?) OR (tenant_status LIKE ?) OR (tenant_email LIKE ?))`,out:`(project_id in (52, 86)) AND ((tenant LIKE ?) OR (created_by LIKE ?) OR (created_at LIKE ?) OR (tenant_type LIKE ?) OR (tenant_status LIKE ?) OR (tenant_email LIKE ?))`},
	}
	for _, tc := range tests {
		out, err := parser.NewParser(lexer.NewLexer(strings.NewReader(tc.in))).Filter()
		//fmt.Println(out.String())
		if err != nil {
			t.Fatalf("\nin:\t\t%s, \nerror:\t%s", tc.in, err)
		} else if tc.out != out.String() {
			t.Fatalf("\nin:\t\t%s, \nexpect:\t%s, \ngot:\t%s", tc.in, tc.out, out)
		} else {
			fmt.Printf("in:\t\t%s, \nexpect:\t%s, \ngot:\t%s\n\n", tc.in, tc.out, out)
		}
	}
}

func TestExpr(t *testing.T) {
	var tests = []struct {
		in  string
		out string
		err string
	}{
		{in: `TimeIn(CloseAt, '{{ $value.WorkTime }}')=true`, out: `(TimeIn( CloseAt, '{{ $value.WorkTime }}' ) = true)`},
		{in: `'cpu alert. current value={{ field \"usr\"}}'`, out: `'cpu alert. current value={{ field \"usr\"}}'`},
		{in: `usr`, out: `usr`},
		{in: `usr+a`, out: `(usr + a)`},
	}
	for _, tc := range tests {
		out, err := parser.NewParser(lexer.NewLexer(strings.NewReader(tc.in))).Filter()
		//fmt.Println(out.String())
		if err != nil {
			t.Fatalf("\nin:\t\t%s, \nerror:\t%s", tc.in, err)
		} else if tc.out != out.String() {
			t.Fatalf("\nin:\t\t%s, \nexpect:\t%s, \ngot:\t%s", tc.in, tc.out, out)
		} else {
			fmt.Printf("in:\t\t%s, \nexpect:\t%s, \ngot:\t%s\n\n", tc.in, tc.out, out)
		}
	}
}
