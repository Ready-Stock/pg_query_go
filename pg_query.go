package pg_query

import (
	"encoding/json"
	"runtime/debug"

	"github.com/Ready-Stock/pg_query_go/parser"
	pq "github.com/Ready-Stock/pg_query_go/nodes"
	"strings"
	"reflect"
	"github.com/kataras/go-errors"
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
	True   contextType = "True"
	False  contextType = "False"
	Select contextType = "Select"
	Update contextType = "Update"
)

func Deparse(n pq.Node) (*string, error) {
	return deparse_item(n, nil)
}

func deparse_item(n pq.Node, ctx *contextType) (*string, error) {
	switch node := n.(type) {
	case pq.A_Expr:
		switch node.Kind {
		case pq.AEXPR_OP:
			return deparse_aexpr(node, ctx)
		default:
			return nil, nil
		}
	case pq.InsertStmt:
		return deparse_insert_into(node)
	case pq.SelectStmt:
		return deparse_select(node)
	case pq.RawStmt:
		return Deparse(node.Stmt)
	default:
		t := reflect.TypeOf(node).String()
		return nil, errors.New("cannot handle node type (%s)").Format(t)
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

func deparse_insert_into(node pq.InsertStmt) (*string, error) {
	out := make([]string, 0)
	if node.WithClause != nil {
		if str, err := deparse_item(node.WithClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	out = append(out, "INSERT INTO")
	if str, err := deparse_item(node.Relation, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.Cols.Items != nil {
		cols := make([]string, 0)
		for _, col := range node.Cols.Items {
			if str, err := deparse_item(col, nil); err != nil {
				return nil, err
			} else {
				cols = append(cols, *str)
			}
		}
		out = append(out, " (" + strings.Join(cols, ",") + ") ")
	}

	if str, err := deparse_item(node.SelectStmt, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
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

	//Get select *distinct* *fields*
	if node.TargetList.Items != nil && len(node.TargetList.Items) > 0 {
		out = append(out, "SELECT")
		if node.DistinctClause.Items != nil && len(node.DistinctClause.Items) > 0 {
			out = append(out, "DISTINCT")
		}
		fields := make([]string, len(node.TargetList.Items))
		ctx := Select
		for i, field := range node.TargetList.Items {
			if str, err := deparse_item(field, &ctx); err != nil {
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
		ctx := Select
		for i, from := range node.FromClause.Items {
			if str, err := deparse_item(from, &ctx); err != nil {
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
			out = append(out, "(" + strings.Join(values, ", ") + ")")
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
