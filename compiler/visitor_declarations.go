package compiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// ============================================================================
// IMPORT DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitImportDecl(ctx *parser.ImportDeclContext) interface{} {
	// 1. Get the import path string (remove quotes)
	rawPath := ctx.STRING_LITERAL().GetText()
	importPath := strings.Trim(rawPath, "\"")

	fmt.Printf("DEBUG: Processing import: %s from %s\n", importPath, v.currentFile)

	// 2. Resolve absolute directory path
	currentDir := filepath.Dir(v.currentFile)
	absPath, err := v.ctx.Importer.ResolvePath(currentDir, importPath)
	if err != nil {
		v.ctx.Diagnostics.Error(fmt.Sprintf("import resolution failed: %v", err))
		return nil
	}

	// 3. Compile that package (recursively)
	pkgInfo, err := v.compiler.CompilePackage(absPath) 
	if err != nil {
		v.ctx.Diagnostics.Error(fmt.Sprintf("failed to compile package %s: %v", importPath, err))
		return nil
	}

	fmt.Printf("DEBUG: Successfully imported package '%s' (namespace: %s)\n", importPath, pkgInfo.Name)
	return nil
}

// ============================================================================
// EXTERN DECLARATIONS
// ============================================================================

func (v *IRVisitor) VisitExternDecl(ctx *parser.ExternDeclContext) interface{} {
	var namespaceName string
	
	if ctx.IDENTIFIER() != nil {
		namespaceName = ctx.IDENTIFIER().GetText()
		fmt.Printf("DEBUG: Processing extern namespace: %s\n", namespaceName)
		
		// Temporarily switch namespace for these externs
		oldNamespace := v.ctx.currentNamespace
		v.ctx.SetNamespace(namespaceName)
		
		// Process all extern members
		for _, member := range ctx.AllExternMember() {
			v.Visit(member)
		}
		
		// Restore namespace
		v.ctx.currentNamespace = oldNamespace
	} else {
		// No namespace, just global extern declarations
		for _, member := range ctx.AllExternMember() {
			v.Visit(member)
		}
	}
	
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
	
	// Register in current namespace
	// If current namespace is root, it goes to global scope
	// If it is a specific namespace, it is registered there
	if v.ctx.currentNamespace != nil {
		v.ctx.currentNamespace.Functions[name] = fn
		fmt.Printf("DEBUG: Declared extern function %s in namespace %s\n", name, v.ctx.currentNamespace.Name)
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
	
	// Check if this is a method inside a class/struct
	var methodPrefix string
	if parent := ctx.GetParent(); parent != nil {
		if classMember, ok := parent.(*parser.ClassMemberContext); ok {
			if classDecl, ok := classMember.GetParent().(*parser.ClassDeclContext); ok {
				className := classDecl.IDENTIFIER().GetText()
				methodPrefix = className + "_"
				name = methodPrefix + name
			}
		} else if structMember, ok := parent.(*parser.StructMemberContext); ok {
			if structDecl, ok := structMember.GetParent().(*parser.StructDeclContext); ok {
				structName := structDecl.IDENTIFIER().GetText()
				methodPrefix = structName + "_"
				name = methodPrefix + name
			}
		}
	}
	
	// Handle Namespacing
	// If we are in a namespace (e.g. "utils"), the function name in IR becomes "utils_Func"
	// but it is stored in the Namespace map as "Func".
	// The LLVM IR name should be unique.
	var irName string = name
	
	// Special Case: The main function in the main namespace should NOT be mangled
	// This ensures the linker can find the entry point.
	isMain := name == "main" && (v.ctx.currentNamespace == nil || v.ctx.currentNamespace.Name == "main" || v.ctx.currentNamespace.Name == "")
	
	if !isMain && v.ctx.currentNamespace != nil && v.ctx.currentNamespace.Name != "" {
		// Check if it already has a prefix (methods might not need namespace prefix if type is unique?)
		// For now, namespace prefixing for everything inside a namespace
		irName = v.ctx.currentNamespace.Name + "_" + name
	}

	fmt.Printf("DEBUG VisitFunctionDecl: Declaring function: %s (IR: %s)\n", name, irName)
	
	var retType types.Type = types.Void
	if ctx.Type_() != nil {
		retType = v.resolveType(ctx.Type_())
	}
	
	paramTypes := make([]types.Type, 0)
	paramNames := make([]string, 0)
	variadic := false
	
	if ctx.ParameterList() != nil {
		paramCtx := ctx.ParameterList()
		if paramCtx.ELLIPSIS() != nil {
			variadic = true
		}
		for _, param := range paramCtx.AllParameter() {
			paramName := param.IDENTIFIER().GetText()
			paramType := v.resolveType(param.Type_())
			paramNames = append(paramNames, paramName)
			paramTypes = append(paramTypes, paramType)
		}
	}
	
	fn := v.ctx.Builder.CreateFunction(irName, retType, paramTypes, variadic)
	
	// Register function in the current namespace
	if v.ctx.currentNamespace != nil {
		v.ctx.currentNamespace.Functions[name] = fn
	}

	for i, paramName := range paramNames {
		fn.Arguments[i].SetName(paramName)
	}
	
	v.ctx.EnterFunction(fn)
	
	if ctx.Block() != nil {
		entry := v.ctx.Builder.CreateBlock("entry")
		v.ctx.SetInsertBlock(entry)
		
		// Allocate space for parameters and store them
		for i, arg := range fn.Arguments {
			alloc := v.ctx.Builder.CreateAlloca(arg.Type(), paramNames[i]+".addr")
			v.ctx.Builder.CreateStore(arg, alloc)
			v.ctx.currentScope.Define(paramNames[i], alloc)
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
	
	v.ctx.ExitFunction()
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