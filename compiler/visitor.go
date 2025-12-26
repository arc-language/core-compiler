package compiler

import (
	"fmt"

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
	namespaces       map[string]map[string]*ir.Function
	currentNamespace string
	
	// Method call tracking
	pendingMethodSelf ir.Value
	pendingMethodFunc *ir.Function
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
	case *parser.ClassDeclContext:
		return v.VisitClassDecl(ctx)
	case *parser.ClassMemberContext:
		return v.VisitClassMember(ctx)
	case *parser.ClassFieldContext:
		return v.VisitClassField(ctx)
	case *parser.DeinitDeclContext:
		return v.VisitDeinitDecl(ctx)
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
	case *parser.ForStmtContext:
		return v.VisitForStmt(ctx)
	case *parser.BreakStmtContext:
		return v.VisitBreakStmt(ctx)
	case *parser.ContinueStmtContext:
		return v.VisitContinueStmt(ctx)
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
	case *parser.RangeExpressionContext:
		return v.VisitRangeExpression(ctx)
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
	case *parser.StructLiteralContext:
		return v.VisitStructLiteral(ctx)
	case *parser.CastExpressionContext:
		return v.VisitCastExpression(ctx)
	case *parser.AllocaExpressionContext:
		return v.VisitAllocaExpression(ctx)
	case *parser.ArgumentListContext:
		return v.VisitArgumentList(ctx)
	case *parser.LeftHandSideContext:
		return v.VisitLeftHandSide(ctx)
	default:
		return v.BaseArcParserVisitor.Visit(tree)
	}
}

// ============================================================================
// COMPILATION UNIT & TOP LEVEL
// ============================================================================

func (v *IRVisitor) VisitCompilationUnit(ctx *parser.CompilationUnitContext) interface{} {
	fmt.Printf("DEBUG VisitCompilationUnit: Starting compilation\n")
	
	// Pass 1: Register all type declarations (structs and classes)
	fmt.Printf("DEBUG VisitCompilationUnit: Pass 1 - Registering types\n")
	for _, decl := range ctx.AllTopLevelDecl() {
		if decl.StructDecl() != nil {
			v.registerStructType(decl.StructDecl().(*parser.StructDeclContext))
		} else if decl.ClassDecl() != nil {
			v.registerClassType(decl.ClassDecl().(*parser.ClassDeclContext))
		}
	}
	
	// Pass 2: Process everything else
	fmt.Printf("DEBUG VisitCompilationUnit: Pass 2 - Processing declarations\n")
	for _, ns := range ctx.AllNamespaceDecl() {
		v.Visit(ns)
	}
	
	for _, imp := range ctx.AllImportDecl() {
		v.Visit(imp)
	}
	
	for i, decl := range ctx.AllTopLevelDecl() {
		fmt.Printf("DEBUG VisitCompilationUnit: Processing top-level decl %d\n", i)
		if decl.FunctionDecl() != nil {
			fmt.Printf("DEBUG VisitCompilationUnit: Decl %d is function\n", i)
			v.Visit(decl.FunctionDecl())
		} else if decl.ExternDecl() != nil {
			fmt.Printf("DEBUG VisitCompilationUnit: Decl %d is extern\n", i)
			v.Visit(decl.ExternDecl())
		} else if decl.ConstDecl() != nil {
			fmt.Printf("DEBUG VisitCompilationUnit: Decl %d is const\n", i)
			v.Visit(decl.ConstDecl())
		} else if decl.VariableDecl() != nil {
			fmt.Printf("DEBUG VisitCompilationUnit: Decl %d is variable\n", i)
			v.Visit(decl.VariableDecl())
		} else if decl.StructDecl() != nil {
			fmt.Printf("DEBUG VisitCompilationUnit: Decl %d is struct - processing methods\n", i)
			v.Visit(decl.StructDecl())
		} else if decl.ClassDecl() != nil {
			fmt.Printf("DEBUG VisitCompilationUnit: Decl %d is class - processing methods\n", i)
			v.Visit(decl.ClassDecl())
		}
	}
	
	fmt.Printf("DEBUG VisitCompilationUnit: Compilation complete\n")
	return nil
}

func (v *IRVisitor) VisitTopLevelDecl(ctx *parser.TopLevelDeclContext) interface{} {
	if ctx.FunctionDecl() != nil {
		return v.Visit(ctx.FunctionDecl())
	}
	if ctx.StructDecl() != nil {
		return v.Visit(ctx.StructDecl())
	}
	if ctx.ClassDecl() != nil {
		return v.Visit(ctx.ClassDecl())
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
	// Namespace declarations are metadata only
	return nil
}

func (v *IRVisitor) VisitImportDecl(ctx *parser.ImportDeclContext) interface{} {
	// Import declarations are handled by a separate module system
	return nil
}

// ============================================================================
// HELPERS
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
		return types.I64
	}
	
	if typeCtx.PointerType() != nil {
		elemType := v.resolveType(typeCtx.PointerType().Type_())
		return types.NewPointer(elemType)
	}
	
	if typeCtx.ReferenceType() != nil {
		elemType := v.resolveType(typeCtx.ReferenceType().Type_())
		return types.NewPointer(elemType)
	}
	
	if typeCtx.VectorType() != nil {
		// TODO: Implement vector types
		v.ctx.Diagnostics.Warning("vector types not yet implemented")
		return types.I64
	}
	
	if typeCtx.MapType() != nil {
		// TODO: Implement map types
		v.ctx.Diagnostics.Warning("map types not yet implemented")
		return types.I64
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
	if fieldIndices, ok := v.ctx.StructFieldIndices[structType.Name]; ok {
		if idx, ok := fieldIndices[fieldName]; ok {
			return idx
		}
	}
	return -1
}

func (v *IRVisitor) castValue(val ir.Value, targetType types.Type) ir.Value {
	srcType := val.Type()
	
	if types.IsInteger(srcType) && types.IsInteger(targetType) {
		srcBits := srcType.(*types.IntType).BitWidth
		destBits := targetType.(*types.IntType).BitWidth
		if srcBits > destBits {
			return v.ctx.Builder.CreateTrunc(val, targetType, "")
		} else if srcBits < destBits {
			return v.ctx.Builder.CreateSExt(val, targetType, "")
		}
	}
	
	return val
}