I'll create a comprehensive logging.go that replaces the diagnostic system entirely, and update all files to use it.

## logging.go (NEW)

```go
// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"fmt"
	"os"
	"sync"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// Global logging configuration
const (
	EnableDebugLogging   = true  // Set to false to disable debug logs
	EnableInfoLogging    = true  // Set to false to disable info logs
	EnableWarningLogging = true  // Set to false to disable warning logs
	EnableErrorLogging   = true  // Always keep errors enabled
)

// Logger provides centralized logging for the compiler
type Logger struct {
	mu          sync.Mutex
	prefix      string
	errorCount  int
	warnCount   int
	infoCount   int
	debugCount  int
}

var (
	globalLogger = &Logger{prefix: "[Arc]"}
)

// NewLogger creates a new logger with a custom prefix
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if EnableDebugLogging {
		l.log(LogLevelDebug, format, args...)
		l.mu.Lock()
		l.debugCount++
		l.mu.Unlock()
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	if EnableInfoLogging {
		l.log(LogLevelInfo, format, args...)
		l.mu.Lock()
		l.infoCount++
		l.mu.Unlock()
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	if EnableWarningLogging {
		l.log(LogLevelWarning, format, args...)
		l.mu.Lock()
		l.warnCount++
		l.mu.Unlock()
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if EnableErrorLogging {
		l.log(LogLevelError, format, args...)
		l.mu.Lock()
		l.errorCount++
		l.mu.Unlock()
	}
}

// ErrorAt logs an error at a specific source location
func (l *Logger) ErrorAt(file string, line, column int, format string, args ...interface{}) {
	if EnableErrorLogging {
		message := fmt.Sprintf(format, args...)
		l.log(LogLevelError, "%s:%d:%d: %s", file, line, column, message)
		l.mu.Lock()
		l.errorCount++
		l.mu.Unlock()
	}
}

// WarningAt logs a warning at a specific source location
func (l *Logger) WarningAt(file string, line, column int, format string, args ...interface{}) {
	if EnableWarningLogging {
		message := fmt.Sprintf(format, args...)
		l.log(LogLevelWarning, "%s:%d:%d: %s", file, line, column, message)
		l.mu.Lock()
		l.warnCount++
		l.mu.Unlock()
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var levelStr string
	var output *os.File
	
	switch level {
	case LogLevelDebug:
		levelStr = "DEBUG"
		output = os.Stdout
	case LogLevelInfo:
		levelStr = "INFO"
		output = os.Stdout
	case LogLevelWarning:
		levelStr = "WARN"
		output = os.Stderr
	case LogLevelError:
		levelStr = "ERROR"
		output = os.Stderr
	}

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(output, "%s [%s] %s\n", l.prefix, levelStr, message)
}

// HasErrors returns true if any errors were logged
func (l *Logger) HasErrors() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errorCount > 0
}

// ErrorCount returns the number of errors logged
func (l *Logger) ErrorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errorCount
}

// WarningCount returns the number of warnings logged
func (l *Logger) WarningCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.warnCount
}

// Reset resets all counters
func (l *Logger) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorCount = 0
	l.warnCount = 0
	l.infoCount = 0
	l.debugCount = 0
}

// PrintSummary prints a summary of logged messages
func (l *Logger) PrintSummary() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.errorCount > 0 || l.warnCount > 0 {
		fmt.Fprintf(os.Stderr, "\n%s Compilation Summary:\n", l.prefix)
		if l.errorCount > 0 {
			fmt.Fprintf(os.Stderr, "  Errors: %d\n", l.errorCount)
		}
		if l.warnCount > 0 {
			fmt.Fprintf(os.Stderr, "  Warnings: %d\n", l.warnCount)
		}
	}
}

// Global logging functions for convenience
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Warning(format string, args ...interface{}) {
	globalLogger.Warning(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

func ErrorAt(file string, line, column int, format string, args ...interface{}) {
	globalLogger.ErrorAt(file, line, column, format, args...)
}

func WarningAt(file string, line, column int, format string, args ...interface{}) {
	globalLogger.WarningAt(file, line, column, format, args...)
}

func HasErrors() bool {
	return globalLogger.HasErrors()
}

func ErrorCount() int {
	return globalLogger.ErrorCount()
}

func WarningCount() int {
	return globalLogger.WarningCount()
}

func PrintSummary() {
	globalLogger.PrintSummary()
}
```

## context.go (UPDATED)

```go
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
```

## compiler.go (UPDATED)

```go
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
```

## codegen.go (UPDATED)

```go
// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"fmt"
	"os"

	"github.com/arc-language/core-codegen/codegen"
)

// CompileToIR generates textual IR from the module
func (c *Compiler) CompileToIR(outputPath string) error {
	c.logger.Info("Generating textual IR to: %s", outputPath)
	
	if c.context.Module == nil {
		c.logger.Error("No module to compile")
		return fmt.Errorf("no module to compile")
	}

	// Generate IR text
	irText := c.context.Module.String()
	c.logger.Debug("Generated %d bytes of IR text", len(irText))

	// Write to file
	if err := os.WriteFile(outputPath, []byte(irText), 0644); err != nil {
		c.logger.Error("Failed to write IR file '%s': %v", outputPath, err)
		return fmt.Errorf("failed to write IR file: %v", err)
	}

	c.logger.Info("Successfully wrote IR to: %s", outputPath)
	return nil
}

// CompileToObject generates an object file from the module
func (c *Compiler) CompileToObject(outputPath string) error {
	c.logger.Info("Generating object file to: %s", outputPath)
	
	if c.context.Module == nil {
		c.logger.Error("No module to compile")
		return fmt.Errorf("no module to compile")
	}

	// Generate object code
	c.logger.Debug("Calling code generator for module '%s'", c.context.Module.Name)
	objData, err := codegen.GenerateObject(c.context.Module)
	if err != nil {
		c.logger.Error("Code generation failed: %v", err)
		return fmt.Errorf("code generation failed: %v", err)
	}

	c.logger.Debug("Generated %d bytes of object code", len(objData))

	// Write object file
	if err := os.WriteFile(outputPath, objData, 0644); err != nil {
		c.logger.Error("Failed to write object file '%s': %v", outputPath, err)
		return fmt.Errorf("failed to write object file: %v", err)
	}

	c.logger.Info("Successfully wrote object file to: %s", outputPath)
	return nil
}
```

## importer.go (UPDATED)

```go
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