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
	s := bufio.NewScanner(os.Stdin)
	exit := -1
	for s.Scan() {
		os.Stdout.Write(s.Bytes())
		os.Stdout.Write([]byte{'\n'})
		if exit >= 0 {
			continue
		}
		switch {
		case !re.Match(s.Bytes()): // No match, continue
		case *invert: // Failure -- matched with -v
			fmt.Fprintf(
				os.Stderr,
				"did not expect to find\n  \"%s\"\nbut matched\n  \"%s\"\nin the text:\n",
				flag.Args()[0],
				s.Text(),
			)
			exit = 1
		default: // Success, match -- copy input to output
			exit = 0
		}
	}
	if exit >= 0 {
		os.Exit(0)
	}
	if !*invert {
		// Failure -- no match
		fmt.Fprintf(
			os.Stderr,
			"failed to find\n  \"%s\"\nin the text:\n",
			flag.Args()[0],
		)
		os.Exit(1)
	}
	// Success -- no match with -v
}
