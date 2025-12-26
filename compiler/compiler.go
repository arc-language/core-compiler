package compiler

import (
	"fmt"
	"path/filepath"

	"github.com/antlr4-go/antlr/v4"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-parser"
)

// Compiler represents the Arc language compiler
type Compiler struct {
	context *Context
}

// NewCompiler creates a new compiler instance
func NewCompiler(moduleName string, entryFile string) *Compiler {
	return &Compiler{
		context: NewContext(entryFile, moduleName),
	}
}

// CompileFile compiles an Arc source file to IR
func (c *Compiler) CompileFile(filename string) (*ir.Module, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %v", err)
	}

	return c.compileFileInternal(absPath, true)
}

// CompilePackage compiles all files in a directory as a single package
func (c *Compiler) CompilePackage(dirPath string) (*PackageInfo, error) {
	// 1. Check Cache
	if pkg, ok := c.context.Importer.GetPackage(dirPath); ok {
		if pkg.IsProcessing {
			return nil, fmt.Errorf("circular dependency detected importing %s", dirPath)
		}
		return pkg, nil
	}

	// 2. Mark as processing
	pkgInfo := &PackageInfo{
		SourcePath:   dirPath,
		IsProcessing: true,
	}
	c.context.Importer.CachePackage(dirPath, pkgInfo)

	// 3. Find source files
	files, err := c.context.Importer.GetSourceFiles(dirPath)
	if err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: Compiling package at %s (%d files)\n", dirPath, len(files))

	// 4. Compile all files in directory
	var packageName string
	
	// Preserve current namespace to restore after compiling package
	prevNs := c.context.currentNamespace
	
	for _, file := range files {
		// Reset namespace to root before parsing a new file in a package,
		// relying on the file's `namespace` decl to set it correctly.
		// However, we must ensure consistency across the package.
		c.context.currentNamespace = c.context.rootNamespace
		
		_, err := c.compileFileInternal(file, false) 
		if err != nil {
			return nil, err
		}
		
		// Validation: Verify package consistency
		currentNsName := c.context.currentNamespace.Name
		if currentNsName == "" {
			// File didn't declare a namespace - implicit "main" or root?
			// For now, allow mixed if logic requires, but strict mode suggests matching.
		} else {
			if packageName == "" {
				packageName = currentNsName
			} else if currentNsName != packageName {
				 return nil, fmt.Errorf("file %s declares namespace '%s', expected '%s' (all files in a directory must belong to the same package)", 
					file, currentNsName, packageName)
			}
		}
	}

	// 5. Finalize
	pkgInfo.Name = packageName
	pkgInfo.Namespace = c.context.GetOrCreateNamespace(packageName)
	pkgInfo.IsProcessing = false
	
	// Restore namespace
	c.context.currentNamespace = prevNs
	
	fmt.Printf("DEBUG: Package %s compiled successfully (Namespace: %s)\n", dirPath, packageName)
	
	return pkgInfo, nil
}

// compileFileInternal handles the parsing and visiting of a single file
// It uses the shared context to append IR to the module
func (c *Compiler) compileFileInternal(filename string, isEntry bool) (*ir.Module, error) {
	// Read input file
	input, err := antlr.NewFileStream(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	
	// Lex
	lexer := parser.NewArcLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	
	// Parse
	p := parser.NewArcParser(stream)
	tree := p.CompilationUnit()
	
	// Generate IR
	// We pass 'c' (the Compiler) to the visitor so it can trigger recursive package compilation
	visitor := NewIRVisitor(c, filename)
	visitor.Visit(tree)
	
	// Check for compilation errors
	if c.context.Diagnostics.HasErrors() {
		if isEntry {
			c.context.Diagnostics.Print()
		}
		return nil, fmt.Errorf("compilation failed with %d error(s) in %s", 
			c.context.Diagnostics.ErrorCount(), filename)
	}
	
	// Print warnings
	if isEntry && c.context.Diagnostics.WarningCount() > 0 {
		c.context.Diagnostics.Print()
	}
	
	return c.context.Module, nil
}

// CompileString compiles Arc source code from a string
func (c *Compiler) CompileString(source string) (*ir.Module, error) {
	// Create input stream from string
	input := antlr.NewInputStream(source)
	
	// Lex
	lexer := parser.NewArcLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	
	// Parse
	p := parser.NewArcParser(stream)
	tree := p.CompilationUnit()
	
	// Generate IR
	// Pass dummy filename
	visitor := NewIRVisitor(c, "string_source")
	visitor.Visit(tree)
	
	// Check for compilation errors
	if c.context.Diagnostics.HasErrors() {
		c.context.Diagnostics.Print()
		return nil, fmt.Errorf("compilation failed with %d error(s)", 
			c.context.Diagnostics.ErrorCount())
	}
	
	return c.context.Module, nil
}

// GetModule returns the compiled module
func (c *Compiler) GetModule() *ir.Module {
	return c.context.Module
}

// GetContext returns the compilation context
func (c *Compiler) GetContext() *Context {
	return c.context
}