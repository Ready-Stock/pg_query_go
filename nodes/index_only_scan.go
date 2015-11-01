// Auto-generated - DO NOT EDIT

package pg_query

import "encoding/json"

type IndexOnlyScan struct {
	Scan          Scan          `json:"scan"`
	Indexid       Oid           `json:"indexid"`       /* OID of index to scan */
	Indexqual     []Node        `json:"indexqual"`     /* list of index quals (usually OpExprs) */
	Indexorderby  []Node        `json:"indexorderby"`  /* list of index ORDER BY exprs */
	Indextlist    []Node        `json:"indextlist"`    /* TargetEntry list describing index's cols */
	Indexorderdir ScanDirection `json:"indexorderdir"` /* forward or backward or don't care */
}

func (node IndexOnlyScan) MarshalJSON() ([]byte, error) {
	type IndexOnlyScanMarshalAlias IndexOnlyScan
	return json.Marshal(map[string]interface{}{
		"INDEXONLYSCAN": (*IndexOnlyScanMarshalAlias)(&node),
	})
}

func (node *IndexOnlyScan) UnmarshalJSON(input []byte) (err error) {
	err = UnmarshalNodeFieldJSON(input, node)
	return
}
