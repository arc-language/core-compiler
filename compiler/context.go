// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"github.com/arc-language/core-builder/builder"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
)

// LoopInfo holds the target blocks for control flow within a loop
type LoopInfo struct {
	ContinueBlock *ir.BasicBlock // Where 'continue' jumps to
	BreakBlock    *ir.BasicBlock // Where 'break' jumps to
}

// Namespace represents a named collection of declarations
type Namespace struct {
	Name      string
	Functions map[string]*ir.Function
	Types     map[string]types.Type
	Parent    *Namespace
}

// NewNamespace creates a new namespace
func NewNamespace(name string, parent *Namespace) *Namespace {
	return &Namespace{
		Name:      name,
		Functions: make(map[string]*ir.Function),
		Types:     make(map[string]types.Type),
		Parent:    parent,
	}
}

// LookupFunction searches for a function in this namespace and parents
func (ns *Namespace) LookupFunction(name string) (*ir.Function, bool) {
	if fn, ok := ns.Functions[name]; ok {
		return fn, true
	}
	if ns.Parent != nil {
		return ns.Parent.LookupFunction(name)
	}
	return nil, false
}

// Context holds the state during compilation
type Context struct {
	Builder  *builder.Builder
	Module   *ir.Module
	Importer *Importer
	Logger   *Logger
	
	// Current compilation scope
	currentFunction *ir.Function
	currentBlock    *ir.BasicBlock
	
	// Symbol tables
	globalScope  *Scope
	currentScope *Scope
	
	// Namespace management
	rootNamespace    *Namespace
	currentNamespace *Namespace
	
	// Registry for all loaded namespaces (Key: Namespace Name)
	NamespaceRegistry map[string]*Namespace
	
	// Type cache
	namedTypes map[string]types.Type
	
	// Struct Field Mapping: StructName -> FieldName -> Index
	StructFieldIndices map[string]map[string]int
	
	// Class Field Mapping: ClassName -> FieldName -> Index
	ClassFieldIndices map[string]map[string]int
	
	// Track which types are classes (for reference semantics)
	classTypes map[string]bool
	
	// Deferred statements stack (per function)
	deferredStmts [][]ir.Instruction

	// Loop stack for break/continue
	loopStack []LoopInfo
}

// NewContext creates a new compilation context
func NewContext(entryFile string, moduleName string) *Context {
	b := builder.New()
	mod := b.CreateModule(moduleName)
	
	rootNs := NewNamespace("", nil)
	logger := NewLogger("[Context]")
	
	ctx := &Context{
		Builder:            b,
		Module:             mod,
		Logger:             logger,
		Importer:           NewImporter(entryFile),
		globalScope:        NewScope(nil),
		namedTypes:         make(map[string]types.Type),
		StructFieldIndices: make(map[string]map[string]int),
		ClassFieldIndices:  make(map[string]map[string]int),
		classTypes:         make(map[string]bool),
		deferredStmts:      make([][]ir.Instruction, 0),
		loopStack:          make([]LoopInfo, 0),
		rootNamespace:      rootNs,
		currentNamespace:   rootNs,
		NamespaceRegistry:  make(map[string]*Namespace),
	}
	
	ctx.currentScope = ctx.globalScope
	ctx.registerBuiltinTypes()
	
	logger.Debug("Context initialized for module '%s'", moduleName)
	
	return ctx
}

// SetNamespace sets the current namespace
func (c *Context) SetNamespace(name string) *Namespace {
	// If the namespace name is empty, we are in the root
	if name == "" {
		c.currentNamespace = c.rootNamespace
		c.Logger.Debug("Set namespace to root")
		return c.rootNamespace
	}

	// Check registry first (cross-file persistence)
	if ns, ok := c.NamespaceRegistry[name]; ok {
		c.currentNamespace = ns
		c.Logger.Debug("Switched to existing namespace '%s'", name)
		return ns
	}

	// Create new namespace attached to root (flat namespace hierarchy for now)
	ns := NewNamespace(name, c.rootNamespace)
	c.NamespaceRegistry[name] = ns
	c.currentNamespace = ns
	c.Logger.Debug("Created new namespace '%s'", name)
	return ns
}

// GetOrCreateNamespace gets or creates a namespace by name
func (c *Context) GetOrCreateNamespace(name string) *Namespace {
	if name == "" {
		return c.rootNamespace
	}
	if ns, ok := c.NamespaceRegistry[name]; ok {
		return ns
	}
	ns := NewNamespace(name, c.rootNamespace)
	c.NamespaceRegistry[name] = ns
	c.Logger.Debug("Created namespace '%s' via GetOrCreate", name)
	return ns
}

// LookupInNamespace looks up a function in a specific namespace
func (c *Context) LookupInNamespace(namespaceName, functionName string) (*ir.Function, bool) {
	ns := c.GetOrCreateNamespace(namespaceName)
	return ns.LookupFunction(functionName)
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
	
	// Arc language type names
	// Signed integers
	c.namedTypes["int8"] = types.I8
	c.namedTypes["int16"] = types.I16
	c.namedTypes["int32"] = types.I32
	c.namedTypes["int64"] = types.I64
	c.namedTypes["int"] = types.I64 // Default int is 64-bit
	c.namedTypes["isize"] = types.I64 

	// Unsigned integers
	c.namedTypes["uint8"] = types.U8
	c.namedTypes["uint16"] = types.U16
	c.namedTypes["uint32"] = types.U32
	c.namedTypes["uint64"] = types.U64
	c.namedTypes["uint"] = types.U64 // Default uint is 64-bit
	c.namedTypes["byte"] = types.U8  // Alias for uint8
	c.namedTypes["usize"] = types.U64 
	
	// Floating point
	c.namedTypes["float32"] = types.F32
	c.namedTypes["float64"] = types.F64
	c.namedTypes["float"] = types.F64 // Default float is 64-bit
	
	// Special types
	c.namedTypes["void"] = types.Void
	c.namedTypes["bool"] = types.I1
	c.namedTypes["char"] = types.U32 // Unicode code point (uint32)
	c.namedTypes["string"] = types.NewPointer(types.I8) // For now, *i8
	
	c.Logger.Debug("Registered %d builtin types", len(c.namedTypes))
}

// GetType resolves a type name to a Type
func (c *Context) GetType(name string) (types.Type, bool) {
	t, ok := c.namedTypes[name]
	return t, ok
}

// RegisterType registers a named type
func (c *Context) RegisterType(name string, typ types.Type) {
	c.namedTypes[name] = typ
	// Also register in current namespace
	c.currentNamespace.Types[name] = typ
	
	// If it's a struct type, also register in module
	if structTy, ok := typ.(*types.StructType); ok {
		c.Module.Types[name] = structTy
	}
	
	c.Logger.Debug("Registered type '%s'", name)
}

// RegisterClass registers a class type (reference type)
func (c *Context) RegisterClass(name string, typ types.Type) {
	c.namedTypes[name] = typ
	c.classTypes[name] = true
	// Also register in current namespace
	c.currentNamespace.Types[name] = typ
	
	if structTy, ok := typ.(*types.StructType); ok {
		c.Module.Types[name] = structTy
	}
	
	c.Logger.Debug("Registered class type '%s'", name)
}

// IsClassType checks if a type name refers to a class
func (c *Context) IsClassType(name string) bool {
	return c.classTypes[name]
}

// PushScope creates a new nested scope
func (c *Context) PushScope() {
	c.currentScope = NewScope(c.currentScope)
	c.Logger.Debug("Pushed new scope")
}

// PopScope returns to the parent scope
func (c *Context) PopScope() {
	if c.currentScope.parent != nil {
		c.currentScope = c.currentScope.parent
		c.Logger.Debug("Popped scope")
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
	
	c.Logger.Debug("Entered function '%s'", fn.Name())
}

// ExitFunction cleans up after compiling a function
func (c *Context) ExitFunction() {
	if c.currentFunction != nil {
		c.Logger.Debug("Exited function '%s'", c.currentFunction.Name())
	}
	
	c.currentFunction = nil
	c.currentBlock = nil
	c.PopScope()
	
	// Pop deferred statements
	if len(c.deferredStmts) > 0 {
		c.deferredStmts = c.deferredStmts[:len(c.deferredStmts)-1]
	}
	// Reset loop stack just in case
	c.loopStack = c.loopStack[:0]
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

// --- Loop Management ---

func (c *Context) PushLoop(cont, brk *ir.BasicBlock) {
	c.loopStack = append(c.loopStack, LoopInfo{
		ContinueBlock: cont,
		BreakBlock:    brk,
	})
	c.Logger.Debug("Pushed loop context")
}

func (c *Context) PopLoop() {
	if len(c.loopStack) > 0 {
		c.loopStack = c.loopStack[:len(c.loopStack)-1]
		c.Logger.Debug("Popped loop context")
	}
}

func (c *Context) CurrentLoop() *LoopInfo {
	if len(c.loopStack) == 0 {
		return nil
	}
	return &c.loopStack[len(c.loopStack)-1]
}