// Auto-generated - DO NOT EDIT

package pg_query

import (
	"fmt"
	"github.com/juju/errors"
	"reflect"
	"strings"
)

func (node BoolExpr) Deparse(ctx Context) (*string, error) {
	// There is no BOOL_EXPR_NOT in go for some reason?
	switch node.Boolop {
	case AND_EXPR:
		return node.deparseBoolExprAnd()
	case OR_EXPR:
		return node.deparseBoolExprOr()
	default:
		return nil, errors.Errorf("cannot handle bool expression type (%d)", node.Boolop)
	}
}

func (node BoolExpr) deparseBoolExprAnd() (*string, error) {
	if node.Args.Items == nil || len(node.Args.Items) == 0 {
		return nil, errors.New("args cannot be empty for boolean expression")
	}
	args := make([]string, len(node.Args.Items))
	for i, arg := range node.Args.Items {
		if str, err := deparseNode(arg, Context_None); err != nil {
			return nil, err
		} else {
			t := reflect.TypeOf(arg)
			if t == reflect.TypeOf(BoolExpr{}) && arg.(BoolExpr).Boolop == OR_EXPR {
				args[i] = fmt.Sprintf("(%s)", *str)
			} else {
				args[i] = *str
			}
		}
	}
	result := strings.Join(args, " AND ")
	return &result, nil
}

func (node BoolExpr) deparseBoolExprOr() (*string, error) {
	if node.Args.Items == nil || len(node.Args.Items) == 0 {
		return nil, errors.New("args cannot be empty for boolean expression")
	}
	args := make([]string, len(node.Args.Items))
	for i, arg := range node.Args.Items {
		if str, err := deparseNode(arg, Context_None); err != nil {
			return nil, err
		} else {
			t := reflect.TypeOf(arg)
			if t == reflect.TypeOf(BoolExpr{}) && (arg.(BoolExpr).Boolop == OR_EXPR || arg.(BoolExpr).Boolop == AND_EXPR) {
				args[i] = fmt.Sprintf("(%s)", *str)
			} else {
				args[i] = *str
			}
		}
	}
	result := strings.Join(args, " OR ")
	return &result, nil
}
