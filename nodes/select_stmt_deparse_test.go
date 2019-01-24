/*
 * Copyright (c) 2019 Ready Stock
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package pg_query

import (
	"testing"
)

func Test_Select_Simple(t *testing.T) {
	DoTest(t, DeparseTest{
		Query:    `SELECT 1;`,
		Expected: `SELECT 1`,
	})
}

func Test_Select_OrderBy(t *testing.T) {
	DoTest(t, DeparseTest{
		Query:    `SELECT id FROM users ORDER BY id DESC;`,
		Expected: `SELECT "id" FROM "users" ORDER BY "id" DESC`,
	})
}

func Test_Select_WeirdBoolExpr(t *testing.T) {
	DoTest(t, DeparseTest{
		Query: `select N.oid::bigint as id, datname as name, D.description
from pg_catalog.pg_database N
  left join pg_catalog.pg_shdescription D on N.oid = D.objoid
where not datistemplate
order by case when datname = pg_catalog.current_database() then -1::bigint else N.oid::bigint end`,
		Expected: `SELECT "n"."oid"::bigint AS id, "datname" AS name, "d"."description" FROM "pg_catalog"."pg_database" n LEFT JOIN "pg_catalog"."pg_shdescription" d ON "n"."oid" = "d"."objoid" WHERE NOT datistemplate ORDER BY CASE WHEN "datname" = current_database() THEN 1::bigint ELSE "n"."oid"::bigint END`,
	})
}