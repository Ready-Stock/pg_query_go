package pg_query

//
// func Test_ParseRegress(t *testing.T) {
// 	files, err := ioutil.ReadDir("./regress")
// 	if err != nil {
// 		t.Error(err)
// 		t.Fail()
// 		return
// 	}
// 	for _, file := range files{
// 		path := fmt.Sprintf("./regress/%s", file.Name())
// 		fmt.Println(fmt.Sprintf("Opening file: %s", path))
// 		d, err := ioutil.ReadFile(path)
// 		if err != nil {
// 			t.Error(err)
// 			t.Fail()
// 			return
// 		}
// 		sql := string(d)
// 		queries := strings.Split(sql, ";\n")
// 		for _, query := range queries {
// 			query = strings.TrimSpace(query)
// 			if query == "" {
// 				continue
// 			}
// 			fmt.Println(fmt.Sprintf("\n====== Parsing ====== File: %s", path))
// 			fmt.Println(fmt.Sprintf("%s", query))
// 			_, err = Parse(query)
// 			if err != nil {
// 				if strings.Contains(query, "-- error") {
// 					// Errors out properly.
// 				} else {
// 					t.Errorf("Could not parse: %s | %s",query, err.Error())
// 					t.Fail()
// 					return
// 				}
// 			}
// 			fmt.Println("=====================")
// 		}
//
// 	}
// }
