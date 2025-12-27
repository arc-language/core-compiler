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
	logger  *Logger
}

// NewCompiler creates a new compiler instance
func NewCompiler(moduleName string, entryFile string) *Compiler {
	logger := NewLogger(fmt.Sprintf("[Compiler:%s]", moduleName))
	logger.Info("Creating compiler for module '%s' with entry file '%s'", moduleName, entryFile)
	
	return &Compiler{
		context: NewContext(entryFile, moduleName),
		logger:  logger,
	}
}

// CompileFile compiles an Arc source file to IR
func (c *Compiler) CompileFile(filename string) (*ir.Module, error) {
	c.logger.Info("Compiling file: %s", filename)
	
	absPath, err := filepath.Abs(filename)
	if err != nil {
		c.logger.Error("Failed to resolve path '%s': %v", filename, err)
		return nil, fmt.Errorf("failed to resolve path: %v", err)
	}

	return c.compileFileInternal(absPath, true)
}

// CompilePackage compiles all files in a directory as a single package
func (c *Compiler) CompilePackage(dirPath string) (*PackageInfo, error) {
	c.logger.Debug("Starting package compilation for directory: %s", dirPath)
	
	// 1. Check Cache
	if pkg, ok := c.context.Importer.GetPackage(dirPath); ok {
		if pkg.IsProcessing {
			c.logger.Error("Circular dependency detected importing '%s'", dirPath)
			return nil, fmt.Errorf("circular dependency detected importing %s", dirPath)
		}
		c.logger.Debug("Package '%s' found in cache", dirPath)
		return pkg, nil
	}

	// 2. Mark as processing
	pkgInfo := &PackageInfo{
		SourcePath:   dirPath,
		IsProcessing: true,
	}
	c.context.Importer.CachePackage(dirPath, pkgInfo)
	c.logger.Debug("Marked package '%s' as processing", dirPath)

	// 3. Find source files
	files, err := c.context.Importer.GetSourceFiles(dirPath)
	if err != nil {
		c.logger.Error("Failed to find source files in '%s': %v", dirPath, err)
		return nil, err
	}

	c.logger.Info("Compiling package at '%s' with %d file(s)", dirPath, len(files))

	// 4. Compile all files in directory
	var packageName string
	
	// Preserve current namespace to restore after compiling package
	prevNs := c.context.currentNamespace
	
	for i, file := range files {
		c.logger.Debug("Compiling file %d/%d: %s", i+1, len(files), file)
		
		// Reset namespace to root before parsing a new file in a package
		c.context.currentNamespace = c.context.rootNamespace
		
		_, err := c.compileFileInternal(file, false) 
		if err != nil {
			c.logger.Error("Compilation failed for file '%s': %v", file, err)
			return nil, err
		}
		
		// Validation: Verify package consistency
		currentNsName := c.context.currentNamespace.Name
		if currentNsName == "" {
			// File didn't declare a namespace
			c.logger.Debug("File '%s' has no namespace declaration", file)
		} else {
			if packageName == "" {
				packageName = currentNsName
				c.logger.Debug("Package namespace set to '%s'", packageName)
			} else if currentNsName != packageName {
				c.logger.Error("File '%s' declares namespace '%s', expected '%s'", file, currentNsName, packageName)
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
	
	c.logger.Info("Package '%s' compiled successfully (Namespace: %s)", dirPath, packageName)
	
	return pkgInfo, nil
}

// compileFileInternal handles the parsing and visiting of a single file
func (c *Compiler) compileFileInternal(filename string, isEntry bool) (*ir.Module, error) {
	c.logger.Debug("Internal compilation of file: %s (isEntry=%v)", filename, isEntry)
	
	// Read input file
	input, err := antlr.NewFileStream(filename)
	if err != nil {
		c.logger.Error("Failed to open file '%s': %v", filename, err)
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	
	// Lex
	c.logger.Debug("Lexing file: %s", filename)
	lexer := parser.NewArcLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	
	// Parse
	c.logger.Debug("Parsing file: %s", filename)
	p := parser.NewArcParser(stream)
	tree := p.CompilationUnit()
	
	// Generate IR
	c.logger.Debug("Generating IR for file: %s", filename)
	visitor := NewIRVisitor(c, filename)
	visitor.Visit(tree)
	
	// Check for compilation errors
	if c.context.Logger.HasErrors() {
		if isEntry {
			c.context.Logger.PrintSummary()
		}
		return nil, fmt.Errorf("compilation failed with %d error(s) in %s", 
			c.context.Logger.ErrorCount(), filename)
	}
	
	// Print warnings summary
	if isEntry && c.context.Logger.WarningCount() > 0 {
		c.context.Logger.PrintSummary()
	}
	
	c.logger.Info("Successfully compiled file: %s", filename)
	
	return c.context.Module, nil
}

// CompileString compiles Arc source code from a string
func (c *Compiler) CompileString(source string) (*ir.Module, error) {
	c.logger.Info("Compiling source string (%d bytes)", len(source))
	
	// Create input stream from string
	input := antlr.NewInputStream(source)
	
	// Lex
	lexer := parser.NewArcLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	
	// Parse
	p := parser.NewArcParser(stream)
	tree := p.CompilationUnit()
	
	// Generate IR
	visitor := NewIRVisitor(c, "<string>")
	visitor.Visit(tree)
	
	// Check for compilation errors
	if c.context.Logger.HasErrors() {
		c.context.Logger.PrintSummary()
		return nil, fmt.Errorf("compilation failed with %d error(s)", 
			c.context.Logger.ErrorCount())
	}
	
	c.logger.Info("Successfully compiled source string")
	
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