// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"github.com/arc-language/core-builder/builder"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-compiler/diagnostics"
)

// Context holds the state during compilation
type Context struct {
	Builder     *builder.Builder
	Module      *ir.Module
	Diagnostics *diagnostics.DiagnosticEngine
	
	// Current compilation scope
	currentFunction *ir.Function
	currentBlock    *ir.BasicBlock
	
	// Symbol tables
	globalScope *Scope
	currentScope *Scope
	
	// Type cache
	namedTypes map[string]types.Type
	
	// Deferred statements stack (per function)
	deferredStmts [][]ir.Instruction
}

// NewContext creates a new compilation context
func NewContext(moduleName string) *Context {
	b := builder.New()
	mod := b.CreateModule(moduleName)
	
	ctx := &Context{
		Builder:     b,
		Module:      mod,
		Diagnostics: diagnostics.NewDiagnosticEngine(),
		globalScope: NewScope(nil),
		namedTypes:  make(map[string]types.Type),
		deferredStmts: make([][]ir.Instruction, 0),
	}
	
	ctx.currentScope = ctx.globalScope
	ctx.registerBuiltinTypes()
	
	return ctx
}

// registerBuiltinTypes registers primitive and builtin types
func (c *Context) registerBuiltinTypes() {
	// LLVM-style type names (for internal use)
	c.namedTypes["i1"] = types.I1
	c.namedTypes["i8"] = types.I8
	c.namedTypes["i16"] = types.I16
	c.namedTypes["i32"] = types.I32
	c.namedTypes["i64"] = types.I64
	c.namedTypes["i128"] = types.I128
	
	c.namedTypes["u8"] = types.U8
	c.namedTypes["u16"] = types.U16
	c.namedTypes["u32"] = types.U32
	c.namedTypes["u64"] = types.U64
	
	c.namedTypes["f16"] = types.F16
	c.namedTypes["f32"] = types.F32
	c.namedTypes["f64"] = types.F64
	c.namedTypes["f128"] = types.F128
	
	// Arc language type names (Go-style)
	// Signed integers
	c.namedTypes["int8"] = types.I8
	c.namedTypes["int16"] = types.I16
	c.namedTypes["int32"] = types.I32
	c.namedTypes["int64"] = types.I64
	c.namedTypes["int"] = types.I64 // Default int is 64-bit
	
	// Unsigned integers
	c.namedTypes["uint8"] = types.U8
	c.namedTypes["uint16"] = types.U16
	c.namedTypes["uint32"] = types.U32
	c.namedTypes["uint64"] = types.U64
	c.namedTypes["uint"] = types.U64 // Default uint is 64-bit
	c.namedTypes["byte"] = types.U8  // Alias for uint8
	
	// Floating point
	c.namedTypes["float32"] = types.F32
	c.namedTypes["float64"] = types.F64
	c.namedTypes["float"] = types.F64 // Default float is 64-bit
	
	// Special types
	c.namedTypes["void"] = types.Void
	c.namedTypes["bool"] = types.I1
	
	// Convenience aliases
	c.namedTypes["rune"] = types.I32 // UTF-32 character (like Go)
}

// GetType resolves a type name to a Type
func (c *Context) GetType(name string) (types.Type, bool) {
	t, ok := c.namedTypes[name]
	return t, ok
}

// RegisterType registers a named type
func (c *Context) RegisterType(name string, typ types.Type) {
	c.namedTypes[name] = typ
	
	// If it's a struct type, also register in module
	if structTy, ok := typ.(*types.StructType); ok {
		c.Module.Types[name] = structTy
	}
}

// PushScope creates a new nested scope
func (c *Context) PushScope() {
	c.currentScope = NewScope(c.currentScope)
}

// PopScope returns to the parent scope
func (c *Context) PopScope() {
	if c.currentScope.parent != nil {
		c.currentScope = c.currentScope.parent
	}
}

// EnterFunction sets up context for compiling a function
func (c *Context) EnterFunction(fn *ir.Function) {
	c.currentFunction = fn
	c.currentBlock = nil
	c.PushScope()
	
	// Add function parameters to scope
	for _, arg := range fn.Arguments {
		c.currentScope.Define(arg.Name(), arg)
	}
	
	// Initialize deferred statements for this function
	c.deferredStmts = append(c.deferredStmts, make([]ir.Instruction, 0))
}

// ExitFunction cleans up after compiling a function
func (c *Context) ExitFunction() {
	c.currentFunction = nil
	c.currentBlock = nil
	c.PopScope()
	
	// Pop deferred statements
	if len(c.deferredStmts) > 0 {
		c.deferredStmts = c.deferredStmts[:len(c.deferredStmts)-1]
	}
}

// SetInsertBlock sets the current basic block for instruction insertion
func (c *Context) SetInsertBlock(block *ir.BasicBlock) {
	c.currentBlock = block
	c.Builder.SetInsertPoint(block)
}

// AddDeferred adds a deferred statement to the current function
func (c *Context) AddDeferred(inst ir.Instruction) {
	if len(c.deferredStmts) > 0 {
		idx := len(c.deferredStmts) - 1
		c.deferredStmts[idx] = append(c.deferredStmts[idx], inst)
	}
}

// GetDeferredStmts returns deferred statements for current function
func (c *Context) GetDeferredStmts() []ir.Instruction {
	if len(c.deferredStmts) > 0 {
		return c.deferredStmts[len(c.deferredStmts)-1]
	}
	return nil
}