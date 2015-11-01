// Auto-generated - DO NOT EDIT

package pg_query

import "encoding/json"

type SetToDefault struct {
	Xpr       Expr  `json:"xpr"`
	TypeId    Oid   `json:"typeId"`    /* type for substituted value */
	TypeMod   int32 `json:"typeMod"`   /* typemod for substituted value */
	Collation Oid   `json:"collation"` /* collation for the substituted value */
	Location  int   `json:"location"`  /* token location, or -1 if unknown */
}

func (node SetToDefault) MarshalJSON() ([]byte, error) {
	type SetToDefaultMarshalAlias SetToDefault
	return json.Marshal(map[string]interface{}{
		"SETTODEFAULT": (*SetToDefaultMarshalAlias)(&node),
	})
}

func (node *SetToDefault) UnmarshalJSON(input []byte) (err error) {
	err = UnmarshalNodeFieldJSON(input, node)
	return
}
