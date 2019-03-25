package main

import "github.com/pachyderm/pachyderm/src/client/pfs"

func Filter(inputs []*pfs.FileInfo) bool {
	var path string
	for _, in := range inputs {
		if path == "" {
			path = in.File.Path
		} else if path != in.File.Path {
			return false
		}
	}
	return true
}
