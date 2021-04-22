# The Nerv Query Language Specification

## Notation

The syntax is specified using Extended Backus-Naur Form ("EBNF").  EBNF is the
same notation used in the [Go](http://golang.org) programming language
specification, which can be found [here](https://golang.org/ref/spec).  Not so
coincidentally, InfluxDB is written in Go.

```
Production  = production_name "=" [ Expression ] "." .
Expression  = Alternative { "|" Alternative } .
Alternative = Term { Term } .
Term        = production_name | token [ "…" token ] | Group | Option | Repetition .
Group       = "(" Expression ")" .
Option      = "[" Expression "]" .
Repetition  = "{" Expression "}" .
```

Notation operators in order of increasing precedence:

```
|   alternation
()  grouping
[]  option (0 or 1 times)
{}  repetition (0 to n times)
```

## Tokens

```
IDENT               = LETTER { LETTER | DIGIT }
STRING              = "'" {.} "'"
INTEGER             = ["-"] DIGIT { DIGIT }
FLOAT               = ["-"] DIGIT "." DIGIT { DIGIT }
LETTER              = ASCII_LETTER | "_"
ASCII_LETTER        = "A" … "Z" | "a" … "z"
DIGIT               = "0" … "9"
TRUE                = "true"
FALSE               = "false"
```

## Dates & Times

The date and time literal format is not specified in EBNF like the rest of this document.  It is specified using Go's date / time parsing format, which is a reference date written in the format required by NervQL.  The reference date time is:

NervQL reference date time: January 2nd, 2006 at 3:04:05 PM

```
time_lit            = "2006-01-02 15:04:05.999999" | "2006-01-02" .
```

## Durations

Duration literals specify a length of time.  An integer literal followed
immediately (with no spaces) by a duration unit listed below is interpreted as
a duration literal.

## Duration units
| Units  | Meaning                                 |
|--------|-----------------------------------------|
| u or µ | microseconds (1 millionth of a second)  |
| ms     | milliseconds (1 thousandth of a second) |
| s      | second                                  |
| m      | minute                                  |
| h      | hour                                    |
| d      | day                                     |
| w      | week                                    |

```
duration_lit        = int_lit duration_unit .
duration_unit       = "u" | "µ" | "ms" | "s" | "m" | "h" | "d" | "w" .
```

## SELECT

```
select          =   "SELECT" fields
                    from
                    [ where ]
                    [ group ]
                    [ order ] 
                    [ limit ]
                    [ offset ]

fields          =   field { "," field }

field           =   expr [ alias ] .

alias           =   "AS" IDENT

from            =   "FROM" table { "," table }

table           =   IDENT

where           =   "WHERE" expr

group           =   "ORDER BY" group_field { "," group_field }


order           =   "ORDER BY" sort_field { "," sort_field }

sort_field      =   IDENT [ ASC | DESC ]

group_field     =   IDENT

limit           =   "LIMIT" INTEGER

offset          =   "OFFSET" INTEGER

```

## Expressions

```
expr                = unary_expr { binary_op unary_expr }

unary_expr          = "(" expr ")" | fieldRef | STRING | INTEGER | FLOAT | TRUE | FALSE

binary_op           = "+" | "-" | "*" | "/" | "%"
                    | "=" | "!=" | "<>" | "<" | "<=" | ">" | ">="
                    | "AND" | "OR"

fieldRef               = IDENT
```

