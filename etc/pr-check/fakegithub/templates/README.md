# GitHub Response Templates

This directory contains Go templates (read with [template.ParseFiles()](https://pkg.go.dev/text/template#Template.ParseFiles)) that populate fake GitHub responses for various RPCs.

GitHub responses are typically large JSON objects with a lot of boilerplate; rather than pollute `fake_github.go` in the parent directory with a lot of static data, we keep them here.
