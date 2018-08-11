package pg_query

import (
	"testing"
	"fmt"
)

var (
	queries = []struct{
		Query string
		Result string
	}{
		{
			"select 1",
			"SELECT 1;",
		},
		{
			"select users.user_id from users",
			`SELECT "users"."user_id" FROM "users";`,
		},
		{
			"insert into users (user_id,email) values(1, 'email@email.com') RETURNING *",
			`INSERT INTO "users" (user_id,email) VALUES (1, 'email@email.com') RETURNING *;`,
		},
		{
			`select t.oid,
						case when nsp.nspname in ('pg_catalog', 'public') then t.typname
							else nsp.nspname||'.'||t.typname
						end
					from pg_type t
					left join pg_type base_type on t.typelem=base_type.oid
					left join pg_namespace nsp on t.typnamespace=nsp.oid
					where (
						  (t.typtype in('b', 'p', 'r', 'e') or 1=1)
						  and (base_type.oid is null or base_type.typtype in('b', 'p', 'r'))
						);`,
			`SELECT "t"."oid", CASE WHEN "nsp"."nspname" IN ('pg_catalog', 'public') THEN "t"."typname" ELSE "nsp"."nspname" || '.' || "t"."typname" END FROM "pg_type" t LEFT JOIN "pg_type" base_type ON "t"."typelem" = "base_type"."oid" LEFT JOIN "pg_namespace" nsp ON "t"."typnamespace" = "nsp"."oid" WHERE ("t"."typtype" IN ('b', 'p', 'r', 'e') OR 1 = 1) AND ("base_type"."oid" IS NULL OR "base_type"."typtype" IN ('b', 'p', 'r'));`,
		},
		// {
		// 	"UPDATE users set is_enabled=true WHERE user_id='2' returning *",
		// 	`UPDATE "users" SET is_enabled = true WHERE "user_id" = '2' RETURNING *;`,
		// },
	}
)

func Test_Deparse(t *testing.T) {
	for _, test := range queries {
		if tree, err := Parse(test.Query); err != nil {
			t.Errorf(`FAILED TO PARSE QUERY: {%s} ERROR: %s`, test.Query, err.Error())
			t.Fail()
			return
		} else {
			if sql, err := Deparse(tree.Statements[0]); err != nil {
				t.Errorf("FAILED TO DEPARSE QUERY: {%s} ERROR: %s", test.Query, err.Error())
				t.Fail()
				return
			} else {
				if *sql != test.Result {
					t.Errorf("ERROR, QUERY {%s} DID NOT DEPARSE INTO {%s}", test.Query, test.Result)
				} else {
					t.Logf("SUCCESS, INPUT: {%s} OUTPUT: {%s}", test.Query, *sql)
				}
			}
		}
	}
}





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


func Test_DeparseInsert1(t *testing.T) {
	input := "insert into public.users (email, password) values ('email@google.com', 'strongpassword') returning *"
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

func Test_DeparseBigSelect(t *testing.T) {
	input := `
		select t.oid,
			case when nsp.nspname in ('pg_catalog', 'public') then t.typname
				else nsp.nspname||'.'||t.typname
			end
		from pg_type t
		left join pg_type base_type on t.typelem=base_type.oid
		left join pg_namespace nsp on t.typnamespace=nsp.oid
		where (
			  (t.typtype in('b', 'p', 'r', 'e') or 1=1)
			  and (base_type.oid is null or base_type.typtype in('b', 'p', 'r'))
			);
		`
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

// func Test_DeparseUpdate(t *testing.T) {
// 	input := `UPDATE users set is_enabled=true WHERE user_id='2' returning *`
// 	fmt.Printf("INPUT: %s\n", input)
// 	tree, _ := Parse(input)
// 	json, _ := tree.MarshalJSON()
// 	fmt.Println(string(json))
// 	if sql, err := Deparse(tree.Statements[0]); err != nil {
// 		t.Error(err)
// 		t.Fail()
// 	} else {
// 		fmt.Printf("OUTPUT: %s\n", *sql)
// 	}
// }



