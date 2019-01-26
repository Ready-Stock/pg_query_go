// Auto-generated - DO NOT EDIT

package pg_query

import (
	"strings"
)

func (node DeleteStmt) Deparse(ctx Context) (*string, error) {
	out := []string{"DELETE FROM", ""}

	if table, err := node.Relation.Deparse(Context_None); err != nil {
		return nil, err
	} else {
		out[1] = *table
	}

	if node.WhereClause != nil {
		out = append(out, "WHERE")
		if str, err := deparseNode(node.WhereClause, Context_None); err != nil {
			return nil, err
		} else {
			out = append(out, *str)
		}
	}

	if node.ReturningList.Items != nil && len(node.ReturningList.Items) > 0 {
		out = append(out, "RETURNING")
		fields := make([]string, len(node.ReturningList.Items))
		for i, field := range node.ReturningList.Items {
			if str, err := deparseNode(field, Context_Select); err != nil {
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
