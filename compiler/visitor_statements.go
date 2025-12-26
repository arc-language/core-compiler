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
	stmts := ctx.AllStatement()
	fmt.Printf("DEBUG VisitBlock: %d statements\n", len(stmts))
	
	for i, stmt := range stmts {
		stmtText := stmt.GetText()
		if len(stmtText) > 100 {
			stmtText = stmtText[:100] + "..."
		}
		fmt.Printf("DEBUG VisitBlock: Statement %d text: %s\n", i, stmtText)
	}
	
	v.ctx.PushScope()
	
	for i, stmt := range stmts {
		fmt.Printf("DEBUG VisitBlock: Processing statement %d/%d\n", i+1, len(stmts))
		v.Visit(stmt)
		
		// Stop if we hit a terminator
		if v.ctx.currentBlock != nil && v.ctx.currentBlock.Terminator() != nil {
			fmt.Printf("DEBUG VisitBlock: Hit terminator, stopping at statement %d\n", i)
			break
		}
	}
	
	v.ctx.PopScope()
	fmt.Printf("DEBUG VisitBlock: completed\n")
	return nil
}

func (v *IRVisitor) VisitAssignmentStmt(ctx *parser.AssignmentStmtContext) interface{} {
	lhsCtx := ctx.LeftHandSide()
	
	fmt.Printf("DEBUG VisitAssignmentStmt:\n")
	fmt.Printf("  LHS has IDENTIFIER: %v\n", lhsCtx.IDENTIFIER() != nil)
	if lhsCtx.IDENTIFIER() != nil {
		fmt.Printf("  LHS IDENTIFIER: %s\n", lhsCtx.IDENTIFIER().GetText())
	}
	fmt.Printf("  LHS has STAR: %v\n", lhsCtx.STAR() != nil)
	fmt.Printf("  LHS has DOT: %v\n", lhsCtx.DOT() != nil)
	fmt.Printf("  LHS has Expression: %v\n", lhsCtx.Expression() != nil)
	fmt.Printf("  RHS expression: %s\n", ctx.Expression().GetText())
	
	// Variable Assignment
	if lhsCtx.IDENTIFIER() != nil && lhsCtx.DOT() == nil && lhsCtx.STAR() == nil {
		name := lhsCtx.IDENTIFIER().GetText()
		fmt.Printf("DEBUG: Simple variable assignment to: %s\n", name)
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
		fmt.Printf("DEBUG: Pointer dereference assignment\n")
		ptr := v.Visit(lhsCtx.Expression()).(ir.Value)
		rhs := v.Visit(ctx.Expression()).(ir.Value)
		v.ctx.Builder.CreateStore(rhs, ptr)
		return nil
	}
	
	// Field Assignment (obj.field = val)
	if lhsCtx.DOT() != nil {
		fmt.Printf("DEBUG: Field assignment\n")
		exprCtx := lhsCtx.Expression()
		var basePtr ir.Value
		
		// Attempt to resolve the variable to get its address (L-Value)
		exprText := exprCtx.GetText()
		fmt.Printf("DEBUG: Base expression text: %s\n", exprText)
		
		if sym, ok := v.ctx.currentScope.Lookup(exprText); ok {
			fmt.Printf("DEBUG: Found symbol for base: %s\n", exprText)
			if alloca, isAlloca := sym.Value.(*ir.AllocaInst); isAlloca {
				// Only use the alloca address if it is a Struct
				if _, isStruct := alloca.AllocatedType.(*types.StructType); isStruct {
					basePtr = alloca
					fmt.Printf("DEBUG: Using alloca as base pointer\n")
				}
			}
		}
		
		if basePtr == nil {
			fmt.Printf("DEBUG: Fallback - visiting expression for base\n")
			// Fallback: This correctly handles pointers-to-structs
			basePtr = v.Visit(exprCtx).(ir.Value)
			fmt.Printf("DEBUG: Base pointer type: %v\n", basePtr.Type())
		}
		
		fieldName := lhsCtx.IDENTIFIER().GetText()
		fmt.Printf("DEBUG: Field name: %s\n", fieldName)
		
		// Check pointer type to find the struct definition
		if ptrType, ok := basePtr.Type().(*types.PointerType); ok {
			fmt.Printf("DEBUG: Base is pointer type\n")
			if structType, ok := ptrType.ElementType.(*types.StructType); ok {
				fmt.Printf("DEBUG: Element is struct type: %s\n", structType.Name)
				
				// Check if this is a class type
				isClass := v.ctx.IsClassType(structType.Name)
				fmt.Printf("DEBUG: Is class type: %v\n", isClass)
				
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
				
				fmt.Printf("DEBUG: Field index: %d\n", fieldIdx)
				
				if fieldIdx >= 0 {
					gep := v.ctx.Builder.CreateStructGEP(structType, basePtr, fieldIdx, "")
					rhs := v.Visit(ctx.Expression()).(ir.Value)
					v.ctx.Builder.CreateStore(rhs, gep)
					fmt.Printf("DEBUG: Field assignment completed\n")
					return nil
				} else {
					v.ctx.Diagnostics.Error(fmt.Sprintf("struct/class '%s' has no field '%s'", structType.Name, fieldName))
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
	
	// Check if this looks like an assignment that wasn't parsed as such
	exprText := ctx.Expression().GetText()
	if strings.Contains(exprText, "=") && !strings.Contains(exprText, "==") && !strings.Contains(exprText, "!=") {
		fmt.Printf("WARNING: Expression contains '=' - might be a failed assignment parse: %s\n", exprText)
	}
	
	result := v.Visit(ctx.Expression())
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