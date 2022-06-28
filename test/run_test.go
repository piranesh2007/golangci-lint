package test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/valyala/quicktemplate"

	"github.com/golangci/golangci-lint/pkg/exitcodes"
	"github.com/golangci/golangci-lint/test/testshared"
)

func getCommonRunArgs() []string {
	return []string{"--skip-dirs", "testdata_etc/,pkg/golinters/goanalysis/(checker|passes)"}
}

func withCommonRunArgs(args ...string) []string {
	return append(getCommonRunArgs(), args...)
}

func TestAutogeneratedNoIssues(t *testing.T) {
	testshared.NewLintRunner(t).Run(getTestDataDir("autogenerated")).ExpectNoIssues()
}

func TestEmptyDirRun(t *testing.T) {
	testshared.NewLintRunner(t, "GO111MODULE=off").Run(getTestDataDir("nogofiles")).
		ExpectExitCode(exitcodes.NoGoFiles).
		ExpectOutputContains(": no go files to analyze")
}

func TestNotExistingDirRun(t *testing.T) {
	testshared.NewLintRunner(t, "GO111MODULE=off").Run(getTestDataDir("no_such_dir")).
		ExpectExitCode(exitcodes.Failure).
		ExpectOutputContains("cannot find package").
		ExpectOutputContains("/testdata/no_such_dir")
}

func TestSymlinkLoop(t *testing.T) {
	testshared.NewLintRunner(t).Run(getTestDataDir("symlink_loop", "...")).ExpectNoIssues()
}

func TestDeadline(t *testing.T) {
	testshared.NewLintRunner(t).Run("--deadline=1ms", getProjectRoot()).
		ExpectExitCode(exitcodes.Timeout).
		ExpectOutputContains(`Timeout exceeded: try increasing it by passing --timeout option`)
}

func TestTimeout(t *testing.T) {
	testshared.NewLintRunner(t).Run("--timeout=1ms", getProjectRoot()).
		ExpectExitCode(exitcodes.Timeout).
		ExpectOutputContains(`Timeout exceeded: try increasing it by passing --timeout option`)
}

func TestTimeoutInConfig(t *testing.T) {
	type tc struct {
		cfg string
	}

	cases := []tc{
		{
			cfg: `
				run:
					deadline: 1ms
			`,
		},
		{
			cfg: `
				run:
					timeout: 1ms
			`,
		},
		{
			// timeout should override deadline
			cfg: `
				run:
					deadline: 100s
					timeout: 1ms
			`,
		},
	}

	r := testshared.NewLintRunner(t)
	for _, c := range cases {
		// Run with disallowed option set only in config
		r.RunWithYamlConfig(c.cfg, withCommonRunArgs(minimalPkg)...).ExpectExitCode(exitcodes.Timeout).
			ExpectOutputContains(`Timeout exceeded: try increasing it by passing --timeout option`)
	}
}

func TestTestsAreLintedByDefault(t *testing.T) {
	testshared.NewLintRunner(t).Run(getTestDataDir("withtests")).
		ExpectHasIssue("don't use `init` function")
}

func TestCgoOk(t *testing.T) {
	testshared.NewLintRunner(t).Run("--no-config", "--enable-all", "-D", "nosnakecase", getTestDataDir("cgo")).ExpectNoIssues()
}

func TestCgoWithIssues(t *testing.T) {
	r := testshared.NewLintRunner(t)
	r.Run("--no-config", "--disable-all", "-Egovet", getTestDataDir("cgo_with_issues")).
		ExpectHasIssue("Printf format %t has arg cs of wrong type")
	r.Run("--no-config", "--disable-all", "-Estaticcheck", getTestDataDir("cgo_with_issues")).
		ExpectHasIssue("SA5009: Printf format %t has arg #1 of wrong type")
}

func TestUnsafeOk(t *testing.T) {
	testshared.NewLintRunner(t).Run("--no-config", "--enable-all", getTestDataDir("unsafe")).ExpectNoIssues()
}

func TestGovetCustomFormatter(t *testing.T) {
	testshared.NewLintRunner(t).Run(getTestDataDir("govet_custom_formatter")).ExpectNoIssues()
}

func TestLineDirectiveProcessedFilesLiteLoading(t *testing.T) {
	r := testshared.NewLintRunner(t).Run("--print-issued-lines=false", "--no-config",
		"--exclude-use-default=false", "-Egolint", getTestDataDir("quicktemplate"))

	output := strings.Join([]string{
		"testdata/quicktemplate/hello.qtpl.go:26:1: exported function `StreamHello` should have comment or be unexported (golint)",
		"testdata/quicktemplate/hello.qtpl.go:50:1: exported function `Hello` should have comment or be unexported (golint)",
		"testdata/quicktemplate/hello.qtpl.go:39:1: exported function `WriteHello` should have comment or be unexported (golint)",
	}, "\n")
	r.ExpectExitCode(exitcodes.IssuesFound).ExpectOutputEq(output + "\n")
}

func TestSortedResults(t *testing.T) {
	var testCases = []struct {
		opt  string
		want string
	}{
		{
			"--sort-results=false",
			strings.Join([]string{
				"testdata/sort_results/main.go:12:5: `db` is unused (deadcode)",
				"testdata/sort_results/main.go:15:13: Error return value is not checked (errcheck)",
				"testdata/sort_results/main.go:8:6: func `returnError` is unused (unused)",
			}, "\n"),
		},
		{
			"--sort-results=true",
			strings.Join([]string{
				"testdata/sort_results/main.go:8:6: func `returnError` is unused (unused)",
				"testdata/sort_results/main.go:12:5: `db` is unused (deadcode)",
				"testdata/sort_results/main.go:15:13: Error return value is not checked (errcheck)",
			}, "\n"),
		},
	}

	dir := getTestDataDir("sort_results")

	t.Parallel()
	for i := range testCases {
		test := testCases[i]
		t.Run(test.opt, func(t *testing.T) {
			r := testshared.NewLintRunner(t).Run("--print-issued-lines=false", test.opt, dir)
			r.ExpectExitCode(exitcodes.IssuesFound).ExpectOutputEq(test.want + "\n")
		})
	}
}

func TestLineDirectiveProcessedFilesFullLoading(t *testing.T) {
	r := testshared.NewLintRunner(t).Run("--print-issued-lines=false", "--no-config",
		"--exclude-use-default=false", "-Egolint,govet", getTestDataDir("quicktemplate"))

	output := strings.Join([]string{
		"testdata/quicktemplate/hello.qtpl.go:26:1: exported function `StreamHello` should have comment or be unexported (golint)",
		"testdata/quicktemplate/hello.qtpl.go:50:1: exported function `Hello` should have comment or be unexported (golint)",
		"testdata/quicktemplate/hello.qtpl.go:39:1: exported function `WriteHello` should have comment or be unexported (golint)",
	}, "\n")
	r.ExpectExitCode(exitcodes.IssuesFound).ExpectOutputEq(output + "\n")
}

func TestLintFilesWithLineDirective(t *testing.T) {
	r := testshared.NewLintRunner(t)
	r.Run("-Edupl", "--disable-all", "--config=testdata/linedirective/dupl.yml", getTestDataDir("linedirective")).
		ExpectHasIssue("21-23 lines are duplicate of `testdata/linedirective/hello.go:25-27` (dupl)")
	r.Run("-Egofmt", "--disable-all", "--no-config", getTestDataDir("linedirective")).
		ExpectHasIssue("File is not `gofmt`-ed with `-s` (gofmt)")
	r.Run("-Egoimports", "--disable-all", "--no-config", getTestDataDir("linedirective")).
		ExpectHasIssue("File is not `goimports`-ed (goimports)")
	r.
		Run("-Egomodguard", "--disable-all", "--config=testdata/linedirective/gomodguard.yml", getTestDataDir("linedirective")).
		ExpectHasIssue("import of package `github.com/ryancurrah/gomodguard` is blocked because the module is not " +
			"in the allowed modules list. (gomodguard)")
	r.Run("-Elll", "--disable-all", "--config=testdata/linedirective/lll.yml", getTestDataDir("linedirective")).
		ExpectHasIssue("line is 57 characters (lll)")
	r.Run("-Emisspell", "--disable-all", "--no-config", getTestDataDir("linedirective")).
		ExpectHasIssue("is a misspelling of `language` (misspell)")
	r.Run("-Ewsl", "--disable-all", "--no-config", getTestDataDir("linedirective")).
		ExpectHasIssue("block should not start with a whitespace (wsl)")
}

func TestSkippedDirsNoMatchArg(t *testing.T) {
	dir := getTestDataDir("skipdirs", "skip_me", "nested")
	res := testshared.NewLintRunner(t).Run("--print-issued-lines=false", "--no-config", "--skip-dirs", dir, "-Egolint", dir)

	res.ExpectExitCode(exitcodes.IssuesFound).
		ExpectOutputEq("testdata/skipdirs/skip_me/nested/with_issue.go:8:9: `if` block ends with " +
			"a `return` statement, so drop this `else` and outdent its block (golint)\n")
}

func TestSkippedDirsTestdata(t *testing.T) {
	r := testshared.NewLintRunner(t).Run("--print-issued-lines=false", "--no-config", "-Egolint", getTestDataDir("skipdirs", "..."))

	r.ExpectNoIssues() // all was skipped because in testdata
}

func TestDeadcodeNoFalsePositivesInMainPkg(t *testing.T) {
	testshared.NewLintRunner(t).Run("--no-config", "--disable-all", "-Edeadcode", getTestDataDir("deadcode_main_pkg")).ExpectNoIssues()
}

func TestIdentifierUsedOnlyInTests(t *testing.T) {
	testshared.NewLintRunner(t).Run("--no-config", "--disable-all", "-Eunused", getTestDataDir("used_only_in_tests")).ExpectNoIssues()
}

func TestUnusedCheckExported(t *testing.T) {
	t.Skip("Issue955")
	testshared.NewLintRunner(t).Run("-c", "testdata_etc/unused_exported/golangci.yml", "testdata_etc/unused_exported/...").ExpectNoIssues()
}

func TestConfigFileIsDetected(t *testing.T) {
	checkGotConfig := func(r *testshared.RunResult) {
		r.ExpectExitCode(exitcodes.Success).
			ExpectOutputEq("test\n") // test config contains InternalTest: true, it triggers such output
	}

	r := testshared.NewLintRunner(t)
	checkGotConfig(r.Run(getTestDataDir("withconfig", "pkg")))
	checkGotConfig(r.Run(getTestDataDir("withconfig", "...")))
}

func TestEnableAllFastAndEnableCanCoexist(t *testing.T) {
	r := testshared.NewLintRunner(t)
	r.Run(withCommonRunArgs("--no-config", "--fast", "--enable-all", "--enable=typecheck", minimalPkg)...).
		ExpectExitCode(exitcodes.Success, exitcodes.IssuesFound)
	r.Run(withCommonRunArgs("--no-config", "--enable-all", "--enable=typecheck", minimalPkg)...).
		ExpectExitCode(exitcodes.Failure)
}

func TestEnabledPresetsAreNotDuplicated(t *testing.T) {
	testshared.NewLintRunner(t).Run("--no-config", "-v", "-p", "style,bugs", minimalPkg).
		ExpectOutputContains("Active presets: [bugs style]")
}

func TestAbsPathDirAnalysis(t *testing.T) {
	dir := filepath.Join("testdata_etc", "abspath") // abs paths don't work with testdata dir
	absDir, err := filepath.Abs(dir)
	assert.NoError(t, err)

	r := testshared.NewLintRunner(t).Run("--print-issued-lines=false", "--no-config", "-Egolint", absDir)
	r.ExpectHasIssue("`if` block ends with a `return` statement")
}

func TestAbsPathFileAnalysis(t *testing.T) {
	dir := filepath.Join("testdata_etc", "abspath", "with_issue.go") // abs paths don't work with testdata dir
	absDir, err := filepath.Abs(dir)
	assert.NoError(t, err)

	r := testshared.NewLintRunner(t).Run("--print-issued-lines=false", "--no-config", "-Egolint", absDir)
	r.ExpectHasIssue("`if` block ends with a `return` statement")
}

func TestDisallowedOptionsInConfig(t *testing.T) {
	type tc struct {
		cfg    string
		option string
	}

	cases := []tc{
		{
			cfg: `
				ruN:
					Args:
						- 1
			`,
		},
		{
			cfg: `
				run:
					CPUProfilePath: path
			`,
			option: "--cpu-profile-path=path",
		},
		{
			cfg: `
				run:
					MemProfilePath: path
			`,
			option: "--mem-profile-path=path",
		},
		{
			cfg: `
				run:
					TracePath: path
			`,
			option: "--trace-path=path",
		},
		{
			cfg: `
				run:
					Verbose: true
			`,
			option: "-v",
		},
	}

	r := testshared.NewLintRunner(t)
	for _, c := range cases {
		// Run with disallowed option set only in config
		r.RunWithYamlConfig(c.cfg, withCommonRunArgs(minimalPkg)...).ExpectExitCode(exitcodes.Failure)

		if c.option == "" {
			continue
		}

		args := []string{c.option, "--fast", minimalPkg}

		// Run with disallowed option set only in command-line
		r.Run(withCommonRunArgs(args...)...).ExpectExitCode(exitcodes.Success)

		// Run with disallowed option set both in command-line and in config
		r.RunWithYamlConfig(c.cfg, withCommonRunArgs(args...)...).ExpectExitCode(exitcodes.Failure)
	}
}

func TestPathPrefix(t *testing.T) {
	for _, tt := range []struct {
		Name    string
		Args    []string
		Pattern string
	}{
		{"empty", nil, "^testdata/withtests/"},
		{"prefixed", []string{"--path-prefix=cool"}, "^cool/testdata/withtests"},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			testshared.NewLintRunner(t).Run(
				append(tt.Args, getTestDataDir("withtests"))...,
			).ExpectOutputRegexp(
				tt.Pattern,
			)
		})
	}
}
