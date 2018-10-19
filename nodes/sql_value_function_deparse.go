// Auto-generated - DO NOT EDIT

package pg_query

func (node SQLValueFunction) Deparse(ctx Context) (*string, error) {
	switch node.Op {
	case SVFOP_CURRENT_TIMESTAMP:
		result := "CURRENT_TIMESTAMP"
		return &result, nil
	}
	return nil, nil
}
