package processors

import (
	"fmt"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/golangci/golangci-lint/pkg/result"
)

func TestAutogeneratedExclude_isGeneratedFileLax_generated(t *testing.T) {
	p := NewAutogeneratedExclude(AutogeneratedModeLax)

	comments := []string{
		`	// generated by stringer -type Pill pill.go; DO NOT EDIT`,
		`// Code generated by "stringer -type Pill pill.go"; DO NOT EDIT`,
		`// Code generated by vfsgen; DO NOT EDIT`,
		`// Created by cgo -godefs - DO NOT EDIT`,
		`/* Created by cgo - DO NOT EDIT. */`,
		`// Generated by stringer -i a.out.go -o anames.go -p ppc64
// Do not edit.`,
		`// DO NOT EDIT
// generated by: x86map -fmt=decoder ../x86.csv`,
		`// DO NOT EDIT.
// Generate with: go run gen.go -full -output md5block.go`,
		`// generated by "go run gen.go". DO NOT EDIT.`,
		`// DO NOT EDIT. This file is generated by mksyntaxgo from the RE2 distribution.`,
		`// GENERATED BY make_perl_groups.pl; DO NOT EDIT.`,
		`// generated by mknacl.sh - do not edit`,
		`// DO NOT EDIT ** This file was generated with the bake tool ** DO NOT EDIT //`,
		`// Generated by running
//  maketables --tables=all --data=http://www.unicode.org/Public/8.0.0/ucd/UnicodeData.txt
// --casefolding=http://www.unicode.org/Public/8.0.0/ucd/CaseFolding.txt
// DO NOT EDIT`,
		`/*
* CODE GENERATED AUTOMATICALLY WITH github.com/ernesto-jimenez/gogen/unmarshalmap
* THIS FILE SHOULD NOT BE EDITED BY HAND
  */`,
		`// AUTOGENERATED FILE: easyjson file.go`,
	}

	for _, comment := range comments {
		comment := comment
		t.Run(comment, func(t *testing.T) {
			t.Parallel()

			generated := p.isGeneratedFileLax(comment)
			assert.True(t, generated)
		})
	}
}

func TestAutogeneratedExclude_isGeneratedFileLax_nonGenerated(t *testing.T) {
	p := NewAutogeneratedExclude(AutogeneratedModeLax)

	comments := []string{
		"code not generated by",
		"test",
	}

	for _, comment := range comments {
		comment := comment
		t.Run(comment, func(t *testing.T) {
			t.Parallel()

			generated := p.isGeneratedFileLax(comment)
			assert.False(t, generated)
		})
	}
}

func TestAutogeneratedExclude_isGeneratedFileStrict(t *testing.T) {
	p := NewAutogeneratedExclude(AutogeneratedModeStrict)

	testCases := []struct {
		desc     string
		filepath string
		assert   assert.BoolAssertionFunc
	}{
		{
			desc:     "",
			filepath: filepath.FromSlash("testdata/autogen_go_strict.go"),
			assert:   assert.True,
		},
		{
			desc:     "",
			filepath: filepath.FromSlash("testdata/autogen_go_strict_invalid.go"),
			assert:   assert.False,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			generated, err := p.isGeneratedFileStrict(test.filepath)
			require.NoError(t, err)

			test.assert(t, generated)
		})
	}
}

func Test_getComments(t *testing.T) {
	testCases := []struct {
		fpath string
		doc   string
	}{
		{
			fpath: filepath.FromSlash("testdata/autogen_exclude.go"),
			doc: `first line
second line
third line
this text also
and this text also`,
		},
		{
			fpath: filepath.FromSlash("testdata/autogen_exclude_doc.go"),
			doc:   `DO NOT EDIT`,
		},
		{
			fpath: filepath.FromSlash("testdata/autogen_exclude_block_comment.go"),
			doc: `* first line
 *
 * second line
 * third line
and this text also
this type of block comment also
this one line comment also`,
		},
	}

	for _, tc := range testCases {
		doc, err := getComments(tc.fpath)
		require.NoError(t, err)
		assert.Equal(t, tc.doc, doc)
	}
}

// Issue 954: Some lines can be very long, e.g. auto-generated
// embedded resources. Reported on file of 86.2KB.
func Test_getComments_fileWithLongLine(t *testing.T) {
	fpath := filepath.FromSlash("testdata/autogen_exclude_long_line.go")
	_, err := getComments(fpath)
	assert.NoError(t, err)
}

func Test_shouldPassIssue(t *testing.T) {
	testCases := []struct {
		desc   string
		mode   string
		issue  *result.Issue
		assert assert.BoolAssertionFunc
	}{
		{
			desc: "typecheck issue",
			mode: AutogeneratedModeLax,
			issue: &result.Issue{
				FromLinter: "typecheck",
			},
			assert: assert.True,
		},
		{
			desc: "lax ",
			mode: AutogeneratedModeLax,
			issue: &result.Issue{
				FromLinter: "example",
				Pos: token.Position{
					Filename: filepath.FromSlash("testdata/autogen_go_strict_invalid.go"),
				},
			},
			assert: assert.False,
		},
		{
			desc: "strict ",
			mode: AutogeneratedModeStrict,
			issue: &result.Issue{
				FromLinter: "example",
				Pos: token.Position{
					Filename: filepath.FromSlash("testdata/autogen_go_strict_invalid.go"),
				},
			},
			assert: assert.True,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := NewAutogeneratedExclude(test.mode)

			pass, err := p.shouldPassIssue(test.issue)
			require.NoError(t, err)

			test.assert(t, pass)
		})
	}
}

func Test_shouldPassIssue_error(t *testing.T) {
	notFoundMsg := "no such file or directory"
	if runtime.GOOS == "windows" {
		notFoundMsg = "The system cannot find the file specified."
	}

	testCases := []struct {
		desc     string
		mode     string
		issue    *result.Issue
		expected string
	}{
		{
			desc: "non-existing file (lax)",
			mode: AutogeneratedModeLax,
			issue: &result.Issue{
				FromLinter: "example",
				Pos: token.Position{
					Filename: filepath.FromSlash("no-existing.go"),
				},
			},
			expected: fmt.Sprintf("failed to get doc (lax) of file %[1]s: failed to parse file: open %[1]s: %[2]s",
				filepath.FromSlash("no-existing.go"), notFoundMsg),
		},
		{
			desc: "non-existing file (strict)",
			mode: AutogeneratedModeStrict,
			issue: &result.Issue{
				FromLinter: "example",
				Pos: token.Position{
					Filename: filepath.FromSlash("no-existing.go"),
				},
			},
			expected: fmt.Sprintf("failed to get doc (strict) of file %[1]s: failed to parse file: open %[1]s: %[2]s",
				filepath.FromSlash("no-existing.go"), notFoundMsg),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := NewAutogeneratedExclude(test.mode)

			pass, err := p.shouldPassIssue(test.issue)

			//nolint:testifylint // It's a loop and the main expectation is the error message.
			assert.EqualError(t, err, test.expected)
			assert.False(t, pass)
		})
	}
}
