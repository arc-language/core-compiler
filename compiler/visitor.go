// visitor_new.go
package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// IRVisitor implements the ANTLR visitor pattern to generate IR
type IRVisitor struct {
	*parser.BaseArcParserVisitor
	ctx *Context

	// Namespace tracking for externs
	// map[namespaceName]map[funcName]*ir.Function
	namespaces       map[string]map[string]*ir.Function
	currentNamespace string
}

// NewIRVisitor creates a new IR visitor
func NewIRVisitor(ctx *Context) *IRVisitor {
	return &IRVisitor{
		BaseArcParserVisitor: &parser.BaseArcParserVisitor{},
		ctx:                  ctx,
		namespaces:           make(map[string]map[string]*ir.Function),
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
    case *parser.ExternMemberContext:
        return v.VisitExternMember(ctx)
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

func (v *IRVisitor) VisitNamespaceDecl(ctx *parser.NamespaceDeclContext) interface{} {
	return nil
}

func (v *IRVisitor) VisitImportDecl(ctx *parser.ImportDeclContext) interface{} {
	return nil
}

// ============================================================================
// EXTERN DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitExternDecl(ctx *parser.ExternDeclContext) interface{} {
	oldNamespace := v.currentNamespace
	
	// Handle named extern blocks (e.g. "extern io { ... }")
	if ctx.IDENTIFIER() != nil {
		nsName := ctx.IDENTIFIER().GetText()
		v.currentNamespace = nsName
		
		// Initialize the map for this namespace
		if _, exists := v.namespaces[nsName]; !exists {
			v.namespaces[nsName] = make(map[string]*ir.Function)
		}

		// Define the namespace in scope so we can find it later
		// We use a dummy Global with a special prefix to identify it as a namespace
		dummyGlobal := &ir.Global{}
		dummyGlobal.SetName("namespace:" + nsName)
		v.ctx.currentScope.Define(nsName, dummyGlobal)
	}

	// Visit all extern members
	for _, member := range ctx.AllExternMember() {
		v.Visit(member)
	}

	v.currentNamespace = oldNamespace
	return nil
}

func (v *IRVisitor) VisitExternMember(ctx *parser.ExternMemberContext) interface{} {
	if ctx.ExternFunctionDecl() != nil {
		return v.Visit(ctx.ExternFunctionDecl())
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

	// Declare external function in the module
	// IMPORTANT: We declare it with the *real* name (e.g., "printf") 
	// so the linker can find it.
	fn := v.ctx.Builder.DeclareFunction(name, retType, paramTypes, variadic)

	// If we are inside an extern namespace (e.g. "io"), register it there
	if v.currentNamespace != "" {
		v.namespaces[v.currentNamespace][name] = fn
	} else {
		// Otherwise, it's global scope
		// (DeclareFunction already adds it to module, but we might want it in scope too)
		v.ctx.currentScope.Define(name, fn)
	}

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

		if paramCtx.ELLIPSIS() != nil {
			variadic = true
		}

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
		entry := v.ctx.Builder.CreateBlock("entry")
		v.ctx.SetInsertBlock(entry)

		// Define arguments in scope
		// In SSA, arguments are values themselves
		for i, arg := range fn.Arguments {
			v.ctx.currentScope.Define(paramNames[i], arg)
		}

		v.Visit(ctx.Block())

		// Ensure block is terminated
		if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
			if retType.Kind() == types.VoidKind {
				v.ctx.Builder.CreateRetVoid()
			} else {
				zero := v.getZeroValue(retType)
				v.ctx.Builder.CreateRet(zero)
			}
		}
	}

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
		if v.ctx.currentBlock != nil && v.ctx.currentBlock.Terminator() != nil {
			break
		}
	}

	v.ctx.PopScope()
	return nil
}

func (v *IRVisitor) VisitVariableDecl(ctx *parser.VariableDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()

	var varType types.Type
	if ctx.Type_() != nil {
		varType = v.resolveType(ctx.Type_())
	}

	var initValue ir.Value
	if ctx.Expression() != nil {
		initValue = v.Visit(ctx.Expression()).(ir.Value)

		if varType == nil {
			varType = initValue.Type()
		}
	} else {
		if varType == nil {
			v.ctx.Diagnostics.Error(fmt.Sprintf("variable '%s' needs type annotation or initializer", name))
			return nil
		}
		initValue = v.getZeroValue(varType)
	}

	// SSA: Store value directly in symbol table
	v.ctx.currentScope.Define(name, initValue)

	return nil
}

func (v *IRVisitor) VisitConstDecl(ctx *parser.ConstDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()

	if ctx.Expression() == nil {
		v.ctx.Diagnostics.Error(fmt.Sprintf("constant '%s' must have an initializer", name))
		return nil
	}

	initValue := v.Visit(ctx.Expression()).(ir.Value)
	v.ctx.currentScope.DefineConst(name, initValue)

	return nil
}

func (v *IRVisitor) VisitAssignmentStmt(ctx *parser.AssignmentStmtContext) interface{} {
	lhsCtx := ctx.LeftHandSide()
	
	// Variable reassignment (SSA Shadowing)
	if lhsCtx.IDENTIFIER() != nil {
		name := lhsCtx.IDENTIFIER().GetText()
		rhs := v.Visit(ctx.Expression()).(ir.Value)

		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return nil
		}
		if sym.IsConst {
			v.ctx.Diagnostics.Error(fmt.Sprintf("cannot assign to constant '%s'", name))
			return nil
		}

		v.ctx.currentScope.Define(name, rhs)
		return nil
	}

	// Pointer/Memory assignment (*ptr = val)
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
	deferred := v.ctx.GetDeferredStmts()
	for i := len(deferred) - 1; i >= 0; i-- {
		_ = deferred[i]
	}

	if ctx.Expression() != nil {
		retVal := v.Visit(ctx.Expression()).(ir.Value)
		
		// Implicit Cast: Check function return type
		// FIX: Use 'currentFunction' (lowercase) instead of 'CurrentFunction'
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

func (v *IRVisitor) VisitIfStmt(ctx *parser.IfStmtContext) interface{} {
	mergeBlock := v.ctx.Builder.CreateBlock("if.end")
	
	// Setup chain
	cond := v.Visit(ctx.Expression(0)).(ir.Value)
	thenBlock := v.ctx.Builder.CreateBlock("if.then")
	nextCheckBlock := v.ctx.Builder.CreateBlock("if.next")
	
	v.ctx.Builder.CreateCondBr(cond, thenBlock, nextCheckBlock)
	
	// 1. Generate First THEN
	v.ctx.SetInsertBlock(thenBlock)
	v.Visit(ctx.Block(0))
	if thenBlock.Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}
	
	// 2. Iterate ELSE IFs
	v.ctx.SetInsertBlock(nextCheckBlock)
	count := len(ctx.AllIF())
	
	for i := 1; i < count; i++ {
		cond := v.Visit(ctx.Expression(i)).(ir.Value)
		
		thenBlock := v.ctx.Builder.CreateBlock("elseif.then")
		newNextBlock := v.ctx.Builder.CreateBlock("elseif.next")
		
		v.ctx.Builder.CreateCondBr(cond, thenBlock, newNextBlock)
		
		v.ctx.SetInsertBlock(thenBlock)
		v.Visit(ctx.Block(i))
		if thenBlock.Terminator() == nil {
			v.ctx.Builder.CreateBr(mergeBlock)
		}
		
		v.ctx.SetInsertBlock(newNextBlock)
	}
	
	// 3. Handle Final ELSE
	if len(ctx.AllBlock()) > count {
		v.Visit(ctx.Block(count))
	}
	
	// Jump to merge if not terminated
	if v.ctx.currentBlock.Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}
	
	v.ctx.SetInsertBlock(mergeBlock)
	return nil
}

func (v *IRVisitor) VisitDeferStmt(ctx *parser.DeferStmtContext) interface{} {
	if ctx.Expression() != nil {
		_ = v.Visit(ctx.Expression())
	}
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
		ptr := v.Visit(ctx.UnaryExpression()).(ir.Value)
		ptrType, ok := ptr.Type().(*types.PointerType)
		if !ok {
			v.ctx.Diagnostics.Error("cannot dereference non-pointer")
			return ptr
		}
		return v.ctx.Builder.CreateLoad(ptrType.ElementType, ptr, "")
	}
	if ctx.AMP() != nil {
		// SSA: Can only take address if explicitly allocated (like alloca)
		// For now, this is tricky in pure SSA without alloca. 
		// We fallback to standard visit which might error if it's not a pointer-backed value.
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
		fieldName := ctx.IDENTIFIER().GetText()

		// 1. Namespace Lookup (io.printf)
		if global, ok := base.(*ir.Global); ok && strings.HasPrefix(global.Name(), "namespace:") {
			nsName := strings.TrimPrefix(global.Name(), "namespace:")
			
			if funcs, ok := v.namespaces[nsName]; ok {
				if fn, ok := funcs[fieldName]; ok {
					return fn
				}
			}
			v.ctx.Diagnostics.Error(fmt.Sprintf("function '%s' not found in namespace '%s'", fieldName, nsName))
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}

		// 2. Struct Field Access
		if ctx.LPAREN() == nil {
			// Needs pointer to struct
			if ptrType, ok := base.Type().(*types.PointerType); ok {
				if structType, ok := ptrType.ElementType.(*types.StructType); ok {
					fieldIdx := v.findFieldIndex(structType, fieldName)
					if fieldIdx < 0 {
						v.ctx.Diagnostics.Error(fmt.Sprintf("struct has no field '%s'", fieldName))
						return base
					}
					return v.ctx.Builder.CreateStructGEP(structType, base, fieldIdx, fieldName)
				}
			}
			v.ctx.Diagnostics.Error("field access requires struct pointer")
			return base
		}
	}

	// Function Call
	if ctx.LPAREN() != nil {
		var args []ir.Value
		if ctx.ArgumentList() != nil {
			args = v.Visit(ctx.ArgumentList()).([]ir.Value)
		}

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
		
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			// Check global functions
			if fn := v.ctx.Module.GetFunction(name); fn != nil {
				return fn
			}
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}
		
		// If variable is an explicitly allocated pointer, load it
		// (Unless we are in a context that wants the address, handled elsewhere)
		// But in SSA visitor, we usually store Values directly. 
		// Only 'var' with explicit addressable requirements (like array) uses pointers.
		// For now, assume Values are SSA unless Type is PointerType AND name matches alloca logic.
		// Actually, in this visitor, we just return the value stored in scope.
		
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
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitLiteral(ctx *parser.LiteralContext) interface{} {
	if ctx.INTEGER_LITERAL() != nil {
		text := ctx.INTEGER_LITERAL().GetText()
		val, _ := strconv.ParseInt(text, 0, 64)
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

	// IMPLEMENTATION OF STRING LITERALS
	if ctx.STRING_LITERAL() != nil {
		rawText := ctx.STRING_LITERAL().GetText()
		
		// Unquote the string (converts "\n" to actual newline, etc.)
		content, err := strconv.Unquote(rawText)
		if err != nil {
			// Fallback if parsing fails (shouldn't happen with valid tokens)
			if len(rawText) >= 2 {
				content = rawText[1 : len(rawText)-1]
			} else {
				content = rawText
			}
		}

		// Append NULL terminator (C-style strings required for printf)
		bytes := append([]byte(content), 0)

		// Create IR constants for the array elements
		elements := make([]ir.Constant, len(bytes))
		for i, b := range bytes {
			elements[i] = v.ctx.Builder.ConstInt(types.I8, int64(b))
		}

		// Create the array type: [N x i8]
		arrType := types.NewArray(types.I8, int64(len(bytes)))

		// Create the constant array value
		constArr := &ir.ConstantArray{
			BaseValue: ir.BaseValue{ValType: arrType},
			Elements:  elements,
		}

		// Create a global constant to hold the string bytes
		// Name it .str.N to avoid collisions
		strName := fmt.Sprintf(".str.%d", len(v.ctx.Module.Globals))
		global := v.ctx.Builder.CreateGlobalConstant(strName, constArr)

		// Decay the array to a pointer (i8*) using GEP
		// This converts [N x i8]* -> i8* which printf expects (*byte)
		zero := v.ctx.Builder.ConstInt(types.I32, 0)
		gep := v.ctx.Builder.CreateInBoundsGEP(
			arrType,
			global,
			[]ir.Value{zero, zero},
			"",
		)

		return gep
	}

	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitCastExpression(ctx *parser.CastExpressionContext) interface{} {
	// 1. Evaluate the expression being cast
	val := v.Visit(ctx.Expression()).(ir.Value)
	
	// 2. Resolve the target type
	destType := v.resolveType(ctx.Type_())
	srcType := val.Type()

	// 3. Pointer <-> Integer Casting (Required for NULL checks like 'file == 0')
	if types.IsPointer(srcType) && types.IsInteger(destType) {
		return v.ctx.Builder.CreatePtrToInt(val, destType, "")
	}
	if types.IsInteger(srcType) && types.IsPointer(destType) {
		return v.ctx.Builder.CreateIntToPtr(val, destType, "")
	}

	// 4. Integer Resizing (Truncate / Extend)
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
		// BitWidths are equal, no-op or bitcast
		if srcInt.Signed != destInt.Signed {
			return v.ctx.Builder.CreateBitCast(val, destType, "")
		}
		return val
	}

	// 5. Integer <-> Float Conversions
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

	// 6. Float Resizing
	if types.IsFloat(srcType) && types.IsFloat(destType) {
		srcFloat := srcType.(*types.FloatType)
		destFloat := destType.(*types.FloatType)
		
		if destFloat.BitWidth > srcFloat.BitWidth {
			return v.ctx.Builder.CreateFPExt(val, destType, "")
		} else if destFloat.BitWidth < srcFloat.BitWidth {
			return v.ctx.Builder.CreateFPTrunc(val, destType, "")
		}
		return val
	}

	// 7. Default Fallback (Bitcast)
	// Useful for pointer casting (ptr<i8> -> ptr<i32>)
	return v.ctx.Builder.CreateBitCast(val, destType, "")
}

func (v *IRVisitor) VisitAllocaExpression(ctx *parser.AllocaExpressionContext) interface{} {
	allocType := v.resolveType(ctx.Type_())
	if ctx.Expression() != nil {
		count := v.Visit(ctx.Expression()).(ir.Value)
		return v.ctx.Builder.CreateAllocaWithCount(allocType, count, "")
	}
	return v.ctx.Builder.CreateAlloca(allocType, "")
}

func (v *IRVisitor) VisitArgumentList(ctx *parser.ArgumentListContext) interface{} {
	args := make([]ir.Value, 0)
	for _, expr := range ctx.AllExpression() {
		args = append(args, v.Visit(expr).(ir.Value))
	}
	return args
}

func (v *IRVisitor) VisitLeftHandSide(ctx *parser.LeftHandSideContext) interface{} {
	// Not used in SSA assignment path, but good for pointer math
	if ctx.IDENTIFIER() != nil {
		name := ctx.IDENTIFIER().GetText()
		sym, ok := v.ctx.currentScope.Lookup(name)
		if ok {
			return sym.Value
		}
	}
	if ctx.STAR() != nil {
		return v.Visit(ctx.Expression())
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
		// Fallback for default int/float names if not in Context
		switch name {
		case "int": return types.I32
		case "int32": return types.I32
		case "int64": return types.I64
		case "float": return types.F64
		case "byte": return types.I8
		case "bool": return types.I1
		case "void": return types.Void
		}
		return types.I64
	}

	if typeCtx.PointerType() != nil {
		elemType := v.resolveType(typeCtx.PointerType().Type_())
		return types.NewPointer(elemType)
	}

	if typeCtx.IDENTIFIER() != nil {
		name := typeCtx.IDENTIFIER().GetText()
		if typ, ok := v.ctx.GetType(name); ok {
			return typ
		}
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
	// In a real implementation, you'd store field names in the StructType or a separate symbol table.
	// Since types.StructType here doesn't have field names (just []Type), we can't implement this properly 
	// without changing the type definition.
	// We'll return 0 as a placeholder or panic.
	v.ctx.Diagnostics.Warning(fmt.Sprintf("Cannot look up field '%s' by name - structs only store types.", fieldName))
	return 0
}

func (v *IRVisitor) castValue(val ir.Value, targetType types.Type) ir.Value {
	srcType := val.Type()

	// Integer resizing
	if types.IsInteger(srcType) && types.IsInteger(targetType) {
		srcBits := srcType.(*types.IntType).BitWidth
		destBits := targetType.(*types.IntType).BitWidth

		if srcBits > destBits {
			return v.ctx.Builder.CreateTrunc(val, targetType, "")
		} else if srcBits < destBits {
			return v.ctx.Builder.CreateSExt(val, targetType, "")
		}
	}
	
	// Default to no-op if types match or unhandled
	return val
}