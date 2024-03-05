package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/golangci/golangci-lint/pkg/config"
	"github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/golangci/golangci-lint/pkg/lint/lintersdb"
)

type cliHelp struct {
	Enable  string `json:"enable"`
	Disable string `json:"disable"`
	Help    string `json:"help"`
}

type linterWrapper struct {
	*linter.Config

	Name string `json:"name"`
	Desc string `json:"desc"`
}

func main() {
	linters := lintersdb.NewLinterBuilder().Build(config.NewDefault())

	var wraps []linterWrapper
	for _, l := range linters {
		wraps = append(wraps, linterWrapper{Config: l, Name: l.Linter.Name(), Desc: l.Linter.Desc()})
	}

	err := saveToJSONFile(filepath.Join("assets", "linters-info.json"), wraps)
	if err != nil {
		log.Fatalf("save linters: %v", err)
	}

	err = saveToJSONFile(filepath.Join("assets", "default-exclusions.json"), config.DefaultExcludePatterns)
	if err != nil {
		log.Fatalf("save default exclusions: %v", err)
	}

	err = saveCLIHelp(filepath.Join("assets", "cli-help.json"))
	if err != nil {
		log.Fatalf("save CLI help: %v", err)
	}
}

func saveCLIHelp(dst string) error {
	err := exec.Command("make", "build").Run()
	if err != nil {
		return fmt.Errorf("can't run make build: %w", err)
	}

	lintersOut, err := exec.Command("./golangci-lint", "help", "linters").Output()
	if err != nil {
		return fmt.Errorf("can't run linters cmd: %w", err)
	}

	lintersOutParts := bytes.Split(lintersOut, []byte("\n\n"))

	helpCmd := exec.Command("./golangci-lint", "run", "-h")
	helpCmd.Env = append(helpCmd.Env, os.Environ()...)
	helpCmd.Env = append(helpCmd.Env, "HELP_RUN=1") // make default concurrency stable: don't depend on machine CPU number
	help, err := helpCmd.Output()
	if err != nil {
		return fmt.Errorf("can't run help cmd: %w", err)
	}

	helpLines := bytes.Split(help, []byte("\n"))
	shortHelp := bytes.Join(helpLines[2:], []byte("\n"))

	data := cliHelp{
		Enable:  string(lintersOutParts[0]),
		Disable: string(lintersOutParts[1]),
		Help:    string(shortHelp),
	}

	return saveToJSONFile(dst, data)
}

func saveToJSONFile(dst string, data any) error {
	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("open file (%s): %w", dst, err)
	}

	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("JSON encode (%s): %w", dst, err)
	}

	return nil
}
