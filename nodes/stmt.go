package pg_query

type Stmt interface {
	StatementType() StmtType
	StatementTag() string
}