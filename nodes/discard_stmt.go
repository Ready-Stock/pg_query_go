// Auto-generated - DO NOT EDIT

package pg_query

import "encoding/json"

type DiscardStmt struct {
	Target DiscardMode `json:"target"`
}

func (node DiscardStmt) MarshalJSON() ([]byte, error) {
	type DiscardStmtMarshalAlias DiscardStmt
	return json.Marshal(map[string]interface{}{
		"DISCARDSTMT": (*DiscardStmtMarshalAlias)(&node),
	})
}

func (node *DiscardStmt) UnmarshalJSON(input []byte) (err error) {
	err = UnmarshalNodeFieldJSON(input, node)
	return
}
