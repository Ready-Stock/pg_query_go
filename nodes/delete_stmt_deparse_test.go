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

func Test_DeleteStmt_Generic(t *testing.T) {
	DoTest(t, DeparseTest{
		Query:    `delete from thing;`,
		Expected: `DELETE FROM "thing"`,
	})
	DoTest(t, DeparseTest{
		Query:    `delete from thing where accountId = 123;`,
		Expected: `DELETE FROM "thing" WHERE "accountid" = 123`,
	})
}

func Test_DeleteStmt_Returning(t *testing.T) {
	DoTest(t, DeparseTest{
		Query:    `delete from thing returning *;`,
		Expected: `DELETE FROM "thing" RETURNING *`,
	})
	DoTest(t, DeparseTest{
		Query:    `delete from thing where accountId = 123 returning accountId;`,
		Expected: `DELETE FROM "thing" WHERE "accountid" = 123 RETURNING "accountid"`,
	})
}
