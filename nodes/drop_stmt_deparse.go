// Auto-generated - DO NOT EDIT

package pg_query

import (
	"fmt"
	"strings"
)

func (node DropStmt) Deparse(ctx Context) (*string, error) {
	out := []string{"DROP", ""}

	switch node.RemoveType {
	case OBJECT_TABLE:
		out[1] = "TABLE"
	default:
		panic(fmt.Sprintf("cannot handle remove type [%v]", node.RemoveType))
	}

	objects := make([]string, len(node.Objects.Items))
	for i, objList := range node.Objects.Items {
		list := objList.(List)
		if objs, err := list.DeparseList(Context_None); err != nil {
			return nil, err
		} else {
			objects[i] = strings.Join(objs, ".")
		}
	}

	out = append(out, strings.Join(objects, ", "))

	switch node.Behavior {
	case DROP_CASCADE:
		out = append(out, "CASCADE")
	default:
		panic(fmt.Sprintf("cannot handle drop behavior [%v]", node.Behavior))
	}

	result := strings.Join(out, " ")
	return &result, nil
}
