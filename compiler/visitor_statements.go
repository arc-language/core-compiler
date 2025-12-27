package compiler

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

func (v *IRVisitor) VisitStatement(ctx *parser.StatementContext) interface{} {
	if ctx.VariableDecl() != nil {
		return v.Visit(ctx.VariableDecl())
	}
	if ctx.ConstDecl() != nil {
		return v.Visit(ctx.ConstDecl())
	}
	if ctx.AssignmentStmt() != nil {
		return v.Visit(ctx.AssignmentStmt())
	}
	if ctx.ReturnStmt() != nil {
		return v.Visit(ctx.ReturnStmt())
	}
	if ctx.IfStmt() != nil {
		return v.Visit(ctx.IfStmt())
	}
	if ctx.ForStmt() != nil {
		return v.Visit(ctx.ForStmt())
	}
	if ctx.BreakStmt() != nil {
		return v.Visit(ctx.BreakStmt())
	}
	if ctx.ContinueStmt() != nil {
		return v.Visit(ctx.ContinueStmt())
	}
	if ctx.DeferStmt() != nil {
		return v.Visit(ctx.DeferStmt())
	}
	if ctx.ExpressionStmt() != nil {
		return v.Visit(ctx.ExpressionStmt())
	}
	if ctx.Block() != nil {
		return v.Visit(ctx.Block())
	}
	return nil
}

func (v *IRVisitor) VisitBlock(ctx *parser.BlockContext) interface{} {
	stmts := ctx.AllStatement()
	v.ctx.PushScope()
	
	for i, stmt := range stmts {
		v.Visit(stmt)
		
		// Stop if we hit a terminator
		if v.ctx.currentBlock != nil && v.ctx.currentBlock.Terminator() != nil {
			v.logger.Debug("Hit terminator at statement %d in block, stopping", i)
			break
		}
	}
	
	v.ctx.PopScope()
	return nil
}

func (v *IRVisitor) VisitAssignmentStmt(ctx *parser.AssignmentStmtContext) interface{} {
	lhsCtx := ctx.LeftHandSide()
	
	// Simple Variable Assignment: IDENTIFIER = value
	if lhsCtx.IDENTIFIER() != nil && lhsCtx.DOT() == nil && lhsCtx.STAR() == nil {
		name := lhsCtx.IDENTIFIER().GetText()
		rhs := v.Visit(ctx.Expression()).(ir.Value)
		
		v.logger.Debug("Assigning to variable: %s", name)
		
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			v.ctx.Logger.Error("Undefined: %s", name)
			return nil
		}
		
		if sym.IsConst {
			v.ctx.Logger.Error("Cannot assign to constant '%s'", name)
			return nil
		}
		
		if ptr, isAlloca := sym.Value.(*ir.AllocaInst); isAlloca {
			v.ctx.Builder.CreateStore(rhs, ptr)
			return nil
		}
		
		v.ctx.currentScope.Define(name, rhs)
		return nil
	}
	
	// Pointer Assignment: *ptr = value
	if lhsCtx.STAR() != nil {
		v.logger.Debug("Assigning through pointer dereference")
		ptr := v.Visit(lhsCtx.PostfixExpression()).(ir.Value)
		rhs := v.Visit(ctx.Expression()).(ir.Value)
		v.ctx.Builder.CreateStore(rhs, ptr)
		return nil
	}
	
	// Field Assignment: obj.field = value
	if lhsCtx.DOT() != nil && lhsCtx.PostfixExpression() != nil {
		postfixCtx := lhsCtx.PostfixExpression()
		var basePtr ir.Value
		
		// Check if the postfix expression is just a simple identifier
		if postfixCtx.PrimaryExpression() != nil {
			primaryCtx := postfixCtx.PrimaryExpression()
			if primaryCtx.IDENTIFIER() != nil {
				varName := primaryCtx.IDENTIFIER().GetText()
				
				if sym, ok := v.ctx.currentScope.Lookup(varName); ok {
					if alloca, isAlloca := sym.Value.(*ir.AllocaInst); isAlloca {
						// Check what the alloca contains
						if _, isPtr := alloca.AllocatedType.(*types.PointerType); isPtr {
							// It's a pointer - load it
							basePtr = v.ctx.Builder.CreateLoad(alloca.AllocatedType, alloca, "")
						} else if _, isStruct := alloca.AllocatedType.(*types.StructType); isStruct {
							// Direct struct allocation
							basePtr = alloca
						}
					}
				}
			}
		}
		
		if basePtr == nil {
			basePtr = v.Visit(postfixCtx).(ir.Value)
		}
		
		fieldName := lhsCtx.IDENTIFIER().GetText()
		
		v.logger.Debug("Assigning to field: %s", fieldName)
		
		// Now basePtr should be a pointer to a struct
		if basePtr != nil {
			if ptrType, ok := basePtr.Type().(*types.PointerType); ok {
				if structType, ok := ptrType.ElementType.(*types.StructType); ok {
					
					isClass := v.ctx.IsClassType(structType.Name)
					var fieldIdx int = -1
					
					if isClass {
						if fieldIndices, ok := v.ctx.ClassFieldIndices[structType.Name]; ok {
							if idx, ok := fieldIndices[fieldName]; ok {
								fieldIdx = idx
							}
						}
					} else {
						fieldIdx = v.findFieldIndex(structType, fieldName)
					}
					
					if fieldIdx >= 0 {
						gep := v.ctx.Builder.CreateStructGEP(structType, basePtr, fieldIdx, "")
						rhs := v.Visit(ctx.Expression()).(ir.Value)
						v.ctx.Builder.CreateStore(rhs, gep)
						return nil
					} else {
						v.ctx.Logger.Error("Struct/class '%s' has no field '%s'", structType.Name, fieldName)
						return nil
					}
				}
			}
		}
		
		v.ctx.Logger.Error("Cannot assign to field (expected pointer to struct, got %v)", basePtr.Type())
		return nil
	}

	v.ctx.Logger.Error("Complex assignment not yet supported")
	return nil
}

func (v *IRVisitor) VisitReturnStmt(ctx *parser.ReturnStmtContext) interface{} {
	v.logger.Debug("Compiling return statement")
	
	// Execute deferred statements
	deferred := v.ctx.GetDeferredStmts()
	for i := len(deferred) - 1; i >= 0; i-- {
		_ = deferred[i]
	}
	
	if ctx.Expression() != nil {
		retVal := v.Visit(ctx.Expression()).(ir.Value)
		
		// Cast to expected return type if needed
		if v.ctx.currentFunction != nil {
			expectedType := v.ctx.currentFunction.FuncType.ReturnType
			if !retVal.Type().Equal(expectedType) {
				retVal = v.castValue(retVal, expectedType)
			}
		}
		
		v.ctx.Builder.CreateRet(retVal)
	} else {
		v.ctx.Builder.CreateRetVoid()
	}
	
	return nil
}

func (v *IRVisitor) VisitExpressionStmt(ctx *parser.ExpressionStmtContext) interface{} {
	// Check if this looks like an assignment that wasn't parsed as such
	exprText := ctx.Expression().GetText()
	if strings.Contains(exprText, "=") && !strings.Contains(exprText, "==") && !strings.Contains(exprText, "!=") {
		v.logger.Warning("Expression contains '=' - might be a failed assignment parse: %s", exprText)
	}
	
	v.Visit(ctx.Expression())
	return nil
}

func (v *IRVisitor) VisitDeferStmt(ctx *parser.DeferStmtContext) interface{} {
	if ctx.Expression() != nil {
		_ = v.Visit(ctx.Expression())
	}
	v.ctx.Logger.Warning("defer statement is not fully implemented yet")
	return nil
}

// Helpers for token ordering
func (v *IRVisitor) isBefore(ctx antlr.ParserRuleContext, token antlr.TerminalNode) bool {
	if ctx == nil || token == nil {
		return false
	}
	return ctx.GetStop().GetTokenIndex() < token.GetSymbol().GetTokenIndex()
}

func (v *IRVisitor) isAfter(ctx antlr.ParserRuleContext, token antlr.TerminalNode) bool {
	if ctx == nil || token == nil {
		return false
	}
	return ctx.GetStart().GetTokenIndex() > token.GetSymbol().GetTokenIndex()
}
```

## visitor_types.go (FULL)

```go
package compiler

import (
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// registerStructType registers a struct type in pass 1
func (v *IRVisitor) registerStructType(ctx *parser.StructDeclContext) {
	name := ctx.IDENTIFIER().GetText()
	
	// Check if already registered
	if _, ok := v.ctx.GetType(name); ok {
		v.logger.Debug("Struct type '%s' already registered", name)
		return
	}
	
	v.logger.Debug("Registering struct type: %s", name)
	
	// Create field map
	fieldMap := make(map[string]int)
	fieldTypes := make([]types.Type, 0)
	
	fieldIndex := 0
	for _, member := range ctx.AllStructMember() {
		if member.StructField() != nil {
			field := member.StructField()
			fieldName := field.IDENTIFIER().GetText()
			fieldType := v.resolveType(field.Type_())
			
			fieldTypes = append(fieldTypes, fieldType)
			fieldMap[fieldName] = fieldIndex
			v.logger.Debug("  Field '%s' at index %d, type: %v", fieldName, fieldIndex, fieldType)
			fieldIndex++
		}
	}
	
	// Register mapping in context
	v.ctx.StructFieldIndices[name] = fieldMap

	structType := types.NewStruct(name, fieldTypes, false)
	v.ctx.RegisterType(name, structType)
}

// registerClassType registers a class type in pass 1
func (v *IRVisitor) registerClassType(ctx *parser.ClassDeclContext) {
	name := ctx.IDENTIFIER().GetText()
	
	v.logger.Info("Registering class type: %s", name)
	
	// Check if already registered
	if _, ok := v.ctx.GetType(name); ok {
		v.logger.Debug("Class '%s' already registered", name)
		return
	}
	
	// Create field map
	fieldMap := make(map[string]int)
	fieldTypes := make([]types.Type, 0)
	
	fieldIndex := 0
	for _, member := range ctx.AllClassMember() {
		if member.ClassField() != nil {
			field := member.ClassField()
			fieldName := field.IDENTIFIER().GetText()
			fieldType := v.resolveType(field.Type_())
			
			v.logger.Debug("  Field '%s' at index %d, type: %v", fieldName, fieldIndex, fieldType)
			
			fieldTypes = append(fieldTypes, fieldType)
			fieldMap[fieldName] = fieldIndex
			fieldIndex++
		}
	}
	
	// Register mapping in context
	v.ctx.ClassFieldIndices[name] = fieldMap

	// Create struct type for the class
	structType := types.NewStruct(name, fieldTypes, false)
	
	v.ctx.RegisterClass(name, structType)
	v.logger.Debug("Registered class '%s' with %d fields", name, len(fieldTypes))
}

func (v *IRVisitor) VisitStructDecl(ctx *parser.StructDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	v.logger.Debug("Processing struct declaration: %s", name)
	
	// Type already registered in pass 1
	// Now compile methods
	for _, member := range ctx.AllStructMember() {
		if member.FunctionDecl() != nil {
			v.Visit(member.FunctionDecl())
		}
	}
	
	return nil
}

func (v *IRVisitor) VisitClassDecl(ctx *parser.ClassDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	v.logger.Info("Processing class declaration: %s", name)
	
	// Type already registered in pass 1
	// Now compile methods
	for i, member := range ctx.AllClassMember() {
		v.logger.Debug("Processing class member %d/%d", i+1, len(ctx.AllClassMember()))
		if member.FunctionDecl() != nil {
			v.Visit(member.FunctionDecl())
		} else if member.DeinitDecl() != nil {
			v.Visit(member.DeinitDecl())
		} else if member.ClassField() != nil {
			// Fields are handled in registerClassType
			v.logger.Debug("Skipping field (already registered)")
		}
	}
	
	v.logger.Info("Completed class declaration: %s", name)
	return nil
}

func (v *IRVisitor) VisitClassField(ctx *parser.ClassFieldContext) interface{} {
	// Field definitions are handled in registerClassType
	v.logger.Debug("VisitClassField called for: %s (should not process here)", ctx.IDENTIFIER().GetText())
	return nil
}

func (v *IRVisitor) VisitClassMember(ctx *parser.ClassMemberContext) interface{} {
	if ctx.ClassField() != nil {
		return v.Visit(ctx.ClassField())
	}
	if ctx.FunctionDecl() != nil {
		return v.Visit(ctx.FunctionDecl())
	}
	if ctx.DeinitDecl() != nil {
		return v.Visit(ctx.DeinitDecl())
	}
	return nil
}

func (v *IRVisitor) VisitDeinitDecl(ctx *parser.DeinitDeclContext) interface{} {
	v.ctx.Logger.Warning("deinit is not yet implemented")
	return nil
}
```

## scope.go (No changes needed - no logging)

```go
package compiler

import (
	"github.com/arc-language/core-builder/ir"
)

// Symbol represents a named value in the symbol table
type Symbol struct {
	Name      string
	Value     ir.Value
	IsConst   bool
	Namespace string // Which namespace this symbol belongs to
}

// Scope represents a lexical scope with symbol table
type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

// NewScope creates a new scope
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

// Define adds a symbol to the current scope
func (s *Scope) Define(name string, value ir.Value) {
	s.symbols[name] = &Symbol{
		Name:    name,
		Value:   value,
		IsConst: false,
	}
}

// DefineConst adds a constant symbol to the current scope
func (s *Scope) DefineConst(name string, value ir.Value) {
	s.symbols[name] = &Symbol{
		Name:    name,
		Value:   value,
		IsConst: true,
	}
}

// DefineInNamespace adds a symbol with namespace information
func (s *Scope) DefineInNamespace(name string, value ir.Value, namespace string) {
	s.symbols[name] = &Symbol{
		Name:      name,
		Value:     value,
		IsConst:   false,
		Namespace: namespace,
	}
}

// Lookup searches for a symbol in the current scope and parent scopes
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	// Check current scope
	if sym, ok := s.symbols[name]; ok {
		return sym, true
	}
	
	// Check parent scopes
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	
	return nil, false
}

// LookupLocal searches only the current scope (not parents)
func (s *Scope) LookupLocal(name string) (*Symbol, bool) {
	sym, ok := s.symbols[name]
	return sym, ok
}

// IsDefined checks if a symbol exists in current scope
func (s *Scope) IsDefined(name string) bool {
	_, ok := s.symbols[name]
	return ok
}