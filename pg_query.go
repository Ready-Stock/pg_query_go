package pg_query

import (
	"encoding/json"
	"runtime/debug"

	"github.com/Ready-Stock/pg_query_go/parser"
	pq "github.com/Ready-Stock/pg_query_go/nodes"
	"strings"
	"reflect"
	"github.com/kataras/go-errors"
	"fmt"
	"strconv"
)

// ParseToJSON - Parses the given SQL statement into an AST (JSON format)
func ParseToJSON(input string) (result string, err error) {
	return parser.ParseToJSON(input)
}

// Parse the given SQL statement into an AST (native Go structs)
func Parse(input string) (tree *ParsetreeList, err error) {
	jsonTree, err := ParseToJSON(input)
	if err != nil {
		return
	}

	// JSON unmarshalling can panic in edge cases we don't support yet. This is
	// still a *bug that needs to be fixed*, but this way the caller can expect an
	// error to be returned always, instead of a panic
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			err = r.(error)
		}
	}()

	err = json.Unmarshal([]byte(jsonTree), &tree)
	tree.Query = input
	return
}

type contextType string

const (
	True      contextType = "True"
	False     contextType = "False"
	Select    contextType = "Select"
	Update    contextType = "Update"
	A_CONST   contextType = "A_CONST"
	FUNC_CALL contextType = "FUNC_CALL"
	TYPE_NAME contextType = "TYPE_NAME"
	Operator  contextType = "Operator"
)

var (
	Star       = "*"
	_True      = True
	_False     = False
	_Select    = Select
	_Update    = Update
	_A_CONST   = A_CONST
	_FUNC_CALL = FUNC_CALL
	_TYPE_NAME = TYPE_NAME
	_Operator  = Operator
)

func Deparse(node pq.Node) (*string, error) {
	return deparse_item(node, nil)
}

func deparse_item(n pq.Node, ctx *contextType) (*string, error) {
	switch node := n.(type) {
	case pq.A_Expr:
		switch node.Kind {
		case pq.AEXPR_OP:
			return deparse_aexpr(node, ctx)
		case pq.AEXPR_IN:
			return deparse_aexpr_in(node)
		default:
			return nil, nil
		}
	case pq.Alias:
		return deparse_alias(node)
	case pq.A_Const:
		return deparse_a_const(node)
	case pq.A_Star:
		return deparse_a_star(node)
	case pq.CaseExpr:
		return deparse_case(node)
	case pq.CaseWhen:
		return deparse_when(node)
	case pq.ColumnRef:
		return deparse_columnref(node)
	case pq.InsertStmt:
		return deparse_insert_into(node)
	case pq.RangeVar:
		return deparse_rangevar(node)
	case pq.RawStmt:
		if result, err := Deparse(node.Stmt); err != nil {
			return nil, err
		} else {
			result := fmt.Sprintf("%s;", *result)
			return &result, nil
		}
	case pq.ResTarget:
		return deparse_restarget(node, ctx)
	case pq.SelectStmt:
		return deparse_select(node)
	case pq.SQLValueFunction:
		return deparse_sqlvaluefunction(node)
	case pq.String:
		switch *ctx {
		case A_CONST:
			result := fmt.Sprintf("'%s'", strings.Replace(node.Str, "'", "''", -1))
			return &result, nil
		case FUNC_CALL, TYPE_NAME:
			return &node.Str, nil
		default:
			result := fmt.Sprintf(`"%s"`, strings.Replace(node.Str, `"`, `""`, -1))
			return &result, nil
		}
	case pq.Integer:
		result := strconv.FormatInt(node.Ival, 10)
		return &result, nil
	default:
		return nil, errors.New("cannot handle node type (%s)").Format(reflect.TypeOf(node).String())
	}
}

func deparse_aexpr(node pq.A_Expr, ctx *contextType) (*string, error) {
	// output := make([]string, 0)
	// if str, err := deparse_item(node.Lexpr, true); err != nil {
	// 	return nil, err
	// } else {
	// 	output = append(output, *str)
	// }
	// if str, err := deparse_item(node.Rexpr, true); err != nil {
	// 	return nil, err
	// } else {
	// 	output = append(output, *str)
	// }
	// if name, err := deparse_item(node.Name[0]); err != nil {
	// 	return nil, err
	// } else {
	// 	str := strings.Join(output, " " + name + " ")
	// }
	return nil, nil
}

func deparse_aexpr_in(node pq.A_Expr) (*string, error) {
	out := make([]string, 0)

	if node.Rexpr == nil {
		return nil, errors.New("rexpr of IN expression cannot be null")
	}


	// TODO (@elliotcourant) convert to handle list
	if str, err := deparse_item(node.Rexpr, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.Name.Items == nil || len(node.Name.Items) == 0 {
		return nil, errors.New("names of IN expression cannot be empty")
	}

	if strs, err := deparse_item_list(node.Name.Items, &_Operator); err != nil {
		return nil, err
	} else {
		operator := ""
		if reflect.DeepEqual(strs, []string{"="}) {
			operator = "IN"
		} else {
			operator = "NOT IN"
		}

		if node.Lexpr == nil {
			return nil, errors.New("lexpr of IN expression cannot be null")
		}

		if str, err := deparse_item(node.Lexpr, nil); err != nil {
			return nil, err
		} else {
			result := fmt.Sprintf("%s %s (%s)", str, operator, strings.Join(out, ", "))
			return &result, nil
		}
	}
}

func deparse_alias(node pq.Alias) (*string, error) {
	if node.Colnames.Items != nil && len(node.Colnames.Items) > 0 {
		if colnames, err := deparse_item_list(node.Colnames.Items, nil); err != nil {
			return nil, err
		} else {
			cols := strings.Join(colnames, ", ")
			result := fmt.Sprintf(`%s (%s)`, node.Aliasname, cols)
			return &result, nil
		}
	} else {
		return node.Aliasname, nil
	}
}

func deparse_a_const(node pq.A_Const) (*string, error) {
	return deparse_item(node.Val, &_A_CONST)
}

func deparse_a_star(node pq.A_Star) (*string, error) {
	return &Star, nil
}

func deparse_case(node pq.CaseExpr) (*string, error) {
	out := []string{"CASE"}

	if node.Arg != nil {
		if str, err := deparse_item(node.Arg, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.Args.Items == nil || len(node.Args.Items) == 0 {
		return nil, errors.New("case expression cannot have no arguments")
	}

	if args, err := deparse_item_list(node.Args.Items, nil); err != nil {
		return nil, err
	} else {
		out = append(out, args...)
	}

	if node.Defresult != nil {
		out = append(out, "ELSE")
		if str, err := deparse_item(node.Defresult, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	out = append(out, "END")
	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_columnref(node pq.ColumnRef) (*string, error) {
	if node.Fields.Items == nil || len(node.Fields.Items) == 0 {
		return nil, errors.New("columnref cannot have no fields")
	}
	out := make([]string, len(node.Fields.Items))
	for i, field := range node.Fields.Items {
		switch f := field.(type) {
		case pq.String:
			out[i] = fmt.Sprintf(`"%s"`, f.Str)
		default:
			if str, err := deparse_item(field, nil); err != nil {
				return nil, err
			} else {
				out[i] = *str
			}
		}
	}
	result := strings.Join(out, ".")
	return &result, nil
}

func deparse_rangevar(node pq.RangeVar) (*string, error) {
	out := make([]string, 0)
	if !node.Inh {
		out = append(out, "ONLY")
	}

	if node.Schemaname != nil && len(*node.Schemaname) > 0 {
		out = append(out, fmt.Sprintf(`"%s"."%s"`, *node.Schemaname, *node.Relname))
	} else {
		out = append(out, fmt.Sprintf(`"%s"`, *node.Relname))
	}

	if node.Alias != nil {
		if str, err := deparse_item(node.Alias, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_insert_into(node pq.InsertStmt) (*string, error) {
	out := make([]string, 0)
	if node.WithClause != nil {
		if str, err := deparse_item(node.WithClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.Relation == nil {
		return nil, errors.New("relation in insert cannot be null!")
	}
	out = append(out, "INSERT INTO")
	if str, err := deparse_item(*node.Relation, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.Cols.Items != nil {
		cols := make([]string, len(node.Cols.Items))
		for i, col := range node.Cols.Items {
			if str, err := deparse_item(col, nil); err != nil {
				return nil, err
			} else {
				cols[i] = *str
			}
		}
		out = append(out, fmt.Sprintf("(%s)", strings.Join(cols, ",")))
	}

	if str, err := deparse_item(node.SelectStmt, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.ReturningList.Items != nil && len(node.ReturningList.Items) > 0 {
		out = append(out, "RETURNING")
		fields := make([]string, len(node.ReturningList.Items))
		for i, field := range node.ReturningList.Items {
			if str, err := deparse_item(field, &_Select); err != nil {
				return nil, err
			} else {
				fields[i] = *str
			}
		}
		out = append(out, strings.Join(fields, ", "))
	}

	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_select(node pq.SelectStmt) (*string, error) {
	out := make([]string, 0)
	if node.Op == pq.SETOP_UNION {
		if str, err := deparse_item(node.Larg, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}

		out = append(out, "UNION")
		if node.All {
			out = append(out, "ALL")
		}

		if str, err := deparse_item(node.Rarg, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}

		result := strings.Join(out, " ")
		return &result, nil
	}

	if node.WithClause != nil {
		if str, err := deparse_item(node.WithClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	// Get select *distinct* *fields*
	if node.TargetList.Items != nil && len(node.TargetList.Items) > 0 {
		out = append(out, "SELECT")
		if node.DistinctClause.Items != nil && len(node.DistinctClause.Items) > 0 {
			out = append(out, "DISTINCT")
		}
		fields := make([]string, len(node.TargetList.Items))
		for i, field := range node.TargetList.Items {
			if str, err := deparse_item(field, &_Select); err != nil {
				return nil, err
			} else {
				fields[i] = *str
			}
		}
		out = append(out, strings.Join(fields, ", "))
	}

	if node.FromClause.Items != nil && len(node.FromClause.Items) > 0 {
		out = append(out, "FROM")
		froms := make([]string, len(node.FromClause.Items))
		for i, from := range node.FromClause.Items {
			if str, err := deparse_item(from, &_Select); err != nil {
				return nil, err
			} else {
				froms[i] = *str
			}
		}
		out = append(out, strings.Join(froms, ", "))
	}

	if node.WhereClause != nil {
		if str, err := deparse_item(node.WhereClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.ValuesLists != nil && len(node.ValuesLists) > 0 {
		out = append(out, "VALUES")
		for _, valuelist := range node.ValuesLists {
			values := make([]string, len(valuelist))
			for i, value := range valuelist {
				if str, err := deparse_item(value, nil); err != nil {
					return nil, err
				} else {
					values[i] = *str
				}
			}
			out = append(out, "("+strings.Join(values, ", ")+")")
		}
	}

	if node.GroupClause.Items != nil && len(node.GroupClause.Items) > 0 {
		out = append(out, "GROUP BY")
		groups := make([]string, len(node.GroupClause.Items))
		for i, group := range node.GroupClause.Items {
			if str, err := deparse_item(group, nil); err != nil {
				return nil, err
			} else {
				groups[i] = *str
			}
		}
		out = append(out, strings.Join(groups, ", "))
	}

	if node.HavingClause != nil {
		if str, err := deparse_item(node.HavingClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	// Sort clause

	if node.LimitCount != nil {
		out = append(out, "LIMIT")
		if str, err := deparse_item(node.LimitCount, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.LimitOffset != nil {
		out = append(out, "OFFSET")
		if str, err := deparse_item(node.LimitOffset, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.LockingClause.Items != nil && len(node.LockingClause.Items) > 0 {
		for _, lock := range node.LockingClause.Items {
			if str, err := deparse_item(lock, nil); err != nil {
				return nil, err
			} else {
				out = append(out, *str)
			}
		}
	}

	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_sqlvaluefunction(node pq.SQLValueFunction) (*string, error) {
	switch node.Op {
	case pq.SVFOP_CURRENT_TIMESTAMP:
		result := "CURRENT_TIMESTAMP"
		return &result, nil
	}
	return nil, nil
}

func deparse_restarget(node pq.ResTarget, ctx *contextType) (*string, error) {
	if ctx == nil {
		return node.Name, nil
	} else if *ctx == Select {
		out := make([]string, 0)
		if str, err := deparse_item(node.Val, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}

		if node.Name != nil && len(*node.Name) > 0 {
			out = append(out, "AS")
			out = append(out, *node.Name)
		}
		result := strings.Join(out, " ")
		return &result, nil
	} else if *ctx == Update {
		return nil, nil
	} else {
		return nil, nil
	}
}

func deparse_item_list(nodes []pq.Node, ctx *contextType) ([]string, error) {
	out := make([]string, len(nodes))
	for i, node := range nodes {
		if str, err := deparse_item(node, ctx); err != nil {
			return nil, err
		} else {
			out[i] = *str
		}
	}
	return out, nil
}

func deparse_when(node pq.CaseWhen) (*string, error) {
	out := []string{"WHEN"}

	if str, err := deparse_item(node.Expr, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	out = append(out, "THEN")

	if str, err := deparse_item(node.Result, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	result := strings.Join(out, " ")
	return &result, nil
}

// ParsePlPgSqlToJSON - Parses the given PL/pgSQL function statement into an AST (JSON format)
func ParsePlPgSqlToJSON(input string) (result string, err error) {
	return parser.ParsePlPgSqlToJSON(input)
}

// Normalize the passed SQL statement to replace constant values with ? characters
func Normalize(input string) (result string, err error) {
	return parser.Normalize(input)
}

// FastFingerprint - Fingerprint the passed SQL statement using the C extension
func FastFingerprint(input string) (result string, err error) {
	return parser.FastFingerprint(input)
}
