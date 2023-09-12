package pfsdb

import (
	"fmt"
	"testing"
)

func TestListResourceConfig(t *testing.T) {
	config := ListResourceConfig{
		BaseQuery: getBranchBaseQuery,
		AndFilters: []Filter{
			{Field: "branch.repo_name", Value: "foo", Function: ValueIn},
		},
	}
	fmt.Println(config.Query())
}
