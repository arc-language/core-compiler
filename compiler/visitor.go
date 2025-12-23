package compiler

import (
	"fmt"
	"strconv"

	"github.com/antlr4-go/antlr/v4"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// IRVisitor implements the ANTLR visitor pattern to generate IR
type IRVisitor struct {
	*parser.BaseArcParserVisitor
	ctx *Context
}

// NewIRVisitor creates a new IR visitor
func NewIRVisitor(ctx *Context) *IRVisitor {
	return &IRVisitor{
		BaseArcParserVisitor: &parser.BaseArcParserVisitor{},
		ctx:                  ctx,
	}
}

// Visit overrides the base Visit to add explicit dispatching
func (v *IRVisitor) Visit(tree antlr.ParseTree) interface{} {
	if tree == nil {
		return nil
	}

	// Explicitly dispatch to the correct visitor method based on context type
	switch ctx := tree.(type) {
	case *parser.CompilationUnitContext:
		return v.VisitCompilationUnit(ctx)
	case *parser.TopLevelDeclContext:
		return v.VisitTopLevelDecl(ctx)
	case *parser.NamespaceDeclContext:
		return v.VisitNamespaceDecl(ctx)
	case *parser.ImportDeclContext:
		return v.VisitImportDecl(ctx)
	case *parser.ExternDeclContext:
		return v.VisitExternDecl(ctx)
	case *parser.ExternFunctionDeclContext:
		return v.VisitExternFunctionDecl(ctx)
	case *parser.FunctionDeclContext:
		return v.VisitFunctionDecl(ctx)
	case *parser.StructDeclContext:
		return v.VisitStructDecl(ctx)
	case *parser.BlockContext:
		return v.VisitBlock(ctx)
	case *parser.StatementContext:
		return v.VisitStatement(ctx)
	case *parser.VariableDeclContext:
		return v.VisitVariableDecl(ctx)
	case *parser.ConstDeclContext:
		return v.VisitConstDecl(ctx)
	case *parser.AssignmentStmtContext:
		return v.VisitAssignmentStmt(ctx)
	case *parser.ReturnStmtContext:
		return v.VisitReturnStmt(ctx)
	case *parser.IfStmtContext:
		return v.VisitIfStmt(ctx)
	case *parser.DeferStmtContext:
		return v.VisitDeferStmt(ctx)
	case *parser.ExpressionStmtContext:
		return v.VisitExpressionStmt(ctx)
	case *parser.ExpressionContext:
		return v.VisitExpression(ctx)
	case *parser.LogicalOrExpressionContext:
		return v.VisitLogicalOrExpression(ctx)
	case *parser.LogicalAndExpressionContext:
		return v.VisitLogicalAndExpression(ctx)
	case *parser.EqualityExpressionContext:
		return v.VisitEqualityExpression(ctx)
	case *parser.RelationalExpressionContext:
		return v.VisitRelationalExpression(ctx)
	case *parser.AdditiveExpressionContext:
		return v.VisitAdditiveExpression(ctx)
	case *parser.MultiplicativeExpressionContext:
		return v.VisitMultiplicativeExpression(ctx)
	case *parser.UnaryExpressionContext:
		return v.VisitUnaryExpression(ctx)
	case *parser.PostfixExpressionContext:
		return v.VisitPostfixExpression(ctx)
	case *parser.PrimaryExpressionContext:
		return v.VisitPrimaryExpression(ctx)
	case *parser.LiteralContext:
		return v.VisitLiteral(ctx)
	case *parser.CastExpressionContext:
		return v.VisitCastExpression(ctx)
	case *parser.AllocaExpressionContext:
		return v.VisitAllocaExpression(ctx)
	case *parser.ArgumentListContext:
		return v.VisitArgumentList(ctx)
	case *parser.LeftHandSideContext:
		return v.VisitLeftHandSide(ctx)
	default:
		// For unhandled types, try the default visitor behavior
		return v.BaseArcParserVisitor.Visit(tree)
	}
}

// ============================================================================
// COMPILATION UNIT
// ============================================================================

func (v *IRVisitor) VisitCompilationUnit(ctx *parser.CompilationUnitContext) interface{} {
	// Visit namespace declarations
	for _, ns := range ctx.AllNamespaceDecl() {
		v.Visit(ns)
	}
	
	// Visit imports
	for _, imp := range ctx.AllImportDecl() {
		v.Visit(imp)
	}
	
	// Visit top-level declarations
	for _, decl := range ctx.AllTopLevelDecl() {
		v.Visit(decl)
	}
	
	return nil
}

// VisitTopLevelDecl handles the intermediate TopLevelDecl node
func (v *IRVisitor) VisitTopLevelDecl(ctx *parser.TopLevelDeclContext) interface{} {
	// Dispatch to the actual declaration type
	if ctx.FunctionDecl() != nil {
		return v.Visit(ctx.FunctionDecl())
	}
	if ctx.StructDecl() != nil {
		return v.Visit(ctx.StructDecl())
	}
	if ctx.ExternDecl() != nil {
		return v.Visit(ctx.ExternDecl())
	}
	if ctx.ConstDecl() != nil {
		return v.Visit(ctx.ConstDecl())
	}
	if ctx.VariableDecl() != nil {
		return v.Visit(ctx.VariableDecl())
	}
	return nil
}

// VisitNamespaceDecl handles namespace declarations (currently just noted)
func (v *IRVisitor) VisitNamespaceDecl(ctx *parser.NamespaceDeclContext) interface{} {
	// Namespace handling could be added here if needed
	return nil
}

// VisitImportDecl handles import declarations (currently just noted)
func (v *IRVisitor) VisitImportDecl(ctx *parser.ImportDeclContext) interface{} {
	// Import handling could be added here if needed
	return nil
}

// ============================================================================
// EXTERN DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitExternDecl(ctx *parser.ExternDeclContext) interface{} {
	// Visit all extern members
	for _, member := range ctx.AllExternMember() {
		v.Visit(member)
	}
	return nil
}

func (v *IRVisitor) VisitExternFunctionDecl(ctx *parser.ExternFunctionDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	
	// Get return type
	var retType types.Type = types.Void
	if ctx.Type_() != nil {
		retType = v.resolveType(ctx.Type_())
	}
	
	// Get parameters
	paramTypes := make([]types.Type, 0)
	variadic := false
	
	if ctx.ExternParameterList() != nil {
		paramCtx := ctx.ExternParameterList()
		
		// Check for variadic
		if paramCtx.ELLIPSIS() != nil {
			variadic = true
		}
		
		// Get parameter types
		for _, typeCtx := range paramCtx.AllType_() {
			paramTypes = append(paramTypes, v.resolveType(typeCtx))
		}
	}
	
	// Declare external function
	v.ctx.Builder.DeclareFunction(name, retType, paramTypes, variadic)
	
	return nil
}

// ============================================================================
// FUNCTION DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitFunctionDecl(ctx *parser.FunctionDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	
	// Get return type
	var retType types.Type = types.Void
	if ctx.Type_() != nil {
		retType = v.resolveType(ctx.Type_())
	}
	
	// Get parameters
	paramTypes := make([]types.Type, 0)
	paramNames := make([]string, 0)
	variadic := false
	
	if ctx.ParameterList() != nil {
		paramCtx := ctx.ParameterList()
		
		// Check for variadic
		if paramCtx.ELLIPSIS() != nil {
			variadic = true
		}
		
		// Get parameters
		for _, param := range paramCtx.AllParameter() {
			paramNames = append(paramNames, param.IDENTIFIER().GetText())
			paramTypes = append(paramTypes, v.resolveType(param.Type_()))
		}
	}
	
	// Create function
	fn := v.ctx.Builder.CreateFunction(name, retType, paramTypes, variadic)
	
	// Set parameter names
	for i, paramName := range paramNames {
		fn.Arguments[i].SetName(paramName)
	}
	
	// Enter function context
	v.ctx.EnterFunction(fn)
	
	// Visit function body
	if ctx.Block() != nil {
		// Create entry block
		entry := v.ctx.Builder.CreateBlock("entry")
		v.ctx.SetInsertBlock(entry)
		
		v.Visit(ctx.Block())
		
		// Ensure block is terminated
		if entry.Terminator() == nil {
			if retType.Kind() == types.VoidKind {
				v.ctx.Builder.CreateRetVoid()
			} else {
				// Return zero value if no explicit return
				zero := v.getZeroValue(retType)
				v.ctx.Builder.CreateRet(zero)
			}
		}
	}
	
	// Exit function context
	v.ctx.ExitFunction()
	
	return nil
}

// ============================================================================
// STRUCT DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitStructDecl(ctx *parser.StructDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	
	// Get fields
	fieldTypes := make([]types.Type, 0)
	
	for _, field := range ctx.AllStructField() {
		fieldType := v.resolveType(field.Type_())
		fieldTypes = append(fieldTypes, fieldType)
	}
	
	// Create struct type
	structType := types.NewStruct(name, fieldTypes, false)
	
	// Register type
	v.ctx.RegisterType(name, structType)
	
	return nil
}

// ============================================================================
// STATEMENTS
// ============================================================================

// VisitStatement handles the intermediate Statement node
func (v *IRVisitor) VisitStatement(ctx *parser.StatementContext) interface{} {
	// Dispatch to the actual statement type
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
	v.ctx.PushScope()
	
	for _, stmt := range ctx.AllStatement() {
		v.Visit(stmt)
		
		// Stop processing if block is already terminated
		if v.ctx.currentBlock != nil && v.ctx.currentBlock.Terminator() != nil {
			break
		}
	}
	
	v.ctx.PopScope()
	return nil
}

func (v *IRVisitor) VisitVariableDecl(ctx *parser.VariableDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	
	// Get type (if specified)
	var varType types.Type
	if ctx.Type_() != nil {
		varType = v.resolveType(ctx.Type_())
	}
	
	// Evaluate initializer
	var initValue ir.Value
	if ctx.Expression() != nil {
		initValue = v.Visit(ctx.Expression()).(ir.Value)
		
		// Infer type from initializer if not specified
		if varType == nil {
			varType = initValue.Type()
		}
	} else {
		// No initializer, use zero value
		if varType == nil {
			v.ctx.Diagnostics.Error(fmt.Sprintf("variable '%s' needs type annotation or initializer", name))
			return nil
		}
		initValue = v.getZeroValue(varType)
	}
	
	// SSA-style: Store the VALUE directly in the symbol table
	// NO alloca, NO store - just pure SSA!
	v.ctx.currentScope.Define(name, initValue)
	
	return nil
}

func (v *IRVisitor) VisitConstDecl(ctx *parser.ConstDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	
	// Constants must have initializers
	if ctx.Expression() == nil {
		v.ctx.Diagnostics.Error(fmt.Sprintf("constant '%s' must have an initializer", name))
		return nil
	}
	
	// Evaluate initializer
	initValue := v.Visit(ctx.Expression()).(ir.Value)
	
	// SSA-style: Constants are just immutable SSA values - no memory needed!
	v.ctx.currentScope.DefineConst(name, initValue)
	
	return nil
}

func (v *IRVisitor) VisitAssignmentStmt(ctx *parser.AssignmentStmtContext) interface{} {
	// Get variable name from left-hand side
	lhsCtx := ctx.LeftHandSide()
	if lhsCtx.IDENTIFIER() != nil {
		name := lhsCtx.IDENTIFIER().GetText()
		
		// Get new value
		rhs := v.Visit(ctx.Expression()).(ir.Value)
		
		// Check if variable exists
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return nil
		}
		
		// Check if it's const
		if sym.IsConst {
			v.ctx.Diagnostics.Error(fmt.Sprintf("cannot assign to constant '%s'", name))
			return nil
		}
		
		// In SSA, we shadow the variable with a new value
		// This creates a new SSA value in the current scope
		v.ctx.currentScope.Define(name, rhs)
		
		return nil
	}
	
	// Handle pointer dereference assignment
	if lhsCtx.STAR() != nil {
		ptr := v.Visit(lhsCtx.Expression()).(ir.Value)
		rhs := v.Visit(ctx.Expression()).(ir.Value)
		v.ctx.Builder.CreateStore(rhs, ptr)
		return nil
	}
	
	v.ctx.Diagnostics.Error("complex assignment not yet supported")
	return nil
}

func (v *IRVisitor) VisitReturnStmt(ctx *parser.ReturnStmtContext) interface{} {
	// Handle deferred statements
	deferred := v.ctx.GetDeferredStmts()
	for i := len(deferred) - 1; i >= 0; i-- {
		// Execute deferred statements in reverse order
		// Note: This is simplified - proper defer needs more sophisticated handling
		_ = deferred[i]
	}
	
	if ctx.Expression() != nil {
		retVal := v.Visit(ctx.Expression()).(ir.Value)
		v.ctx.Builder.CreateRet(retVal)
	} else {
		v.ctx.Builder.CreateRetVoid()
	}
	
	return nil
}

func (v *IRVisitor) VisitIfStmt(ctx *parser.IfStmtContext) interface{} {
	// Get condition
	cond := v.Visit(ctx.Expression(0)).(ir.Value)
	
	// Create blocks
	thenBlock := v.ctx.Builder.CreateBlock("if.then")
	mergeBlock := v.ctx.Builder.CreateBlock("if.end")
	
	var elseBlock *ir.BasicBlock
	hasElse := len(ctx.AllELSE()) > 0
	
	if hasElse {
		elseBlock = v.ctx.Builder.CreateBlock("if.else")
		v.ctx.Builder.CreateCondBr(cond, thenBlock, elseBlock)
	} else {
		v.ctx.Builder.CreateCondBr(cond, thenBlock, mergeBlock)
	}
	
	// Then block
	v.ctx.SetInsertBlock(thenBlock)
	v.Visit(ctx.Block(0))
	if thenBlock.Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}
	
	// Else block
	if hasElse {
		v.ctx.SetInsertBlock(elseBlock)
		// Handle else-if or else
		if len(ctx.AllIF()) > 1 {
			// This is else-if, recursively handle
			// Simplified: just visit the else block
			v.Visit(ctx.Block(1))
		} else {
			v.Visit(ctx.Block(1))
		}
		if elseBlock.Terminator() == nil {
			v.ctx.Builder.CreateBr(mergeBlock)
		}
	}
	
	// Continue in merge block
	v.ctx.SetInsertBlock(mergeBlock)
	
	return nil
}

func (v *IRVisitor) VisitDeferStmt(ctx *parser.DeferStmtContext) interface{} {
	// For now, we'll store the statement context
	// Proper defer implementation requires more sophisticated handling
	// including unwinding and cleanup
	
	// Visit the expression/call to be deferred
	if ctx.Expression() != nil {
		_ = v.Visit(ctx.Expression())
	}
	
	// TODO: Implement proper defer mechanism
	v.ctx.Diagnostics.Warning("defer statement is not fully implemented yet")
	
	return nil
}

func (v *IRVisitor) VisitExpressionStmt(ctx *parser.ExpressionStmtContext) interface{} {
	v.Visit(ctx.Expression())
	return nil
}

// ============================================================================
// EXPRESSIONS
// ============================================================================

func (v *IRVisitor) VisitExpression(ctx *parser.ExpressionContext) interface{} {
	return v.Visit(ctx.LogicalOrExpression())
}

func (v *IRVisitor) VisitLogicalOrExpression(ctx *parser.LogicalOrExpressionContext) interface{} {
	result := v.Visit(ctx.LogicalAndExpression(0)).(ir.Value)
	
	for i := 1; i < len(ctx.AllLogicalAndExpression()); i++ {
		rhs := v.Visit(ctx.LogicalAndExpression(i)).(ir.Value)
		result = v.ctx.Builder.CreateOr(result, rhs, "")
	}
	
	return result
}

func (v *IRVisitor) VisitLogicalAndExpression(ctx *parser.LogicalAndExpressionContext) interface{} {
	result := v.Visit(ctx.EqualityExpression(0)).(ir.Value)
	
	for i := 1; i < len(ctx.AllEqualityExpression()); i++ {
		rhs := v.Visit(ctx.EqualityExpression(i)).(ir.Value)
		result = v.ctx.Builder.CreateAnd(result, rhs, "")
	}
	
	return result
}

func (v *IRVisitor) VisitEqualityExpression(ctx *parser.EqualityExpressionContext) interface{} {
	result := v.Visit(ctx.RelationalExpression(0)).(ir.Value)
	
	for i := 1; i < len(ctx.AllRelationalExpression()); i++ {
		rhs := v.Visit(ctx.RelationalExpression(i)).(ir.Value)
		
		if i-1 < len(ctx.AllEQ()) {
			result = v.ctx.Builder.CreateICmpEQ(result, rhs, "")
		} else {
			result = v.ctx.Builder.CreateICmpNE(result, rhs, "")
		}
	}
	
	return result
}

func (v *IRVisitor) VisitRelationalExpression(ctx *parser.RelationalExpressionContext) interface{} {
	result := v.Visit(ctx.AdditiveExpression(0)).(ir.Value)
	
	for i := 1; i < len(ctx.AllAdditiveExpression()); i++ {
		rhs := v.Visit(ctx.AdditiveExpression(i)).(ir.Value)
		
		// Determine which operator
		if i-1 < len(ctx.AllLT()) {
			result = v.ctx.Builder.CreateICmpSLT(result, rhs, "")
		} else if i-1-len(ctx.AllLT()) < len(ctx.AllLE()) {
			result = v.ctx.Builder.CreateICmpSLE(result, rhs, "")
		} else if i-1-len(ctx.AllLT())-len(ctx.AllLE()) < len(ctx.AllGT()) {
			result = v.ctx.Builder.CreateICmpSGT(result, rhs, "")
		} else {
			result = v.ctx.Builder.CreateICmpSGE(result, rhs, "")
		}
	}
	
	return result
}

func (v *IRVisitor) VisitAdditiveExpression(ctx *parser.AdditiveExpressionContext) interface{} {
	result := v.Visit(ctx.MultiplicativeExpression(0)).(ir.Value)
	
	for i := 1; i < len(ctx.AllMultiplicativeExpression()); i++ {
		rhs := v.Visit(ctx.MultiplicativeExpression(i)).(ir.Value)
		
		if i-1 < len(ctx.AllPLUS()) {
			result = v.ctx.Builder.CreateAdd(result, rhs, "")
		} else {
			result = v.ctx.Builder.CreateSub(result, rhs, "")
		}
	}
	
	return result
}

func (v *IRVisitor) VisitMultiplicativeExpression(ctx *parser.MultiplicativeExpressionContext) interface{} {
	result := v.Visit(ctx.UnaryExpression(0)).(ir.Value)
	
	for i := 1; i < len(ctx.AllUnaryExpression()); i++ {
		rhs := v.Visit(ctx.UnaryExpression(i)).(ir.Value)
		
		if i-1 < len(ctx.AllSTAR()) {
			result = v.ctx.Builder.CreateMul(result, rhs, "")
		} else if i-1-len(ctx.AllSTAR()) < len(ctx.AllSLASH()) {
			result = v.ctx.Builder.CreateSDiv(result, rhs, "")
		} else {
			result = v.ctx.Builder.CreateSRem(result, rhs, "")
		}
	}
	
	return result
}

func (v *IRVisitor) VisitUnaryExpression(ctx *parser.UnaryExpressionContext) interface{} {
	if ctx.MINUS() != nil {
		val := v.Visit(ctx.UnaryExpression()).(ir.Value)
		zero := v.getZeroValue(val.Type())
		return v.ctx.Builder.CreateSub(zero, val, "")
	}
	
	if ctx.NOT() != nil {
		val := v.Visit(ctx.UnaryExpression()).(ir.Value)
		return v.ctx.Builder.CreateXor(val, v.ctx.Builder.ConstInt(types.I1, 1), "")
	}
	
	if ctx.STAR() != nil {
		// Dereference
		ptr := v.Visit(ctx.UnaryExpression()).(ir.Value)
		ptrType := ptr.Type().(*types.PointerType)
		return v.ctx.Builder.CreateLoad(ptrType.ElementType, ptr, "")
	}
	
	if ctx.AMP() != nil {
		// Address-of operator
		// For SSA values, we need to create an alloca and store the value
		val := v.Visit(ctx.UnaryExpression()).(ir.Value)
		alloca := v.ctx.Builder.CreateAlloca(val.Type(), "")
		v.ctx.Builder.CreateStore(val, alloca)
		return alloca
	}
	
	return v.Visit(ctx.PostfixExpression())
}

func (v *IRVisitor) VisitPostfixExpression(ctx *parser.PostfixExpressionContext) interface{} {
	result := v.Visit(ctx.PrimaryExpression()).(ir.Value)
	
	for _, op := range ctx.AllPostfixOp() {
		result = v.visitPostfixOp(result, op.(*parser.PostfixOpContext))
	}
	
	return result
}

func (v *IRVisitor) visitPostfixOp(base ir.Value, ctx *parser.PostfixOpContext) ir.Value {
	if ctx.DOT() != nil && ctx.IDENTIFIER() != nil {
		// Field access or method call
		fieldName := ctx.IDENTIFIER().GetText()
		
		if ctx.LPAREN() != nil {
			// Method call
			// TODO: Implement method dispatch
			v.ctx.Diagnostics.Error("method calls not yet implemented")
			return base
		} else {
			// Field access
			// Get struct type
			var basePtr ir.Value
			if ptrType, ok := base.Type().(*types.PointerType); ok {
				basePtr = base
				if structType, ok := ptrType.ElementType.(*types.StructType); ok {
					// Find field index
					fieldIdx := v.findFieldIndex(structType, fieldName)
					if fieldIdx < 0 {
						v.ctx.Diagnostics.Error(fmt.Sprintf("struct has no field '%s'", fieldName))
						return base
					}
					
					return v.ctx.Builder.CreateStructGEP(structType, basePtr, fieldIdx, fieldName)
				}
			}
			
			v.ctx.Diagnostics.Error("field access requires struct pointer")
			return base
		}
	}
	
	if ctx.LPAREN() != nil {
		// Function call
		var args []ir.Value
		if ctx.ArgumentList() != nil {
			args = v.Visit(ctx.ArgumentList()).([]ir.Value)
		}
		
		// Get function - base should be a function
		if fn, ok := base.(*ir.Function); ok {
			return v.ctx.Builder.CreateCall(fn, args, "")
		}
		
		v.ctx.Diagnostics.Error("cannot call non-function")
		return base
	}
	
	return base
}

func (v *IRVisitor) VisitPrimaryExpression(ctx *parser.PrimaryExpressionContext) interface{} {
	if ctx.IDENTIFIER() != nil {
		name := ctx.IDENTIFIER().GetText()
		
		// Look up symbol
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			// Check if it's a function
			if fn := v.ctx.Module.GetFunction(name); fn != nil {
				return fn
			}
			
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}
		
		// Just return the SSA value directly - no loading!
		// The value is already an SSA value, not a pointer
		return sym.Value
	}
	
	if ctx.Literal() != nil {
		return v.Visit(ctx.Literal())
	}
	
	if ctx.Expression() != nil {
		return v.Visit(ctx.Expression())
	}
	
	if ctx.CastExpression() != nil {
		return v.Visit(ctx.CastExpression())
	}
	
	if ctx.AllocaExpression() != nil {
		return v.Visit(ctx.AllocaExpression())
	}
	
	if ctx.StructLiteral() != nil {
		return v.Visit(ctx.StructLiteral())
	}
	
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitLiteral(ctx *parser.LiteralContext) interface{} {
	if ctx.INTEGER_LITERAL() != nil {
		text := ctx.INTEGER_LITERAL().GetText()
		val, _ := strconv.ParseInt(text, 0, 64)
		// Use I64 for integer literals (matches int64 type)
		return v.ctx.Builder.ConstInt(types.I64, val)
	}
	
	if ctx.FLOAT_LITERAL() != nil {
		text := ctx.FLOAT_LITERAL().GetText()
		val, _ := strconv.ParseFloat(text, 64)
		return v.ctx.Builder.ConstFloat(types.F64, val)
	}
	
	if ctx.BOOLEAN_LITERAL() != nil {
		if ctx.BOOLEAN_LITERAL().GetText() == "true" {
			return v.ctx.Builder.True()
		}
		return v.ctx.Builder.False()
	}
	
	if ctx.STRING_LITERAL() != nil {
		// TODO: Implement string literals as global constants
		v.ctx.Diagnostics.Warning("string literals not fully implemented")
		return v.ctx.Builder.ConstInt(types.I64, 0)
	}
	
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitCastExpression(ctx *parser.CastExpressionContext) interface{} {
	val := v.Visit(ctx.Expression()).(ir.Value)
	destType := v.resolveType(ctx.Type_())
	
	// Determine cast type
	srcType := val.Type()
	
	if types.IsInteger(srcType) && types.IsInteger(destType) {
		srcInt := srcType.(*types.IntType)
		destInt := destType.(*types.IntType)
		
		if destInt.BitWidth > srcInt.BitWidth {
			if srcInt.Signed {
				return v.ctx.Builder.CreateSExt(val, destType, "")
			}
			return v.ctx.Builder.CreateZExt(val, destType, "")
		} else if destInt.BitWidth < srcInt.BitWidth {
			return v.ctx.Builder.CreateTrunc(val, destType, "")
		}
	}
	
	if types.IsInteger(srcType) && types.IsFloat(destType) {
		if srcType.(*types.IntType).Signed {
			return v.ctx.Builder.CreateSIToFP(val, destType, "")
		}
		return v.ctx.Builder.CreateUIToFP(val, destType, "")
	}
	
	if types.IsFloat(srcType) && types.IsInteger(destType) {
		if destType.(*types.IntType).Signed {
			return v.ctx.Builder.CreateFPToSI(val, destType, "")
		}
		return v.ctx.Builder.CreateFPToUI(val, destType, "")
	}
	
	// Default: bitcast
	return v.ctx.Builder.CreateBitCast(val, destType, "")
}

func (v *IRVisitor) VisitAllocaExpression(ctx *parser.AllocaExpressionContext) interface{} {
	allocType := v.resolveType(ctx.Type_())
	
	if ctx.Expression() != nil {
		// Array allocation
		count := v.Visit(ctx.Expression()).(ir.Value)
		return v.ctx.Builder.CreateAllocaWithCount(allocType, count, "")
	}
	
	return v.ctx.Builder.CreateAlloca(allocType, "")
}

func (v *IRVisitor) VisitStructLiteral(ctx *parser.StructLiteralContext) interface{} {
	// TODO: Implement struct literal construction
	// This would typically involve creating a temporary struct value or alloca
	// and populating fields.
	v.ctx.Diagnostics.Warning("struct literals not fully implemented")
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitArgumentList(ctx *parser.ArgumentListContext) interface{} {
	args := make([]ir.Value, 0)
	
	for _, expr := range ctx.AllExpression() {
		arg := v.Visit(expr).(ir.Value)
		args = append(args, arg)
	}
	
	return args
}

func (v *IRVisitor) VisitLeftHandSide(ctx *parser.LeftHandSideContext) interface{} {
	// In the new SSA visitor, variable assignment is handled directly in VisitAssignmentStmt.
	// VisitLeftHandSide is primarily used for obtaining memory addresses for pointer operations
	// (e.g., &x, *p = val).

	if ctx.IDENTIFIER() != nil {
		name := ctx.IDENTIFIER().GetText()
		
		// In SSA, we generally don't get the address of a local variable unless it
		// was explicitly allocated with `alloca` (which creates a pointer type variable).
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}
		
		// If the value is already a pointer (explicit allocation), return it.
		// If it's a direct SSA value, we technically can't take its address 
		// without spilling it to the stack first.
		if _, isPtr := sym.Value.Type().(*types.PointerType); isPtr {
			return sym.Value
		}
		
		v.ctx.Diagnostics.Error(fmt.Sprintf("cannot take address of SSA value '%s'", name))
		return sym.Value
	}
	
	if ctx.STAR() != nil {
		// Dereference pattern: *ptr
		// The expression inside must evaluate to a pointer
		ptr := v.Visit(ctx.Expression()).(ir.Value)
		return ptr
	}
	
	if ctx.DOT() != nil {
		// Field access: obj.field
		// This requires the object to be a pointer to a struct
		// Implementation depends on how GEP is handled in your IR builder
		v.ctx.Diagnostics.Error("field assignment address not fully implemented")
		return v.ctx.Builder.ConstInt(types.I64, 0)
	}
	
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func (v *IRVisitor) resolveType(ctx parser.ITypeContext) types.Type {
	if ctx == nil {
		return types.Void
	}
	
	typeCtx := ctx.(*parser.TypeContext)
	
	if typeCtx.PrimitiveType() != nil {
		name := typeCtx.PrimitiveType().GetText()
		if typ, ok := v.ctx.GetType(name); ok {
			return typ
		}
		// Standardize default fallback to I64 for this new visitor
		v.ctx.Diagnostics.Error(fmt.Sprintf("unknown type: %s", name))
		return types.I64
	}
	
	if typeCtx.PointerType() != nil {
		elemType := v.resolveType(typeCtx.PointerType().Type_())
		return types.NewPointer(elemType)
	}
	
	if typeCtx.ReferenceType() != nil {
		// References are treated as pointers in IR
		elemType := v.resolveType(typeCtx.ReferenceType().Type_())
		return types.NewPointer(elemType)
	}
	
	if typeCtx.IDENTIFIER() != nil {
		name := typeCtx.IDENTIFIER().GetText()
		if typ, ok := v.ctx.GetType(name); ok {
			return typ
		}
		v.ctx.Diagnostics.Error(fmt.Sprintf("unknown type: %s", name))
		return types.I64
	}
	
	return types.I64
}

func (v *IRVisitor) getZeroValue(typ types.Type) ir.Value {
	switch typ.Kind() {
	case types.IntegerKind:
		return v.ctx.Builder.ConstInt(typ.(*types.IntType), 0)
	case types.FloatKind:
		return v.ctx.Builder.ConstFloat(typ.(*types.FloatType), 0.0)
	case types.PointerKind:
		return v.ctx.Builder.ConstNull(typ.(*types.PointerType))
	default:
		return v.ctx.Builder.ConstZero(typ)
	}
}

func (v *IRVisitor) findFieldIndex(structType *types.StructType, fieldName string) int {
	// This logic assumes StructType has a way to look up fields or we iterate types
	// Since the exact definition of types.Struct isn't visible, we assume standard index lookup.
	// In a real implementation, you might need to iterate `structType.Fields` to match the name.
	
	// Placeholder logic:
	v.ctx.Diagnostics.Warning(fmt.Sprintf("field name resolution not fully implemented for '%s'", fieldName))
	return 0
}