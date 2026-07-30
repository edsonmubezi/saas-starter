package pagination

import (
	"fmt"
	"strconv"
	"strings"
)

func BuildSearchFilter(fields []string, paramIndex int, search string) (string, []interface{}) {
	if search == "" || len(fields) == 0 {
		return "", nil
	}

	likeExprs := []string{}
	for _, field := range fields {
		likeExprs = append(likeExprs, field+" ILIKE $"+strconv.Itoa(paramIndex))
	}
	filter := "(" + strings.Join(likeExprs, " OR ") + ")"
	args := []interface{}{"%" + search + "%"}
	return filter, args
}

func BuildSearchFiltertwo(columns []string, startIndex int, search string) (string, []interface{}) {
	var parts []string
	var args []interface{}

	for i, col := range columns {
		// Placeholder position, e.g. $1, $2, ...
		placeholder := fmt.Sprintf("$%d", startIndex+i)
		parts = append(parts, fmt.Sprintf("%s ILIKE %s", col, placeholder))
		args = append(args, "%"+search+"%")
	}

	// Join conditions with OR
	filter := "(" + strings.Join(parts, " OR ") + ")"
	return filter, args
}
