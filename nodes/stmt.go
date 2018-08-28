package pg_query

type Stmt interface {
	StatementType()
	StatementTag() string
}