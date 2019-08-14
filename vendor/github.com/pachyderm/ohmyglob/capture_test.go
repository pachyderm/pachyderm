package glob

import (
	"testing"
)

type capTest struct {
	pattern    string
	match      string
	submatches []string
}

func capture(p, m string, d ...string) capTest {
	return capTest{p, m, d}
}

type repTest struct {
	pattern     string
	fixture     string
	template    string
	replacement string
}

func replace(p, f, t, r string) repTest {
	return repTest{p, f, t, r}
}

// tests derived from https://github.com/micromatch/extglob/test
func TestCaptureGlob(t *testing.T) {
	for _, test := range []capTest{
		capture("test/(a|b)", "hi/123"),

		capture("test/(a|b)/x.go", "test/a/x.go", "a"),
		capture("test/(a|b)/x.go", "test/b/x.go", "b"),

		capture("test/a*(a|b)/x.go", "test/a/x.go", ""),
		capture("test/a*(a|b)/x.go", "test/aa/x.go", "a"),
		capture("test/a*(a|b)/x.go", "test/ab/x.go", "b"),
		capture("test/a*(a|b)/x.go", "test/aba/x.go", "ba"),

		capture("test/+(a|b)/x.go", "test/a/x.go", "a"),
		capture("test/+(a|b)/x.go", "test/b/x.go", "b"),
		capture("test/+(a|b)/x.go", "test/ab/x.go", "ab"),
		capture("test/+(a|b)/x.go", "test/aba/x.go", "aba"),

		capture("test/a?(a|b)/x.go", "test/a/x.go", ""),
		capture("test/a?(a|b)/x.go", "test/ab/x.go", "b"),
		capture("test/a?(a|b)/x.go", "test/aa/x.go", "a"),

		capture("test/@(a|b)/x.go", "test/a/x.go", "a"),
		capture("test/@(a|b)/x.go", "test/b/x.go", "b"),

		capture("test/!(a|b)/x.go", "test/x/x.go", "x"),
		capture("test/!(a|b)/x.go", "test/y/x.go", "y"),

		// multi captures
		capture("test/(a|b)/(*).go", "test/a/x.go", "a", "x"),
		capture("test/+(a|b)/(*).go", "test/ab/x.go", "ab", "x"),
		capture("test/@(a|b)/@(*).go", "test/a/x.go", "a", "x"),
		capture("test/a*(a|b)/*(*).go", "test/aaaa/x.go", "aaa", "x"),
		capture("test/a?(a|b)/?(*).go", "test/aa/x.go", "a", "x"),

		// nested captures
		capture("test*(/?(+(a|b)/*.go))", "test/a/x.go", "/a/x.go", "a/x.go", "a"),
	} {
		t.Run("", func(t *testing.T) {
			g, err := Compile(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			results := g.Capture(test.match)

			if len(results) > 0 {
				test.submatches = append([]string{test.match}, test.submatches...)
			}

			ok := true
			if len(results) != len(test.submatches) {
				ok = false
			}
			for i, subgroup := range results {
				if subgroup != test.submatches[i] {
					ok = false
				}
			}
			if !ok {
				t.Errorf("pattern %q matching %q should have captured subgroups %+v, but got %+v\n",
					test.pattern, test.match, test.submatches, results)
			}
		})
	}
}

func TestReplaceGlob(t *testing.T) {
	for _, test := range []repTest{
		replace("name=(*)", "name=adele", "$1", "adele"),
		replace("name=(*)", "n[.ipxi", "$1", ""),
		replace("name=(*)", "name=adele", "${1}1", "adele1"),
		replace("name=(*)", "name=adele", "$$1", "$1"),
		replace("(((((((((??)))))))))*(a|b)", "xfbaaaababbbabaa", "$10", "baaaababbbabaa"),
		replace("name=(*)/year=(*)/(*).dat", "name=adele/year=2016/photos.dat", "$1", "adele"),
		replace("name=(*)/year=(*)/(*).dat", "name=adele/year=2016/photos.dat", "$3 of $1 from $2", "photos of adele from 2016"),
		replace("name=(*)/year=(*)/(*).dat", "name=adele/year=2016/photos.dat", "$0", "name=adele/year=2016/photos.dat"),
		replace("name=!(adele)/year=(2?1(?))/(*).dat", "name=roxanne/year=2514/puzzle.dat", "$1 from the $3th year of the decade, obtained a $4", "roxanne from the 4th year of the decade, obtained a puzzle"),
	} {
		t.Run("", func(t *testing.T) {
			g, err := Compile(test.pattern)
			if err != nil {
				t.Fatal(err)
			}
			replacement := g.Replace(test.fixture, test.template)

			if replacement != test.replacement {
				t.Errorf("replacement template %q matching %q should have returned %+v, but got %+v\n",
					test.template, test.fixture, test.replacement, replacement)
			}
		})
	}
}
