// Auto-generated - DO NOT EDIT

package pg_query

type DropBehavior uint

const (
	DROP_RESTRICT = iota /* drop fails if any dependent objects */
	DROP_CASCADE         /* remove dependent objects too */

)
