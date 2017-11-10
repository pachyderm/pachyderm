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
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
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

	// Read through input looking for match
	prefix := &bytes.Buffer{}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		prefix.Write(s.Bytes())
		prefix.Write([]byte{'\n'})
		switch {
		case !re.Match(s.Bytes()): // No match, continue
		case *invert: // Failure -- matched with -v
			fmt.Fprintf(
				os.Stderr,
				"did not expect to find\n  \"%s\"\nbut matched\n  \"%s\"\nin the text:\n",
				flag.Args()[0],
				s.Text(),
			)
			io.Copy(os.Stderr, prefix)
			io.Copy(os.Stderr, os.Stdin)
			os.Exit(1)
		default: // Success, match -- copy input to output
			io.Copy(os.Stdout, prefix)
			io.Copy(os.Stdout, os.Stdin)
			os.Exit(0)
		}
	}
	if !*invert {
		// Failure -- no match
		fmt.Fprintf(
			os.Stderr,
			"failed to find\n  \"%s\"\nin the text:\n",
			flag.Args()[0],
		)
		io.Copy(os.Stderr, prefix)
		os.Exit(1)
	}
	// Success -- no match with -v
	io.Copy(os.Stdout, prefix)
}
