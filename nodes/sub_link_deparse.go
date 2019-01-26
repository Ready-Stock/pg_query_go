// Auto-generated - DO NOT EDIT

package pg_query

import (
	"fmt"
)

func (node SubLink) Deparse(ctx Context) (*string, error) {
	switch node.SubLinkType {
	case EXPR_SUBLINK:
		if subSelect, err := node.Subselect.Deparse(Context_None); err != nil {
			return nil, err
		} else {
			result := fmt.Sprintf("(%s)", *subSelect)
			return &result, err
		}
	default:
		panic(fmt.Sprintf("cannot handle sub link type [%s]", node.SubLinkType.String()))
	}
}
