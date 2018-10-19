package pg_query

import (
	"encoding/json"
	"fmt"
	pq "github.com/Ready-Stock/pg_query_go/nodes"
	"github.com/kataras/go-errors"
	"github.com/kataras/golog"
	"reflect"
	"strconv"
	"strings"
)

type contextType int64

const (
	True      contextType = 1
	False     contextType = 2
	Select    contextType = 4
	Update    contextType = 8
	A_CONST   contextType = 16
	FUNC_CALL contextType = 32
	TYPE_NAME contextType = 64
	Operator  contextType = 128
)

var (
	_Select    = Select
	_Update    = Update
	_TYPE_NAME = TYPE_NAME
	_A_CONST   = A_CONST
)

func Deparse(node pq.Node) (*string, error) {
	if sql, err := deparse_item(node, nil); err != nil {
		j, _ := json.Marshal(node)
		golog.Debugf("JSON: %s", string(j))
		return nil, err
	} else {
		return sql, nil
	}
}

func DeparseValue(aconst pq.A_Const) (interface{}, error) {
	switch c := aconst.Val.(type) {
	case pq.String:
		return c.Str, nil
	case pq.Integer:
		return c.Ival, nil
	case pq.Null:
		return nil, nil
	default:
		return nil, errors.New("cannot parse type %s").Format(reflect.TypeOf(c).Name())
	}
}

func deparse_item(n pq.Node, ctx *contextType) (*string, error) {
	switch node := n.(type) {
	case pq.WithClause:
		return deparse_with_clause(node)
	case pq.TypeCast:
		return deparse_typecast(node)
	case pq.TypeName:
		return deparse_typename(node)
	case pq.TransactionStmt:
		return deparse_transaction(node)
	case pq.SQLValueFunction:
		return deparse_sqlvaluefunction(node)
	case pq.VariableSetStmt:
		return deparse_variable_set_stmt(node)
	case pq.VariableShowStmt:
		return deparse_variable_show_stmt(node)
	case pq.String:
		switch *ctx {
		case A_CONST:
			result := fmt.Sprintf("'%s'", strings.Replace(node.Str, "'", "''", -1))
			return &result, nil
		case FUNC_CALL, TYPE_NAME, Operator:
			return &node.Str, nil
		default:
			result := fmt.Sprintf(`"%s"`, strings.Replace(node.Str, `"`, `""`, -1))
			return &result, nil
		}
	case pq.Integer:
		result := strconv.FormatInt(node.Ival, 10)
		return &result, nil
	case pq.Float:
		return &node.Str, nil
	case pq.Null:
		result := "NULL"
		return &result, nil
	default:
		return nil, errors.New("cannot deparse node type %s").Format(reflect.TypeOf(node).String())
	}
}

func deparse_sqlvaluefunction(node pq.SQLValueFunction) (*string, error) {
	switch node.Op {
	case pq.SVFOP_CURRENT_TIMESTAMP:
		result := "CURRENT_TIMESTAMP"
		return &result, nil
	}
	return nil, nil
}

func deparse_with_clause(node pq.WithClause) (*string, error) {
	out := []string{"WITH"}
	if node.Recursive {
		out = append(out, "RECURSIVE")
	}

	if node.Ctes.Items == nil || len(node.Ctes.Items) == 0 {
		return nil, errors.New("cannot have with clause without ctes")
	}

	ctes := make([]string, len(node.Ctes.Items))
	for i, cte := range node.Ctes.Items {
		if str, err := deparse_item(cte, nil); err != nil {
			return nil, err
		} else {
			ctes[i] = *str
		}
	}
	out = append(out, strings.Join(ctes, ", "))
	result := strings.Join(out, " ")
	return &result, nil
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

var transactionCmds = map[pq.TransactionStmtKind]string{
	pq.TRANS_STMT_BEGIN:             "BEGIN",
	pq.TRANS_STMT_START:             "BEGIN",
	pq.TRANS_STMT_COMMIT:            "COMMIT",
	pq.TRANS_STMT_ROLLBACK:          "ROLLBACK",
	pq.TRANS_STMT_SAVEPOINT:         "SAVEPOINT",
	pq.TRANS_STMT_RELEASE:           "RELEASE",
	pq.TRANS_STMT_ROLLBACK_TO:       "ROLLBACK TO SAVEPOINT",
	pq.TRANS_STMT_PREPARE:           "PREPARE TRANSACTION",
	pq.TRANS_STMT_COMMIT_PREPARED:   "COMMIT TRANSACTION",
	pq.TRANS_STMT_ROLLBACK_PREPARED: "ROLLBACK TRANSACTION",
}

func deparse_transaction(node pq.TransactionStmt) (*string, error) {
	out := make([]string, 0)
	if kind, ok := transactionCmds[node.Kind]; !ok {
		return nil, errors.New("couldn't deparse transaction kind: %s").Format(node.Kind)
	} else {
		out = append(out, kind)
	}

	if node.Kind == pq.TRANS_STMT_PREPARE ||
		node.Kind == pq.TRANS_STMT_COMMIT_PREPARED ||
		node.Kind == pq.TRANS_STMT_ROLLBACK_PREPARED {
		if node.Gid != nil {
			out = append(out, fmt.Sprintf("'%s'", *node.Gid))
		}
	} else {
		if node.Options.Items != nil && len(node.Options.Items) > 0 {

		}
	}

	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_typecast(node pq.TypeCast) (*string, error) {
	if node.TypeName == nil {
		return nil, errors.New("typename cannot be null in typecast")
	}
	if str, err := deparse_item(*node.TypeName, nil); err != nil {
		return nil, err
	} else {
		if val, err := deparse_item(node.Arg, nil); err != nil {
			return nil, err
		} else {
			if *str == "boolean" {
				if *val == "'t'" {
					result := "true"
					return &result, nil
				} else {
					result := "false"
					return &result, nil
				}
			} else {
				if typename, err := deparse_typename(*node.TypeName); err != nil {
					return nil, err
				} else {
					result := fmt.Sprintf("%s::%s", *val, *typename)
					return &result, nil
				}
			}
		}
	}
}

func deparse_typename(node pq.TypeName) (*string, error) {
	if node.Names.Items == nil || len(node.Names.Items) == 0 {
		return nil, errors.New("cannot have no names on type name")
	}
	names := make([]string, len(node.Names.Items))
	for i, name := range node.Names.Items {
		if str, err := deparse_item(name, &_TYPE_NAME); err != nil {
			return nil, err
		} else {
			names[i] = *str
		}
	}

	// Intervals are tricky and should be handled in a seperate method because they require some bitmask operations
	if reflect.DeepEqual(names, []string{"pg_catalog", "interval"}) {
		return deparse_interval_type(node)
	}

	out := make([]string, 0)
	if node.Setof {
		out = append(out, "SETOF")
	}

	args := ""
	if node.Typmods.Items != nil && len(node.Typmods.Items) > 0 {
		arguments := make([]string, len(node.Typmods.Items))
		for i, arg := range node.Typmods.Items {
			if str, err := deparse_item(arg, nil); err != nil {
				return nil, err
			} else {
				arguments[i] = *str
			}
		}
		args = strings.Join(arguments, ", ")
	}

	if str, err := deparse_typename_cast(names, args); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.ArrayBounds.Items != nil || len(node.ArrayBounds.Items) > 0 {
		out[len(out)-1] = fmt.Sprintf("%s[]", out[len(out)-1])
	}

	result := strings.Join(out, ", ")
	return &result, nil
}

func deparse_typename_cast(names []string, arguments string) (*string, error) {
	if names[0] != "pg_catalog" {
		result := strings.Join(names, ".")
		return &result, nil
	}

	switch names[len(names)-1] {
	case "bpchar":
		if len(arguments) == 0 {
			result := "char"
			return &result, nil
		} else {
			result := fmt.Sprintf("char(%s)", arguments)
			return &result, nil
		}
	case "varchar":
		if len(arguments) == 0 {
			result := "varchar"
			return &result, nil
		} else {
			result := fmt.Sprintf("varchar(%s)", arguments)
			return &result, nil
		}
	case "numeric":
		if len(arguments) == 0 {
			result := "numeric"
			return &result, nil
		} else {
			result := fmt.Sprintf("numeric(%s)", arguments)
			return &result, nil
		}
	case "bool":
		result := "boolean"
		return &result, nil
	case "int2":
		result := "smallint"
		return &result, nil
	case "int4":
		result := "int"
		return &result, nil
	case "int8":
		result := "bigint"
		return &result, nil
	case "real", "float4":
		result := "real"
		return &result, nil
	case "float8":
		result := "double"
		return &result, nil
	case "time":
		result := "time"
		return &result, nil
	case "timezt":
		result := "time with time zone"
		return &result, nil
	case "timestamp":
		result := "timestamp"
		return &result, nil
	case "timestamptz":
		result := "timestamp with time zone"
		return &result, nil
	default:
		return nil, errors.New("cannot deparse type: %s").Format(names[len(names)-1])
	}
	return nil, nil
}

func deparse_interval_type(node pq.TypeName) (*string, error) {
	out := []string{"interval"}

	if node.Typmods.Items != nil && len(node.Typmods.Items) > 0 {
		return nil, nil
		// In the ruby version of this code this was here to
		// handle `interval hour to second(5)` but i've not
		// ever seen that syntax and will come back to it
	}

	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_variable_set_stmt(node pq.VariableSetStmt) (*string, error) {
	out := []string{"SET"}
	if node.IsLocal {
		out = append(out, "LOCAL")
	}
	out = append(out, *node.Name)
	out = append(out, "TO")
	if args, err := deparse_item_list(node.Args.Items, nil); err != nil {
		return nil, err
	} else {
		out = append(out, args...)
	}
	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_variable_show_stmt(node pq.VariableShowStmt) (*string, error) {
	out := []string{"SHOW"}
	out = append(out, *node.Name)
	result := strings.Join(out, " ")
	return &result, nil
}
