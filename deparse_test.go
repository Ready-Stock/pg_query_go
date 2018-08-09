package pg_query

import (
	"testing"
	"fmt"
)

func Test_Deparse1(t *testing.T) {
	input := "SELECT 1"
	fmt.Printf("INPUT: %s\n", input)
	tree, _ := Parse(input)
	json, _ := tree.MarshalJSON()
	fmt.Println(string(json))
	if sql, err := Deparse(tree.Statements[0]); err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Printf("OUTPUT: %s\n", *sql)
	}
}

func Test_Deparse2(t *testing.T) {
	input := "SELECT test FROM users;"
	fmt.Printf("INPUT: %s\n", input)
	tree, _ := Parse(input)
	json, _ := tree.MarshalJSON()
	fmt.Println(string(json))
	if sql, err := Deparse(tree.Statements[0]); err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Printf("OUTPUT: %s\n", *sql)
	}
}

func Test_DeparseCurrentTimestamp(t *testing.T) {
	input := "select    current_timestamp"
	fmt.Printf("INPUT: %s\n", input)
	tree, _ := Parse(input)
	json, _ := tree.MarshalJSON()
	fmt.Println(string(json))
	if sql, err := Deparse(tree.Statements[0]); err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Printf("OUTPUT: %s\n", *sql)
	}
}
