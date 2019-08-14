# Go Capture Globbing Library

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

## Performance

This library is created for compile-once patterns. This means that the compilation could take time, but
string matching is done faster, compared to the case when the template is compiled each time.

Since it uses the Regexp2 library to do the matching and capturing, it performs about on par with regular expressions. 
If you need something faster, and don't need capture groups, we recommend https://github.com/gobwas/glob.

## Syntax

Syntax is inspired by [standard wildcards](http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm),
except that `**` is a super-asterisk, which is not sensitive to separators.
