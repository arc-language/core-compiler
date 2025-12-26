package compiler

import (
	"fmt"

	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// ============================================================================
// EXTERN DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitExternDecl(ctx *parser.ExternDeclContext) interface{} {
	oldNamespace := v.currentNamespace
	
	if ctx.IDENTIFIER() != nil {
		nsName := ctx.IDENTIFIER().GetText()
		v.currentNamespace = nsName
		
		if _, exists := v.namespaces[nsName]; !exists {
			v.namespaces[nsName] = make(map[string]*ir.Function)
		}
		
		// Create a dummy global to represent the namespace
		dummyGlobal := &ir.Global{}
		dummyGlobal.SetName("namespace:" + nsName)
		v.ctx.currentScope.Define(nsName, dummyGlobal)
	}
	
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
	
	var retType types.Type = types.Void
	if ctx.Type_() != nil {
		retType = v.resolveType(ctx.Type_())
	}
	
	paramTypes := make([]types.Type, 0)
	variadic := false
	
	if ctx.ExternParameterList() != nil {
		paramCtx := ctx.ExternParameterList()
		if paramCtx.ELLIPSIS() != nil {
			variadic = true
		}
		for _, typeCtx := range paramCtx.AllType_() {
			paramTypes = append(paramTypes, v.resolveType(typeCtx))
		}
	}
	
	fn := v.ctx.Builder.DeclareFunction(name, retType, paramTypes, variadic)
	
	if v.currentNamespace != "" {
		v.namespaces[v.currentNamespace][name] = fn
	} else {
		v.ctx.currentScope.Define(name, fn)
	}
	
	return nil
}

// ============================================================================
// FUNCTION DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitFunctionDecl(ctx *parser.FunctionDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	
	var retType types.Type = types.Void
	if ctx.Type_() != nil {
		retType = v.resolveType(ctx.Type_())
	}
	
	paramTypes := make([]types.Type, 0)
	paramNames := make([]string, 0)
	variadic := false
	
	if ctx.ParameterList() != nil {
		paramCtx := ctx.ParameterList()
		if pafunc (v *IRVisitor) VisitFunctionDecl(ctx *parser.FunctionDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	fmt.Printf("DEBUG VisitFunctionDecl: Function name: %s\n", name)
	
	var retType types.Type = types.Void
	if ctx.Type_() != nil {
		retType = v.resolveType(ctx.Type_())
		fmt.Printf("DEBUG VisitFunctionDecl: Return type: %v\n", retType)
	}
	
	paramTypes := make([]types.Type, 0)
	paramNames := make([]string, 0)
	variadic := false
	
	if ctx.ParameterList() != nil {
		paramCtx := ctx.ParameterList()
		fmt.Printf("DEBUG VisitFunctionDecl: %d parameters\n", len(paramCtx.AllParameter()))
		if paramCtx.ELLIPSIS() != nil {
			variadic = true
		}
		for i, param := range paramCtx.AllParameter() {
			paramName := param.IDENTIFIER().GetText()
			paramType := v.resolveType(param.Type_())
			fmt.Printf("DEBUG VisitFunctionDecl: Param %d: %s : %v\n", i, paramName, paramType)
			paramNames = append(paramNames, paramName)
			paramTypes = append(paramTypes, paramType)
		}
	}
	
	fmt.Printf("DEBUG VisitFunctionDecl: Creating function with %d params\n", len(paramTypes))
	fn := v.ctx.Builder.CreateFunction(name, retType, paramTypes, variadic)
	
	for i, paramName := range paramNames {
		fn.Arguments[i].SetName(paramName)
	}
	
	fmt.Printf("DEBUG VisitFunctionDecl: Entering function context\n")
	v.ctx.EnterFunction(fn)
	
	if ctx.Block() != nil {
		fmt.Printf("DEBUG VisitFunctionDecl: Processing function body\n")
		entry := v.ctx.Builder.CreateBlock("entry")
		v.ctx.SetInsertBlock(entry)
		
		// Allocate space for parameters and store them
		for i, arg := range fn.Arguments {
			alloc := v.ctx.Builder.CreateAlloca(arg.Type(), paramNames[i]+".addr")
			v.ctx.Builder.CreateStore(arg, alloc)
			v.ctx.currentScope.Define(paramNames[i], alloc)
			fmt.Printf("DEBUG VisitFunctionDecl: Defined param %s\n", paramNames[i])
		}
		
		v.Visit(ctx.Block())
		
		// Add default return if needed
		if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
			if retType.Kind() == types.VoidKind {
				v.ctx.Builder.CreateRetVoid()
			} else {
				zero := v.getZeroValue(retType)
				v.ctx.Builder.CreateRet(zero)
			}
		}
	}
	
	fmt.Printf("DEBUG VisitFunctionDecl: Exiting function context\n")
	v.ctx.ExitFunction()
	fmt.Printf("DEBUG VisitFunctionDecl: Function %s completed\n", name)
	return nil
}

// ============================================================================
// VARIABLE & CONSTANT DECLARATIONS
// ============================================================================

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

	alloca := v.ctx.Builder.CreateAlloca(varType, name+".addr")
	v.ctx.Builder.CreateStore(initValue, alloca)
	v.ctx.currentScope.Define(name, alloca)
	
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