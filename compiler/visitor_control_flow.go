package compiler

import (
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-parser"
)

func (v *IRVisitor) VisitIfStmt(ctx *parser.IfStmtContext) interface{} {
	mergeBlock := v.ctx.Builder.CreateBlock("if.end")
	
	// First if condition
	cond := v.Visit(ctx.Expression(0)).(ir.Value)
	thenBlock := v.ctx.Builder.CreateBlock("if.then")
	nextCheckBlock := v.ctx.Builder.CreateBlock("if.next")
	
	v.ctx.Builder.CreateCondBr(cond, thenBlock, nextCheckBlock)
	
	// Then block
	v.ctx.SetInsertBlock(thenBlock)
	v.Visit(ctx.Block(0))
	if thenBlock.Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}
	
	// Handle else-if and else
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
	
	// Final else block (if present)
	if len(ctx.AllBlock()) > count {
		v.Visit(ctx.Block(count))
	}
	
	if v.ctx.currentBlock.Terminator() == nil {
		v.ctx.Builder.CreateBr(mergeBlock)
	}
	
	// ONLY set insert point to merge block if it has predecessors
	// (i.e., if any branch actually jumps to it)
	if len(mergeBlock.Predecessors) > 0 {
		v.ctx.SetInsertBlock(mergeBlock)
	}
	// Otherwise, leave the insert point wherever it is (likely unreachable)
	
	return nil
}

func (v *IRVisitor) VisitForStmt(ctx *parser.ForStmtContext) interface{} {
	v.ctx.PushScope()
	defer v.ctx.PopScope()

	// Check for for-in loop (iteration)
	if ctx.IN() != nil {
		return v.visitForInLoop(ctx)
	}

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

	condBlock := v.ctx.Builder.CreateBlock("loop.cond")
	bodyBlock := v.ctx.Builder.CreateBlock("loop.body")
	postBlock := v.ctx.Builder.CreateBlock("loop.post")
	endBlock := v.ctx.Builder.CreateBlock("loop.end")

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
	// TODO: Implement for-in loop iteration
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