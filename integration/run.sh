#!/bin/sh
# Integration checks for yup-dirname, run inside a Debian (GNU coreutils)
# container.
#
# parity CASE  — yup-dirname must produce byte-identical output to GNU `dirname`.
# assert WANT  — yup-dirname must produce WANT exactly (used where yup-dirname
#                diverges from GNU by design; see cmd-dirname COMPATIBILITY.md).
set -eu

fails=0

parity() {
	ours=$(yup-dirname "$@" 2>/dev/null || true)
	gnu=$(dirname "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  dirname %s\n' "$*"
	else
		printf 'FAIL  parity  dirname %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	shift
	got=$(yup-dirname "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  dirname %s\n' "$*"
	else
		printf 'FAIL  assert  dirname %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Absolute paths: drop the last component.
parity /usr/local/bin/script.sh
parity /usr/bin
parity /a/b/c

# Relative paths and the no-separator case (output '.').
parity filename
parity dir/file
parity ./file

# Trailing slashes are stripped before and after dropping the last component.
parity /usr/local/bin/
parity /usr/local/bin///
parity dir/

# Root and edge cases.
parity /
parity ///
parity .
parity ..

# Interior separators are preserved; "." / ".." are not resolved.
parity a//b//c
parity /foo/..

# Multiple operands: one output line per NAME (matches GNU `dirname NAME...`).
parity /usr/local/bin/script.sh /var/log/app.log
parity a/b c/d e

# Documented divergence: GNU `dirname` requires at least one operand and errors
# with no arguments. yup-dirname instead reads newline-separated paths from
# stdin when given no operands, applying dirname to each line.
got_stdin=$(printf '/usr/local/bin/script.sh\n/etc/nginx/nginx.conf\nfilename\n' | yup-dirname 2>/dev/null || true)
want_stdin=$(printf '/usr/local/bin\n/etc/nginx\n.')
if [ "$got_stdin" = "$want_stdin" ]; then
	printf 'ok    assert  dirname <stdin>\n'
else
	printf 'FAIL  assert  dirname <stdin>\n        want: %s\n        got:  %s\n' "$want_stdin" "$got_stdin"
	fails=$((fails + 1))
fi

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
