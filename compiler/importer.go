package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PackageInfo holds metadata about a compiled package
type PackageInfo struct {
	Name          string      // The namespace name (e.g., "utils")
	SourcePath    string      // Absolute path to directory
	Namespace     *Namespace  // The symbol table for this package
	IsProcessing  bool        // To detect circular imports
}

// Importer handles resolving and loading imports
type Importer struct {
	entryDir string                  // Directory of the entry point file
	cache    map[string]*PackageInfo // Path -> Package
}

// NewImporter creates a new importer based on the entry file location
func NewImporter(entryFile string) *Importer {
	absPath, _ := filepath.Abs(entryFile)
	return &Importer{
		entryDir: filepath.Dir(absPath),
		cache:    make(map[string]*PackageInfo),
	}
}

// ResolvePath converts an import string to an absolute directory path
func (imp *Importer) ResolvePath(currentFileDir, importPath string) (string, error) {
	// Handle local relative imports (starting with ./ or ../)
	if strings.HasPrefix(importPath, ".") {
		if currentFileDir == "" {
			currentFileDir = imp.entryDir
		}
		return filepath.Abs(filepath.Join(currentFileDir, importPath))
	}

	// TODO: Handle standard library and module imports
	// For now, treat non-relative imports as relative to entry directory or vendor
	// This is a placeholder for a real module resolution strategy
	return filepath.Abs(filepath.Join(imp.entryDir, importPath))
}

// GetSourceFiles returns all .arc files in a directory
func (imp *Importer) GetSourceFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".arc") || strings.HasSuffix(entry.Name(), ".lang")) {
			files = append(files, filepath.Join(dirPath, entry.Name()))
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no source files found in %s", dirPath)
	}
	return files, nil
}

// GetPackage returns a cached package if it exists
func (imp *Importer) GetPackage(path string) (*PackageInfo, bool) {
	pkg, ok := imp.cache[path]
	return pkg, ok
}

// CachePackage stores a compiled package
func (imp *Importer) CachePackage(path string, pkg *PackageInfo) {
	imp.cache[path] = pkg
}