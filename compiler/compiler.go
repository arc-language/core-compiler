package compiler

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-parser"
)

// Compiler represents the Arc language compiler
type Compiler struct {
	context *Context
}

// NewCompiler creates a new compiler instance
func NewCompiler(moduleName string) *Compiler {
	return &Compiler{
		context: NewContext(moduleName),
	}
}

// CompileFile compiles an Arc source file to IR
func (c *Compiler) CompileFile(filename string) (*ir.Module, error) {
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
	visitor := NewIRVisitor(c.context)
	visitor.Visit(tree)
	
	// Check for compilation errors
	if c.context.Diagnostics.HasErrors() {
		c.context.Diagnostics.Print()
		return nil, fmt.Errorf("compilation failed with %d error(s)", 
			c.context.Diagnostics.ErrorCount())
	}
	
	// Print warnings
	if c.context.Diagnostics.WarningCount() > 0 {
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
	visitor := NewIRVisitor(c.context)
	visitor.Visit(tree)
	
	// Check for compilation errors
	if c.context.Diagnostics.HasErrors() {
		c.context.Diagnostics.Print()
		return nil, fmt.Errorf("compilation failed with %d error(s)", 
			c.context.Diagnostics.ErrorCount())
	}
	
	// Print warnings
	if c.context.Diagnostics.WarningCount() > 0 {
		c.context.Diagnostics.Print()
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