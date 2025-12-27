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
	logger   *Logger
}

// NewImporter creates a new importer based on the entry file location
func NewImporter(entryFile string) *Importer {
	absPath, _ := filepath.Abs(entryFile)
	logger := NewLogger("[Importer]")
	logger.Debug("Created importer with entry directory: %s", filepath.Dir(absPath))
	
	return &Importer{
		entry: filepath.Dir(absPath),
		cache:    make(map[string]*PackageInfo),
		logger:   logger,
	}
}

// ResolvePath converts an import string to an absolute directory path
func (imp *Importer) ResolvePath(currentFileDir, importPath string) (string, error) {
	imp.logger.Debug("Resolving import path '%s' from directory '%s'", importPath, currentFileDir)
	
	// Handle local relative imports (starting with ./ or ../)
	if strings.HasPrefix(importPath, ".") {
		if currentFileDir == "" {
			currentFileDir = imp.entryDir
		}
		absPath, err := filepath.Abs(filepath.Join(currentFileDir, importPath))
		if err != nil {
			imp.logger.Error("Failed to resolve relative path '%s': %v", importPath, err)
			return "", err
		}
		imp.logger.Debug("Resolved relative import to: %s", absPath)
		return absPath, nil
	}

	// TODO: Handle standard library and module imports
	// For now, treat non-relative imports as relative to entry directory or vendor
	absPath, err := filepath.Abs(filepath.Join(imp.entryDir, importPath))
	if err != nil {
		imp.logger.Error("Failed to resolve import path '%s': %v", importPath, err)
		return "", err
	}
	
	imp.logger.Debug("Resolved import to: %s", absPath)
	return absPath, nil
}

// GetSourceFiles returns all .arc files in a directory
func (imp *Importer) GetSourceFiles(dirPath string) ([]string, error) {
	imp.logger.Debug("Scanning directory for source files: %s", dirPath)
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		imp.logger.Error("Failed to read directory '%s': %v", dirPath, err)
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".arc") || strings.HasSuffix(entry.Name(), ".lang")) {
			files = append(files, filepath.Join(dirPath, entry.Name()))
		}
	}

	if len(files) == 0 {
		imp.logger.Warning("No source files found in directory: %s", dirPath)
		return nil, fmt.Errorf("no source files found in %s", dirPath)
	}
	
	imp.logger.Debug("Found %d source file(s) in '%s'", len(files), dirPath)
	return files, nil
}

// GetPackage returns a cached package if it exists
func (imp *Importer) GetPackage(path string) (*PackageInfo, bool) {
	pkg, ok := imp.cache[path]
	if ok {
		imp.logger.Debug("Package cache hit for: %s", path)
	}
	return pkg, ok
}

// CachePackage stores a compiled package
func (imp *Importer) CachePackage(path string, pkg *PackageInfo) {
	imp.cache[path] = pkg
	imp.logger.Debug("Cached package at path: %s", path)
}