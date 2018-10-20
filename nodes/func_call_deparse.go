// Auto-generated - DO NOT EDIT

package pg_query

import (
	"fmt"
	"strings"
)

func (node FuncCall) Deparse(ctx Context) (*string, error) {
	out := make([]string, 0)

	args := make([]string, len(node.Args.Items))
	args, err := deparseNodeList(node.Args.Items, Context_None)
	if err != nil {
		return nil, err
	}

	if node.AggStar {
		args = append(args, "*")
	}

	difference := func(slice1 []string, slice2 []string) []string {
		var diff []string
		// Loop two times, first to find slice1 strings not in slice2,
		// second loop to find slice2 strings not in slice1
		for i := 0; i < 2; i++ {
			for _, s1 := range slice1 {
				found := false
				for _, s2 := range slice2 {
					if s1 == s2 {
						found = true
						break
					}
				}
				// String not found. We add it to return slice
				if !found {
					diff = append(diff, s1)
				}
			}
			// Swap the slices, only if it was the first loop
			if i == 0 {
				slice1, slice2 = slice2, slice1
			}
		}
		return diff
	}

	name := ""
	if names, err := deparseNodeList(node.Funcname.Items, Context_FuncCall); err != nil {
		return nil, err
	} else {
		name = strings.Join(difference([]string{"pg_catalog"}, names), ".")
	}

	distinct := ""
	if node.AggDistinct {
		distinct = "DISTINCT "
	}

	out = append(out, fmt.Sprintf("%s(%s%s)", name, distinct, strings.Join(args, ", ")))

	if node.Over != nil {
		if over, err := deparseNode(node.Over, Context_None); err != nil {
			return nil, err
		} else {
			out = append(out, fmt.Sprintf("OVER (%s)", *over))
		}
	}

	result := strings.Join(out, " ")
	return &result, nil
}
