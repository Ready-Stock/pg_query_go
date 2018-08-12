package pg_query

import (
	"strings"
	"fmt"
	"strconv"
	"reflect"
	pq "github.com/Ready-Stock/pg_query_go/nodes"
	"github.com/kataras/go-errors"
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
		return deparse_a_star()
	case pq.BoolExpr:
		//There is no BOOL_EXPR_NOT in go for some reason?
		switch node.Boolop {
		case pq.AND_EXPR:
			return deparse_bool_expr_and(node)
		case pq.OR_EXPR:
			return deparse_bool_expr_or(node)
		default:
			return nil, errors.New("cannot handle bool expression type (%d)").Format(node.Boolop)
		}
	case pq.CaseExpr:
		return deparse_case(node)
	case pq.CaseWhen:
		return deparse_when(node)
	case pq.ColumnRef:
		return deparse_columnref(node)
	case pq.InsertStmt:
		return deparse_insert_into(node)
	case pq.JoinExpr:
		return deparse_joinexpr(node)
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
	case pq.SQLValueFunction:
		return deparse_sqlvaluefunction(node)
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
		return nil, errors.New("cannot handle node type (%s)").Format(reflect.TypeOf(node).String())
	}
}

func deparse_aexpr(node pq.A_Expr, ctx *contextType) (*string, error) {
	out := make([]string, 0)
	if node.Lexpr == nil {
		return nil, errors.New("lexpr of expression cannot be null")
	} else {
		switch n := node.Lexpr.(type) {
		case pq.List:
			if n.Items == nil || len(n.Items) == 0 {
				return nil, errors.New("lexpr list cannot be empty")
			}
			if str, err := deparse_item(n.Items[0], ctx); err != nil {
				return nil, err
			} else {
				out = append(out, *str)
			}
		default:
			if str, err := deparse_item(n, ctx); err != nil {
				return nil, err
			} else {
				out = append(out, *str)
			}
		}
	}

	if node.Lexpr == nil {
		return nil, errors.New("rexpr of expression cannot be null")
	} else {
		if str, err := deparse_item(node.Rexpr, ctx); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.Name.Items == nil || len(node.Name.Items) == 0 {
		return nil, errors.New("error, expression name cannot be null")
	}

	if name, err := deparse_item(node.Name.Items[0], &_Operator); err != nil {
		return nil, err
	} else {
		result := strings.Join(out, fmt.Sprintf(" %s ", *name))
		if ctx != nil {
			result = fmt.Sprintf("(%s)", result)
		}
		return &result, nil
	}
}

func deparse_aexpr_in(node pq.A_Expr) (*string, error) {
	out := make([]string, 0)

	if node.Rexpr == nil {
		return nil, errors.New("rexpr of IN expression cannot be null")
	}


	// TODO (@elliotcourant) convert to handle list
	if strs, err := deparse_item_list(node.Rexpr.(pq.List).Items, nil); err != nil {
		return nil, err
	} else {
		out = append(out, strs...)
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
			result := fmt.Sprintf("%s %s (%s)", *str, operator, strings.Join(out, ", "))
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
			result := fmt.Sprintf(`%s (%s)`, *node.Aliasname, cols)
			return &result, nil
		}
	} else {
		return node.Aliasname, nil
	}
}

func deparse_a_const(node pq.A_Const) (*string, error) {
	return deparse_item(node.Val, &_A_CONST)
}

func deparse_a_star() (*string, error) {
	return &Star, nil
}

func deparse_bool_expr_and(node pq.BoolExpr) (*string, error) {
	if node.Args.Items == nil || len(node.Args.Items) == 0 {
		return nil, errors.New("args cannot be empty for boolean expression")
	}
	args := make([]string, len(node.Args.Items))
	for i, arg := range node.Args.Items {
		if str, err := deparse_item(arg, nil); err != nil {
			return nil, err
		} else {
			t := reflect.TypeOf(arg)
			if t == reflect.TypeOf(pq.BoolExpr{}) && arg.(pq.BoolExpr).Boolop == pq.OR_EXPR {
				args[i] = fmt.Sprintf("(%s)", *str)
			} else {
				args[i] = *str
			}
		}
	}
	result := strings.Join(args, " AND ")
	return &result, nil
}

func deparse_bool_expr_or(node pq.BoolExpr) (*string, error) {
	if node.Args.Items == nil || len(node.Args.Items) == 0 {
		return nil, errors.New("args cannot be empty for boolean expression")
	}
	args := make([]string, len(node.Args.Items))
	for i, arg := range node.Args.Items {
		if str, err := deparse_item(arg, nil); err != nil {
			return nil, err
		} else {
			t := reflect.TypeOf(arg)
			if t == reflect.TypeOf(pq.BoolExpr{}) && (arg.(pq.BoolExpr).Boolop == pq.OR_EXPR || arg.(pq.BoolExpr).Boolop == pq.AND_EXPR)  {
				args[i] = fmt.Sprintf("(%s)", *str)
			} else {
				args[i] = *str
			}
		}
	}
	result := strings.Join(args, " OR ")
	return &result, nil
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
		if str, err := deparse_item(*node.Alias, nil); err != nil {
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

func deparse_joinexpr(node pq.JoinExpr) (*string, error) {
	out := make([]string, 0)

	if node.Larg == nil {
		return nil, errors.New("larg of join cannot be null")
	}

	if str, err := deparse_item(node.Larg, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	switch node.Jointype {
	case pq.JOIN_INNER:
		if node.IsNatural {
			out = append(out, "NATURAL")
		} else if node.Quals == nil && (node.UsingClause.Items == nil || len(node.UsingClause.Items) == 0) {
			out = append(out, "CROSS")
		}
	case pq.JOIN_LEFT:
		out = append(out, "LEFT")
	case pq.JOIN_FULL:
		out = append(out, "FULL")
	case pq.JOIN_RIGHT:
		out = append(out, "RIGHT")
	default:
		return nil, errors.New("cannot handle join type (%d)").Format(node.Jointype)
	}
	out = append(out, "JOIN")

	if node.Rarg == nil {
		return nil, errors.New("rarg of join cannot be null")
	}

	if str, err := deparse_item(node.Rarg, nil); err != nil {
		return nil, err
	} else {
		out = append(out, *str)
	}

	if node.Quals != nil {
		out = append(out, "ON")
		if str, err := deparse_item(node.Quals, nil); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.UsingClause.Items != nil && len(node.UsingClause.Items) > 0 {
		clauses := make([]string, len(node.UsingClause.Items))
		for i, field := range node.UsingClause.Items {
			if str, err := deparse_item(field, &_Select); err != nil {
				return nil, err
			} else {
				clauses[i] = *str
			}
		}
		out = append(out, fmt.Sprintf("USING (%s)", strings.Join(clauses, ", ")))
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
	if reflect.DeepEqual(names, []string {"pg_catalog", "interval"}) {
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
		out[len(out) - 1] = fmt.Sprintf("%s[]", out[len(out) - 1])
	}

	result := strings.Join(out, ", ")
	return &result, nil
}

func deparse_typename_cast(names []string, arguments string) (*string, error) {
	if names[0] != "pg_catalog" {
		result := strings.Join(names, ".")
		return &result, nil
	}

	switch names[len(names) - 1] {
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
		return nil, errors.New("cannot deparse type: %s").Format(names[len(names) - 1])
	}
	return nil, nil
}

func deparse_interval_type(node pq.TypeName) (*string, error) {
	out := []string{ "interval" }

	if node.Typmods.Items != nil && len(node.Typmods.Items) > 0 {
		return nil, nil
		// In the ruby version of this code this was here to
		// handle `interval hour to second(5)` but i've not
		// ever seen that syntax and will come back to it
	}

	result := strings.Join(out, " ")
	return &result, nil
}
