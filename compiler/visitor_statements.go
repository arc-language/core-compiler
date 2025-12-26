package compiler

import (
	"fmt"

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
	v.ctx.PushScope()
	
	for _, stmt := range ctx.AllStatement() {
		v.Visit(stmt)
		
		// Stop if we hit a terminator
		if v.ctx.currentBlock != nil && v.ctx.currentBlock.Terminator() != nil {
			break
		}
	}
	
	v.ctx.PopScope()
	return nil
}

func (v *IRVisitor) VisitAssignmentStmt(ctx *parser.AssignmentStmtContext) interface{} {
	lhsCtx := ctx.LeftHandSide()
	
	// Variable Assignment
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
		
		if ptr, isAlloca := sym.Value.(*ir.AllocaInst); isAlloca {
			v.ctx.Builder.CreateStore(rhs, ptr)
			return nil
		}
		
		v.ctx.currentScope.Define(name, rhs)
		return nil
	}
	
	// Pointer Assignment (*ptr = val)
	if lhsCtx.STAR() != nil {
		ptr := v.Visit(lhsCtx.Expression()).(ir.Value)
		rhs := v.Visit(ctx.Expression()).(ir.Value)
		v.ctx.Builder.CreateStore(rhs, ptr)
		return nil
	}
	
	// Field Assignment (obj.field = val)
	if lhsCtx.DOT() != nil {
		exprCtx := lhsCtx.Expression()
		var basePtr ir.Value
		
		// Attempt to resolve the variable to get its address (L-Value)
		exprText := exprCtx.GetText()
		if sym, ok := v.ctx.currentScope.Lookup(exprText); ok {
			if alloca, isAlloca := sym.Value.(*ir.AllocaInst); isAlloca {
				// Only use the alloca address if it is a Struct
				if _, isStruct := alloca.AllocatedType.(*types.StructType); isStruct {
					basePtr = alloca
				}
			}
		}
		
		if basePtr == nil {
			// Fallback: This correctly handles pointers-to-structs
			basePtr = v.Visit(exprCtx).(ir.Value)
		}
		
		fieldName := lhsCtx.IDENTIFIER().GetText()
		
		// Check pointer type to find the struct definition
		if ptrType, ok := basePtr.Type().(*types.PointerType); ok {
			if structType, ok := ptrType.ElementType.(*types.StructType); ok {
				fieldIdx := v.findFieldIndex(structType, fieldName)
				if fieldIdx >= 0 {
					gep := v.ctx.Builder.CreateStructGEP(structType, basePtr, fieldIdx, "")
					rhs := v.Visit(ctx.Expression()).(ir.Value)
					v.ctx.Builder.CreateStore(rhs, gep)
					return nil
				}
			}
		}
		
		v.ctx.Diagnostics.Error("cannot assign to field (invalid struct pointer or unknown field)")
		return nil
	}

	v.ctx.Diagnostics.Error("complex assignment not yet supported")
	return nil
}

func (v *IRVisitor) VisitReturnStmt(ctx *parser.ReturnStmtContext) interface{} {
	fmt.Printf("DEBUG VisitReturnStmt: Has expression: %v\n", ctx.Expression() != nil)
	if ctx.Expression() != nil {
		fmt.Printf("DEBUG VisitReturnStmt: Expression text: %s\n", ctx.Expression().GetText())
	}
	
	// Execute deferred statements
	deferred := v.ctx.GetDeferredStmts()
	for i := len(deferred) - 1; i >= 0; i-- {
		_ = deferred[i]
	}
	
	if ctx.Expression() != nil {
		fmt.Printf("DEBUG VisitReturnStmt: About to visit return expression\n")
		retVal := v.Visit(ctx.Expression()).(ir.Value)
		fmt.Printf("DEBUG VisitReturnStmt: Return value type: %v\n", retVal.Type())
		
		// Cast to expected return type if needed
		if v.ctx.currentFunction != nil {
			expectedType := v.ctx.currentFunction.FuncType.ReturnType
			if !retVal.Type().Equal(expectedType) {
				retVal = v.castValue(retVal, expectedType)
			}
		}
		
		v.ctx.Builder.CreateRet(retVal)
		fmt.Printf("DEBUG VisitReturnStmt: Created return instruction\n")
	} else {
		v.ctx.Builder.CreateRetVoid()
		fmt.Printf("DEBUG VisitReturnStmt: Created void return\n")
	}
	
	fmt.Printf("DEBUG VisitReturnStmt: completed\n")
	return nil
}

func (v *IRVisitor) VisitExpressionStmt(ctx *parser.ExpressionStmtContext) interface{} {
	fmt.Printf("DEBUG VisitExpressionStmt: expression text = %s\n", ctx.Expression().GetText())
	v.Visit(ctx.Expression())
	fmt.Printf("DEBUG VisitExpressionStmt: completed\n")
	return nil
}

func (v *IRVisitor) VisitDeferStmt(ctx *parser.DeferStmtContext) interface{} {
	if ctx.Expression() != nil {
		_ = v.Visit(ctx.Expression())
	}
	v.ctx.Diagnostics.Warning("defer statement is not fully implemented yet")
	return nil
}

func (v *IRVisitor) VisitLeftHandSide(ctx *parser.LeftHandSideContext) interface{} {
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