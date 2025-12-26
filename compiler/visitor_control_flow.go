package compiler

import (
	"fmt"

	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-parser"
)

func (v *IRVisitor) VisitIfStmt(ctx *parser.IfStmtContext) interface{} {
	// Generate unique suffix based on the source position (Line_Column).
	// This ensures a unique, deterministic ID for every if-statement.
	token := ctx.GetStart()
	uniqueID := fmt.Sprintf("%d_%d", token.GetLine(), token.GetColumn())

	mergeBlock := v.ctx.Builder.CreateBlock("if.end." + uniqueID)

	// First if condition
	cond := v.Visit(ctx.Expression(0)).(ir.Value)
	thenBlock := v.ctx.Builder.CreateBlock("if.then." + uniqueID)
	nextCheckBlock := v.ctx.Builder.CreateBlock("if.next." + uniqueID)

	v.ctx.Builder.CreateCondBr(cond, thenBlock, nextCheckBlock)

	// Then block
	v.ctx.SetInsertBlock(thenBlock)
	v.Visit(ctx.Block(0))
	
	// Ensure we don't double-terminate if the block already has a return/break
	if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}

	// Handle else-if and else
	v.ctx.SetInsertBlock(nextCheckBlock)
	count := len(ctx.AllIF())

	for i := 1; i < count; i++ {
		cond := v.Visit(ctx.Expression(i)).(ir.Value)
		
		// Use index 'i' to ensure unique block names for else-if chains
		thenName := fmt.Sprintf("elseif.then.%s.%d", uniqueID, i)
		nextName := fmt.Sprintf("elseif.next.%s.%d", uniqueID, i)
		
		thenBlock := v.ctx.Builder.CreateBlock(thenName)
		newNextBlock := v.ctx.Builder.CreateBlock(nextName)

		v.ctx.Builder.CreateCondBr(cond, thenBlock, newNextBlock)

		v.ctx.SetInsertBlock(thenBlock)
		v.Visit(ctx.Block(i))
		if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
			v.ctx.Builder.CreateBr(mergeBlock)
		}

		v.ctx.SetInsertBlock(newNextBlock)
	}

	// Final else block (if present)
	// The number of blocks is count+1 if there is an 'else' clause
	if len(ctx.AllBlock()) > count {
		v.Visit(ctx.Block(count))
	}

	if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}

	// Only set insert point to merge block if it has predecessors (is reachable).
	// If both 'if' and 'else' return, mergeBlock is unreachable and should be ignored.
	if len(mergeBlock.Predecessors) > 0 {
		v.ctx.SetInsertBlock(mergeBlock)
	}

	return nil
}

func (v *IRVisitor) VisitForStmt(ctx *parser.ForStmtContext) interface{} {
	v.ctx.PushScope()
	defer v.ctx.PopScope()

	// Check for for-in loop (iteration)
	if ctx.IN() != nil {
		return v.visitForInLoop(ctx)
	}

	// Use Line_Column for loop blocks too, to prevent collisions in nested loops
	token := ctx.GetStart()
	uniqueID := fmt.Sprintf("%d_%d", token.GetLine(), token.GetColumn())

	semicolons := ctx.AllSEMICOLON()
	isClause := len(semicolons) == 2

	// Initialize statement
	if isClause {
		if ctx.VariableDecl() != nil {
			v.Visit(ctx.VariableDecl())
		} else if len(ctx.AllAssignmentStmt()) > 0 {
			firstAssign := ctx.AssignmentStmt(0)
			semi1 := semicolons[0]
			if v.isBefore(firstAssign, semi1) {
				v.Visit(firstAssign)
			}
		}
	}

	condBlock := v.ctx.Builder.CreateBlock("loop.cond." + uniqueID)
	bodyBlock := v.ctx.Builder.CreateBlock("loop.body." + uniqueID)
	postBlock := v.ctx.Builder.CreateBlock("loop.post." + uniqueID)
	endBlock := v.ctx.Builder.CreateBlock("loop.end." + uniqueID)

	continueTarget := condBlock
	if isClause {
		continueTarget = postBlock
	}

	v.ctx.Builder.CreateBr(condBlock)

	// Condition block
	v.ctx.SetInsertBlock(condBlock)

	var cond ir.Value
	if isClause {
		semi1 := semicolons[0]
		semi2 := semicolons[1]
		found := false
		for _, expr := range ctx.AllExpression() {
			if v.isAfter(expr, semi1) && v.isBefore(expr, semi2) {
				cond = v.Visit(expr).(ir.Value)
				found = true
				break
			}
		}
		if !found {
			cond = v.ctx.Builder.True()
		}
	} else if ctx.Expression(0) != nil {
		cond = v.Visit(ctx.Expression(0)).(ir.Value)
	} else {
		cond = v.ctx.Builder.True()
	}

	v.ctx.Builder.CreateCondBr(cond, bodyBlock, endBlock)

	// Body block
	v.ctx.SetInsertBlock(bodyBlock)
	v.ctx.PushLoop(continueTarget, endBlock)
	v.Visit(ctx.Block())
	v.ctx.PopLoop()

	if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
		v.ctx.Builder.CreateBr(continueTarget)
	}

	// Post block
	v.ctx.SetInsertBlock(postBlock)
	if isClause {
		semi2 := semicolons[1]
		for _, assign := range ctx.AllAssignmentStmt() {
			if v.isAfter(assign, semi2) {
				v.Visit(assign)
			}
		}
		for _, expr := range ctx.AllExpression() {
			if v.isAfter(expr, semi2) {
				v.Visit(expr)
			}
		}
	}

	if v.ctx.Builder.GetInsertBlock().Terminator() == nil {
		v.ctx.Builder.CreateBr(condBlock)
	}

	v.ctx.SetInsertBlock(endBlock)
	return nil
}

func (v *IRVisitor) visitForInLoop(ctx *parser.ForStmtContext) interface{} {
	v.ctx.Diagnostics.Warning("for-in loops are not yet fully implemented")
	v.Visit(ctx.Block())
	return nil
}

func (v *IRVisitor) VisitBreakStmt(ctx *parser.BreakStmtContext) interface{} {
	loop := v.ctx.CurrentLoop()
	if loop == nil {
		v.ctx.Diagnostics.Error("break statement outside of loop")
		return nil
	}
	v.ctx.Builder.CreateBr(loop.BreakBlock)
	return nil
}

func (v *IRVisitor) VisitContinueStmt(ctx *parser.ContinueStmtContext) interface{} {
	loop := v.ctx.CurrentLoop()
	if loop == nil {
		v.ctx.Diagnostics.Error("continue statement outside of loop")
		return nil
	}
	v.ctx.Builder.CreateBr(loop.ContinueBlock)
	return nil
}