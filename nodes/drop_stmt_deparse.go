// Auto-generated - DO NOT EDIT

package pg_query

import (
	"fmt"
	"strings"
)

func (node DropStmt) Deparse(ctx Context) (*string, error) {
	out := []string{"DROP", ""}

	out[1] = func() string {
		switch node.RemoveType {
		case OBJECT_ACCESS_METHOD:
			return "ACCESS METHOD"
		case OBJECT_AGGREGATE:
			return "AGGREGATE"
		case OBJECT_CAST:
			return "CAST"
		case OBJECT_TABLE:
			return "TABLE"
		default:
			panic(fmt.Sprintf("cannot handle remove type [%s]", node.RemoveType.String()))
		}
	}()

	if node.MissingOk {
		out = append(out, "IF EXISTS")
	}

	objects := make([]string, len(node.Objects.Items))
	for i, obj := range node.Objects.Items {
		switch obj.(type) {
		case List:
			list := obj.(List)
			if objs, err := list.DeparseList(Context_None); err != nil {
				return nil, err
			} else {
				objects[i] = strings.Join(objs, ".")
			}
		default:
			if str, err := obj.Deparse(Context_None); err != nil {
				return nil, err
			} else {
				objects[i] = *str
			}
		}
	}

	out = append(out, strings.Join(objects, ", "))

	switch node.Behavior {
	case DROP_CASCADE:
		out = append(out, "CASCADE")
	case DROP_RESTRICT:
		// By default the drop will restrict, so there is no need to have this in there.
	}

	result := strings.Join(out, " ")
	return &result, nil
}
