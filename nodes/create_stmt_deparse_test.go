package pg_query

import (
	"testing"
)

func Test_CreateStmt_Generic(t *testing.T) {
	DoTest(t, DeparseTest{
		Query:    `CREATE TABLE test (id BIGSERIAL PRIMARY KEY, name TEXT);`,
		Expected: `CREATE TABLE "test" (id bigserial PRIMARY KEY, name text)`,
	})
}

func Test_CreateStmt_Tablespace(t *testing.T) {
	DoTest(t, DeparseTest{
		Query:    `CREATE TABLE test (id BIGSERIAL PRIMARY KEY, name TEXT) TABLESPACE thing;`,
		Expected: `CREATE TABLE "test" (id bigserial PRIMARY KEY, name text) TABLESPACE "thing"`,
	})
}
