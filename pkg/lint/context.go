package lint

import (
	"context"
	"fmt"

	"github.com/golangci/golangci-lint/internal/pkgcache"
	"github.com/golangci/golangci-lint/pkg/config"
	"github.com/golangci/golangci-lint/pkg/exitcodes"
	"github.com/golangci/golangci-lint/pkg/fsutils"
	"github.com/golangci/golangci-lint/pkg/golinters/goanalysis/load"
	"github.com/golangci/golangci-lint/pkg/lint/linter"
	"github.com/golangci/golangci-lint/pkg/logutils"
)

type ContextBuilder struct {
	cfg *config.Config

	pkgLoader *PackageLoader

	lineCache *fsutils.LineCache
	fileCache *fsutils.FileCache
	pkgCache  *pkgcache.Cache

	loadGuard *load.Guard
}

func NewContextBuilder(cfg *config.Config, pkgLoader *PackageLoader,
	lineCache *fsutils.LineCache, fileCache *fsutils.FileCache, pkgCache *pkgcache.Cache, loadGuard *load.Guard,
) *ContextBuilder {
	return &ContextBuilder{
		cfg:       cfg,
		pkgLoader: pkgLoader,
		lineCache: lineCache,
		fileCache: fileCache,
		pkgCache:  pkgCache,
		loadGuard: loadGuard,
	}
}

func (cl *ContextBuilder) Build(ctx context.Context, log logutils.Log, linters []*linter.Config) (*linter.Context, error) {
	pkgs, deduplicatedPkgs, err := cl.pkgLoader.Load(ctx, linters)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if len(deduplicatedPkgs) == 0 {
		return nil, exitcodes.ErrNoGoFiles
	}

	ret := &linter.Context{
		Packages: deduplicatedPkgs,

		// At least `unused` linters works properly only on original (not deduplicated) packages,
		// see https://github.com/golangci/golangci-lint/pull/585.
		OriginalPackages: pkgs,

		Cfg:       cl.cfg,
		Log:       log,
		FileCache: cl.fileCache,
		LineCache: cl.lineCache,
		PkgCache:  cl.pkgCache,
		LoadGuard: cl.loadGuard,
	}

	return ret, nil
}
