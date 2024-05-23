package starlark

import "go.starlark.net/starlark"

func Union(dicts ...starlark.StringDict) starlark.StringDict {
	result := make(starlark.StringDict)
	for _, d := range dicts {
		for k, v := range d {
			result[k] = v
		}
	}
	return result
}
