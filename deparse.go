package pg_query

import (
	"fmt"
	pq "github.com/Ready-Stock/pg_query_go/nodes"
	"github.com/kataras/go-errors"
	"reflect"
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
	case pq.TransactionStmt:
		return deparse_transaction(node)
	case pq.VariableShowStmt:
		return deparse_variable_show_stmt(node)
	default:
		return nil, errors.New("cannot deparse node type %s").Format(reflect.TypeOf(node).String())
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

func deparse_variable_show_stmt(node pq.VariableShowStmt) (*string, error) {
	out := []string{"SHOW"}
	out = append(out, *node.Name)
	result := strings.Join(out, " ")
	return &result, nil
}
