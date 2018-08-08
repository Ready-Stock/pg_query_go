package pg_query

import (
	"testing"
	"fmt"
)

func Test_Deparse1(t *testing.T) {
	tree, _ := Parse("SELECT 1")
	if sql, err := Deparse(tree.Statements[0]); err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Println(sql)
	}
}
