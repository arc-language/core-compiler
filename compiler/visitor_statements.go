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