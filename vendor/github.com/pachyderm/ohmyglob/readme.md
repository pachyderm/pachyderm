# Oh my glob! Go capture globbing library
<img align="right" width="220" height="220" src="https://i.imgur.com/kQoBuaZ.png">

This library allows you to use capture groups in globs
by using the extended globbing functions.

This is implemented by compiling the glob patterns to regex,
and then doing the matching and capturing with the [Regexp2 library](https://github.com/dlclark/regexp2).

The parser, lexer, and general structure for this library are derived from the excellent https://github.com/gobwas/glob library.

## Install

```shell
    go get github.com/pachyderm/ohmyglob
```

## Example

```go

package main

import "github.com/pachyderm/ohmyglob"

func main() {
    var g glob.Glob

// create simple glob
    g = glob.MustCompile("*.github.com")
    g.Match("api.github.com") // true

// quote meta characters and then create simple glob
    g = glob.MustCompile(glob.QuoteMeta("*.github.com"))
    g.Match("*.github.com") // true

// create new glob with set of delimiters as ["."]
    g = glob.MustCompile("api.*.com", '.')
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // false

// create new glob with set of delimiters as ["."]
    // but now with super wildcard
    g = glob.MustCompile("api.**.com", '.')
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // true

    // create glob with single symbol wildcard
    g = glob.MustCompile("?at")
    g.Match("cat") // true
    g.Match("fat") // true
    g.Match("at") // false

// create glob with single symbol wildcard and delimiters ['f']
    g = glob.MustCompile("?at", 'f')
    g.Match("cat") // true
    g.Match("fat") // false
    g.Match("at") // false

// create glob with character-list matchers
    g = glob.MustCompile("[abc]at")
    g.Match("cat") // true
    g.Match("bat") // true
    g.Match("fat") // false
    g.Match("at") // false

// create glob with character-list matchers
    g = glob.MustCompile("[!abc]at")
    g.Match("cat") // false
    g.Match("bat") // false
    g.Match("fat") // true
    g.Match("at") // false

// create glob with character-range matchers
    g = glob.MustCompile("[a-c]at")
    g.Match("cat") // true
    g.Match("bat") // true
    g.Match("fat") // false
    g.Match("at") // false

// create glob with character-range matchers
    g = glob.MustCompile("[!a-c]at")
    g.Match("cat") // false
    g.Match("bat") // false
    g.Match("fat") // true
    g.Match("at") // false

// create glob with pattern-alternatives list
    g = glob.MustCompile("{cat,bat,[fr]at}")
    g.Match("cat") // true
    g.Match("bat") // true
    g.Match("fat") // true
    g.Match("rat") // true
    g.Match("at") // false
    g.Match("zat") // false

// create glob with extended glob patterns
    g = glob.MustCompile("@(a)")
    g.Match("") // false
    g.Match("a") // true
    g.Match("aa") // false
    g.Match("aaa") // false
    g.Match("aaX") // false
    g.Match("bbb") // false

    g = glob.MustCompile("*(a)")
    g.Match("") // true
    g.Match("a") // true
    g.Match("aa") // true
    g.Match("aaa") // true
    g.Match("aaX") // false
    g.Match("bbb") // false

    g = glob.MustCompile("+(a)")
    g.Match("") // false
    g.Match("a") // true
    g.Match("aa") // true
    g.Match("aaa") // true
    g.Match("aaX") // false
    g.Match("bbb") // false

    g = glob.MustCompile("?(a)")
    g.Match("") // true
    g.Match("a") // true
    g.Match("aa") // false
    g.Match("aaa") // false
    g.Match("aaX") // false
    g.Match("bbb") // false

    g = glob.MustCompile("!(a)")
    g.Match("") // true
    g.Match("a") // false
    g.Match("aa") // true
    g.Match("aaa") // true
    g.Match("aaX") // true
    g.Match("bbb") // true

// create a glob and get the capture groups
    g = glob.MustCompile("test/a*(a|b)/*(*).go")
    g.Capture("test/aaaa/x.go") // ["test/aaaa/x.go", "aaa", "x"]

}

```
