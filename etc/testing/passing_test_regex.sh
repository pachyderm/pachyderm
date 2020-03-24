#!/bin/bash
#
# This script prints a regex (which should be passed to 'go test -run=<regex>')
# that matches all tests in the pachyderm test suite that aren't in
# etc/testing/blacklist.txt. Tests in blacklist.txt are flaky or otherwise
# cause our test suite to consitently flake or fail when run in a cloud
# provider's cluster and would render our nightly tests useless.

basedir="$(realpath "$(dirname "${0}")/../..")"
tmpfile="$(mktemp --dry-run --tmpdir=.)"

# Put blacklisted tests in ${tmpfile}
grep -v '^$' "${basedir}/etc/testing/blacklist.txt" >"${tmpfile}"

# List all tests in src/server, and put them in ${tmpfile} as well
ls "${basedir}/src/server" \
  | grep '_test\.go$' \
  | while read -r f; do
      grep --no-filename '^func Test[A-Za-z_]\+(.* \*testing\.T)' "${basedir}/src/server/${f}" \
        | sed 's/^func \(Test[A-Za-z_]\+\)(.*$/\1/' >>"${tmpfile}"
    done

# List tests that haven't been blacklisted. These tests only appear once
# in ${tmpfile}, which includes both the list of all tests and the blacklist,
# so that blacklisted tests appear twice. Format the result as a regex with awk
sort "${tmpfile}" | uniq -u | awk '
   BEGIN { first = 1 }
   first == 0 { pattern = pattern "|" $0 }
   first == 1 { pattern = $0; first = 0 }
   END { print pattern; }'
rm "${tmpfile}"
