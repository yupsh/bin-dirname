package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		stdin      string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			name:    "stdin single path",
			args:    []string{"dirname"},
			stdin:   "/usr/local/bin/script.sh\n",
			wantOut: "/usr/local/bin\n",
		},
		{
			name:    "stdin multiple paths",
			args:    []string{"dirname"},
			stdin:   "/usr/bin/ls\n/etc/nginx/nginx.conf\n",
			wantOut: "/usr/bin\n/etc/nginx\n",
		},
		{
			name:    "arg single path",
			args:    []string{"dirname", "/a/b/c"},
			wantOut: "/a/b\n",
		},
		{
			name:    "arg multiple paths",
			args:    []string{"dirname", "/usr/local/bin/script.sh", "/var/log/app.log"},
			wantOut: "/usr/local/bin\n/var/log\n",
		},
		{
			name:    "arg filename only",
			args:    []string{"dirname", "filename"},
			wantOut: ".\n",
		},
		{
			name:    "arg root",
			args:    []string{"dirname", "/"},
			wantOut: "/\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"dirname", "--version"},
			wantOut: "dirname version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"dirname", "--nope"},
			wantCode:   1,
			wantErrSub: "dirname:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
