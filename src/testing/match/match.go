// match is a regex matching program similar to 'grep'. However, unlike 'grep',
// it's designed to be used in tests. If its input doesn't match its search
// expression, it prints an error to that effect on stderr. If the input *does*
// match, then it's sent to stdout unchanged, so that multiple 'match' commands
// can be chained.
//
// Note that to make the best use of this behavior, 'match' should be used in a
// bash shell after 'set -e -o pipefail', so that a command chaining several
// 'match' commands exits on the first failure and prints its error message

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var invert = flag.Bool("v", false, "If set, match returns an error if the "+
	"given regex does not match any text on stdin")

func die(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		die("Must provide a regex to match")
	} else if flag.NArg() > 1 {
		die("Currently, match can only accept a single regex")
	}
	re := regexp.MustCompile(flag.Args()[0])

	// Read through input looking for match && copying input to inputBuf (which is
	// written to stdout at the end)
	inputBuf := &strings.Builder{}
	matchFound := false
	s := bufio.NewScanner(io.TeeReader(os.Stdin, inputBuf))
	for s.Scan() {
		switch {
		case matchFound:
			continue // match was found in the past--just copy input to inputBuf
		case !re.Match(s.Bytes()):
			continue // no match--keep copying and looking
		case *invert:
			// Failure: matched with -v
			fmt.Fprintf(
				os.Stderr,
				"match for forbidden regex\n  \"%s\"\nin the text:\n  \"%s\"\n",
				flag.Args()[0],
				s.Text(),
			)
			os.Exit(1)
		default:
			matchFound = true // match found on this line
		}
	}
	if matchFound || *invert {
		// Success--dump inputBuf (with guaranteed newline @ end)
		// Note: we must ignore errors here--os.Stdout might be closed if a
		// downstream cmd has already finished, but we don't want to cause a test to
		// fail because of that
		io.Copy(os.Stdout, strings.NewReader(strings.TrimSuffix(inputBuf.String(), "\n")))
		os.Stdout.Write([]byte{'\n'})
		os.Exit(0)
	} else if !*invert {
		// Failure: match expected, no match found
		fmt.Fprintf(
			os.Stderr,
			"failed to find\n  \"%s\"\nin the text:\n  \"\"\"\n%s\n  \"\"\"\n",
			flag.Args()[0],
			strings.TrimSuffix(inputBuf.String(), "\n"),
		)
		os.Exit(1)
	}
}
