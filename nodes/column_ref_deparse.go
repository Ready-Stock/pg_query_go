// Auto-generated - DO NOT EDIT

package pg_query

import (
	"fmt"
	"github.com/juju/errors"
	"strings"
)

func (node ColumnRef) Deparse(ctx Context) (*string, error) {
	if node.Fields.Items == nil || len(node.Fields.Items) == 0 {
		return nil, errors.New("columnref cannot have 0 fields")
	}
	out := make([]string, len(node.Fields.Items))
	for i, field := range node.Fields.Items {
		switch f := field.(type) {
		case String:
			out[i] = fmt.Sprintf(`"%s"`, f.Str)
		default:
			if str, err := deparseNode(field, Context_None); err != nil {
				return nil, err
			} else {
				out[i] = *str
			}
		}
	}
	result := strings.Join(out, ".")
	return &result, nil
}
