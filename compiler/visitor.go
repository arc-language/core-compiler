package compiler

import (
	"fmt"
	"strconv"

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
	var retType types.Type = types.Void  // Changed from retType := types.Void
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
	var allocType types.Type
	if ctx.Type_() != nil {
		allocType = v.resolveType(ctx.Type_())
	}
	
	// Evaluate initializer
	var initValue ir.Value
	if ctx.Expression() != nil {
		initValue = v.Visit(ctx.Expression()).(ir.Value)
		
		// Infer type from initializer if not specified
		if allocType == nil {
			allocType = initValue.Type()
		}
	} else {
		// No initializer, use zero value
		if allocType == nil {
			v.ctx.Diagnostics.Error(fmt.Sprintf("variable '%s' needs type annotation or initializer", name))
			return nil
		}
	}
	
	// Allocate stack space
	alloca := v.ctx.Builder.CreateAlloca(allocType, name)
	
	// Store initial value
	if initValue != nil {
		v.ctx.Builder.CreateStore(initValue, alloca)
	} else {
		zero := v.getZeroValue(allocType)
		v.ctx.Builder.CreateStore(zero, alloca)
	}
	
	// Add to symbol table
	v.ctx.currentScope.Define(name, alloca)
	
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
	
	// For now, treat constants as immutable stack allocations
	allocType := initValue.Type()
	alloca := v.ctx.Builder.CreateAlloca(allocType, name)
	v.ctx.Builder.CreateStore(initValue, alloca)
	
	// Add to symbol table as const
	v.ctx.currentScope.DefineConst(name, alloca)
	
	return nil
}

func (v *IRVisitor) VisitAssignmentStmt(ctx *parser.AssignmentStmtContext) interface{} {
	// Get left-hand side (should be a pointer)
	lhs := v.Visit(ctx.LeftHandSide()).(ir.Value)
	
	// Get right-hand side value
	rhs := v.Visit(ctx.Expression()).(ir.Value)
	
	// Store value
	v.ctx.Builder.CreateStore(rhs, lhs)
	
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
		// Address-of - the value should already be a pointer (lvalue)
		return v.Visit(ctx.UnaryExpression())
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
			return v.ctx.Builder.ConstInt(types.I32, 0)
		}
		
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
	
	return v.ctx.Builder.ConstInt(types.I32, 0)
}

func (v *IRVisitor) VisitLiteral(ctx *parser.LiteralContext) interface{} {
	if ctx.INTEGER_LITERAL() != nil {
		text := ctx.INTEGER_LITERAL().GetText()
		val, _ := strconv.ParseInt(text, 0, 64)
		return v.ctx.Builder.ConstInt(types.I32, val)
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
		return v.ctx.Builder.ConstInt(types.I32, 0)
	}
	
	return v.ctx.Builder.ConstInt(types.I32, 0)
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
		return v.ctx.Builder.CreateSIToFP(val, destType, "")
	}
	
	if types.IsFloat(srcType) && types.IsInteger(destType) {
		return v.ctx.Builder.CreateFPToSI(val, destType, "")
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

func (v *IRVisitor) VisitArgumentList(ctx *parser.ArgumentListContext) interface{} {
	args := make([]ir.Value, 0)
	
	for _, expr := range ctx.AllExpression() {
		arg := v.Visit(expr).(ir.Value)
		args = append(args, arg)
	}
	
	return args
}

func (v *IRVisitor) VisitLeftHandSide(ctx *parser.LeftHandSideContext) interface{} {
	// Check for different left-hand side patterns
	if ctx.IDENTIFIER() != nil {
		name := ctx.IDENTIFIER().GetText()
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return v.ctx.Builder.ConstInt(types.I32, 0)
		}
		return sym.Value
	}
	
	if ctx.STAR() != nil {
		// Dereference pattern: *ptr = value
		ptr := v.Visit(ctx.Expression()).(ir.Value)
		return ptr
	}
	
	if ctx.DOT() != nil {
		// Field access: obj.field = value
		// This would need more sophisticated handling
		v.ctx.Diagnostics.Error("field assignment not fully implemented")
		return v.ctx.Builder.ConstInt(types.I32, 0)
	}
	
	return v.ctx.Builder.ConstInt(types.I32, 0)
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
		v.ctx.Diagnostics.Error(fmt.Sprintf("unknown type: %s", name))
		return types.I32
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
		return types.I32
	}
	
	return types.I32
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
	// This is simplified - in a real compiler, you'd maintain field name mappings
	// For now, we can't easily map field names to indices without additional metadata
	v.ctx.Diagnostics.Warning(fmt.Sprintf("field name resolution not fully implemented for '%s'", fieldName))
	return 0
}