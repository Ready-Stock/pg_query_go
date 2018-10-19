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
	case pq.NullTest:
		return deparse_nulltest(node)
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
	case pq.UpdateStmt:
		return deparse_update(node)
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
		if str, err := deparse_item(*node.Alias, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	result := strings.Join(out, " ")
	return &result, nil
}

func deparse_nulltest(node pq.NullTest) (*string, error) {
	out := make([]string, 0)
	if node.Arg == nil {
		return nil, errors.New("argument cannot be null for null test (ironically)")
	}

	if str, err := deparse_item(node.Arg, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	switch node.Nulltesttype {
	case pq.IS_NULL:
		out = append(out, "IS NULL")
	case pq.IS_NOT_NULL:
		out = append(out, "IS NOT NULL")
	default:
		return nil, errors.New("could not parse null test type (%d)").Format(node.Nulltesttype)
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
		out = append(out, "WHERE")
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

func deparse_update(node pq.UpdateStmt) (*string, error) {
	out := make([]string, 0)

	if node.WithClause != nil {
		if str, err := deparse_item(node.WithClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	out = append(out, "UPDATE")

	if node.Relation == nil {
		return nil, errors.New("relation of update statement cannot be null")
	}

	if str, err := deparse_item(*node.Relation, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.TargetList.Items == nil || len(node.TargetList.Items) == 0 {
		return nil, errors.New("update statement cannot have no sets")
	}

	if node.TargetList.Items != nil && len(node.TargetList.Items) > 0 {
		out = append(out, "SET")
		for _, target := range node.TargetList.Items {
			if str, err := deparse_item(target, &_Update); err != nil {
				return nil, err
			} else {
				out = append(out, *str)
			}
		}
	}

	if node.WhereClause != nil {
		out = append(out, "WHERE")
		if str, err := deparse_item(node.WhereClause, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.ReturningList.Items != nil && len(node.ReturningList.Items) > 0 {
		out = append(out, "RETURNING")
		returning := make([]string, len(node.ReturningList.Items))
		for i, slct := range node.ReturningList.Items {
			if str, err := deparse_item(slct, &_Select); err != nil {
				return nil, err
			} else {
				returning[i] = *str
			}
		}
		out = append(out, strings.Join(returning, ", "))
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
		out := make([]string, 0)
		if node.Name == nil || len(*node.Name) == 0 {
			return nil, errors.New("cannot have blank name for res target in update")
		}
		out = append(out, *node.Name)

		if node.Val == nil {
			return nil, errors.New("cannot have null value for res target in update")
		}

		if str, err := deparse_item(node.Val, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}

		result := strings.Join(out, " = ")
		return &result, nil
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
