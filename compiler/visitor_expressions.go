package compiler

import (
	"fmt"
	"strconv"

	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

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
	result := v.Visit(ctx.RangeExpression(0)).(ir.Value)
	for i := 1; i < len(ctx.AllRangeExpression()); i++ {
		rhs := v.Visit(ctx.RangeExpression(i)).(ir.Value)
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

func (v *IRVisitor) VisitRangeExpression(ctx *parser.RangeExpressionContext) interface{} {
	return v.Visit(ctx.AdditiveExpression(0))
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
		return v.Visit(ctx.UnaryExpression())
	}
	
	return v.Visit(ctx.PostfixExpression())
}

func (v *IRVisitor) VisitPostfixExpression(ctx *parser.PostfixExpressionContext) interface{} {
	result := v.Visit(ctx.PrimaryExpression()).(ir.Value)
	
	// Track if we're starting with a namespace identifier
	var baseIdentifier string
	if primaryCtx := ctx.PrimaryExpression(); primaryCtx != nil {
		if primaryCtx.IDENTIFIER() != nil {
			baseIdentifier = primaryCtx.IDENTIFIER().GetText()
		}
	}
	
	for _, op := range ctx.AllPostfixOp() {
		result = v.visitPostfixOp(result, op.(*parser.PostfixOpContext), baseIdentifier)
		baseIdentifier = "" // Clear after first use
	}
	return result
}

func (v *IRVisitor) visitPostfixOp(base ir.Value, ctx *parser.PostfixOpContext, baseIdentifier string) ir.Value {
	// Function call (check this FIRST)
	if ctx.LPAREN() != nil {
		var args []ir.Value
		if ctx.ArgumentList() != nil {
			args = v.Visit(ctx.ArgumentList()).([]ir.Value)
		}
		
		// Check if this is a method call (we have a pending self parameter)
		if fn, ok := base.(*ir.Function); ok {
			// Prepend self parameter if we have one pending
			if v.pendingMethodSelf != nil {
				args = append([]ir.Value{v.pendingMethodSelf}, args...)
				v.pendingMethodSelf = nil
			}
			
			return v.ctx.Builder.CreateCall(fn, args, "")
		}
		
		v.ctx.Diagnostics.Error("cannot call non-function")
		return base
	}
	
	// Member access (DOT)
	if ctx.DOT() != nil && ctx.IDENTIFIER() != nil {
		memberName := ctx.IDENTIFIER().GetText()
		
		// Reset pending method state from any previous operation
		v.pendingMethodSelf = nil
		
		// 1. Check if this is a namespace.function access
		// Usage: utils.Func() -> baseIdentifier="utils", memberName="Func"
		if baseIdentifier != "" {
			// Lookup namespace in global context registry
			if ns, ok := v.ctx.NamespaceRegistry[baseIdentifier]; ok {
				if fn, ok := ns.LookupFunction(memberName); ok {
					return fn
				}
				v.ctx.Diagnostics.Error(fmt.Sprintf("function '%s' not found in namespace '%s'", memberName, baseIdentifier))
				return v.ctx.Builder.ConstInt(types.I64, 0)
			}
		}
		
		// 2. Not a namespace access - check for class method
		if ptrType, ok := base.Type().(*types.PointerType); ok {
			if structType, ok := ptrType.ElementType.(*types.StructType); ok {
				if v.ctx.IsClassType(structType.Name) {
					// Look for a method with the naming convention: ClassName_methodName
					// Also check if the class is part of a namespace
					// For now, simpler implementation: Class methods are just functions
					
					// Try Class_Member
					methodName := structType.Name + "_" + memberName
					
					// If we are in a namespace, checking module functions might need namespace prefix?
					// But usually, methods are looked up via the type name which is unique.
					if fn := v.ctx.Module.GetFunction(methodName); fn != nil {
						// Store the self pointer to be prepended when the function is called
						v.pendingMethodSelf = base
						return fn
					}
					
					// If the class was defined in a namespace, the method might be "Namespace_Class_Member"
					// This gets complicated without Type metadata storing the namespace.
				}
			}
		}
		
		// 3. Field access
		return v.handleFieldAccess(base, memberName)
	}
	
	return base
}

func (v *IRVisitor) handleFieldAccess(base ir.Value, fieldName string) ir.Value {
	// Case 1: Pointer to struct/class
	if ptrType, ok := base.Type().(*types.PointerType); ok {
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
			
			if fieldIdx < 0 {
				v.ctx.Diagnostics.Error(fmt.Sprintf("type '%s' has no field '%s'", structType.Name, fieldName))
				return base
			}
			
			gep := v.ctx.Builder.CreateStructGEP(structType, base, fieldIdx, "")
			return v.ctx.Builder.CreateLoad(structType.Fields[fieldIdx], gep, "")
		}
	}
	
	// Case 2: Struct value (direct value)
	if structType, ok := base.Type().(*types.StructType); ok {
		if v.ctx.IsClassType(structType.Name) {
			v.ctx.Diagnostics.Error("class instances must be accessed via pointer")
			return base
		}
		
		fieldIdx := v.findFieldIndex(structType, fieldName)
		if fieldIdx < 0 {
			v.ctx.Diagnostics.Error(fmt.Sprintf("struct has no field '%s'", fieldName))
			return base
		}
		return v.ctx.Builder.CreateExtractValue(base, []int{fieldIdx}, "")
	}
	
	v.ctx.Diagnostics.Error(fmt.Sprintf("field access requires struct or class instance"))
	return base
}

func (v *IRVisitor) VisitPrimaryExpression(ctx *parser.PrimaryExpressionContext) interface{} {
	if ctx.StructLiteral() != nil {
		return v.Visit(ctx.StructLiteral())
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

	if ctx.SyscallExpression() != nil {
		return v.Visit(ctx.SyscallExpression())
	}
	
	// Check identifier
	if ctx.IDENTIFIER() != nil {
		name := ctx.IDENTIFIER().GetText()
		
		// First check if this is a type name
		if _, isType := v.ctx.GetType(name); isType {
			v.ctx.Diagnostics.Error(fmt.Sprintf("type '%s' used as value (did you mean '%s{}'?)", name, name))
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}
		
		// Check if this is a namespace
		if _, isNamespace := v.ctx.NamespaceRegistry[name]; isNamespace {
			// Return a dummy value - the postfix operator will use the baseIdentifier parameter
			// to actually look up the function in the namespace
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}
		
		// Normal variable lookup
		sym, ok := v.ctx.currentScope.Lookup(name)
		if !ok {
			// Check if it's a function in the current namespace
			if v.ctx.currentNamespace != nil {
				if fn, ok := v.ctx.currentNamespace.Functions[name]; ok {
					return fn
				}
			}
			
			// Fallback: Check module directly
			if fn := v.ctx.Module.GetFunction(name); fn != nil {
				return fn
			}
			
			v.ctx.Diagnostics.Error(fmt.Sprintf("undefined: %s", name))
			return v.ctx.Builder.ConstInt(types.I64, 0)
		}

		if ptr, isAlloca := sym.Value.(*ir.AllocaInst); isAlloca {
			ptrType := ptr.Type().(*types.PointerType)
			loaded := v.ctx.Builder.CreateLoad(ptrType.ElementType, ptr, "")
			return loaded
		}

		return sym.Value
	}
	
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitStructLiteral(ctx *parser.StructLiteralContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	typ, ok := v.ctx.GetType(name)
	if !ok {
		v.ctx.Diagnostics.Error(fmt.Sprintf("unknown struct/class type: %s", name))
		return v.ctx.Builder.ConstInt(types.I64, 0)
	}
	
	structType, ok := typ.(*types.StructType)
	if !ok {
		v.ctx.Diagnostics.Error(fmt.Sprintf("%s is not a struct/class type", name))
		return v.ctx.Builder.ConstInt(types.I64, 0)
	}

	// Check if this is a class (requires heap allocation)
	if v.ctx.IsClassType(name) {
		// Allocate on heap for class instances
		ptrToClass := v.ctx.Builder.CreateAlloca(structType, name+".instance")
		
		// Zero-initialize all fields first
		for i := 0; i < len(structType.Fields); i++ {
			gep := v.ctx.Builder.CreateStructGEP(structType, ptrToClass, i, "")
			zero := v.getZeroValue(structType.Fields[i])
			v.ctx.Builder.CreateStore(zero, gep)
		}
		
		// Initialize specified fields
		for _, field := range ctx.AllFieldInit() {
			fieldName := field.IDENTIFIER().GetText()
			fieldVal := v.Visit(field.Expression()).(ir.Value)
			
			var idx int = -1
			if fieldIndices, ok := v.ctx.ClassFieldIndices[name]; ok {
				if fieldIdx, ok := fieldIndices[fieldName]; ok {
					idx = fieldIdx
				}
			}
			
			if idx < 0 {
				v.ctx.Diagnostics.Error(fmt.Sprintf("class %s has no field %s", name, fieldName))
				continue
			}
			
			gep := v.ctx.Builder.CreateStructGEP(structType, ptrToClass, idx, "")
			v.ctx.Builder.CreateStore(fieldVal, gep)
		}
		
		return ptrToClass
	}

	// Regular struct - build value directly
	var agg ir.Value = v.ctx.Builder.ConstZero(structType)

	// Populate specified fields
	for _, field := range ctx.AllFieldInit() {
		fieldName := field.IDENTIFIER().GetText()
		fieldVal := v.Visit(field.Expression()).(ir.Value)
		
		idx := v.findFieldIndex(structType, fieldName)
		if idx < 0 {
			v.ctx.Diagnostics.Error(fmt.Sprintf("struct %s has no field %s", name, fieldName))
			continue
		}
		
		agg = v.ctx.Builder.CreateInsertValue(agg, fieldVal, []int{idx}, "")
	}
	
	return agg
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
	
	if ctx.STRING_LITERAL() != nil {
		rawText := ctx.STRING_LITERAL().GetText()
		content, err := strconv.Unquote(rawText)
		if err != nil {
			if len(rawText) >= 2 {
				content = rawText[1 : len(rawText)-1]
			} else {
				content = rawText
			}
		}
		
		bytes := append([]byte(content), 0)
		elements := make([]ir.Constant, len(bytes))
		for i, b := range bytes {
			elements[i] = v.ctx.Builder.ConstInt(types.I8, int64(b))
		}
		
		arrType := types.NewArray(types.I8, int64(len(bytes)))
		constArr := &ir.ConstantArray{
			BaseValue: ir.BaseValue{ValType: arrType},
			Elements:  elements,
		}
		
		strName := fmt.Sprintf(".str.%d", len(v.ctx.Module.Globals))
		global := v.ctx.Builder.CreateGlobalConstant(strName, constArr)
		zero := v.ctx.Builder.ConstInt(types.I32, 0)
		
		return v.ctx.Builder.CreateInBoundsGEP(arrType, global, []ir.Value{zero, zero}, "")
	}
	
	return v.ctx.Builder.ConstInt(types.I64, 0)
}

func (v *IRVisitor) VisitCastExpression(ctx *parser.CastExpressionContext) interface{} {
	val := v.Visit(ctx.Expression()).(ir.Value)
	destType := v.resolveType(ctx.Type_())
	srcType := val.Type()
	
	if types.IsPointer(srcType) && types.IsInteger(destType) {
		return v.ctx.Builder.CreatePtrToInt(val, destType, "")
	}
	if types.IsInteger(srcType) && types.IsPointer(destType) {
		return v.ctx.Builder.CreateIntToPtr(val, destType, "")
	}
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
		if srcInt.Signed != destInt.Signed {
			return v.ctx.Builder.CreateBitCast(val, destType, "")
		}
		return val
	}
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
	if types.IsFloat(srcType) && types.IsFloat(destType) {
		srcFloat := srcType.(*types.FloatType)
		destFloat := destType.(*types.FloatType)
		if destFloat.BitWidth > srcFloat.BitWidth {
			return v.ctx.Builder.CreateFPExt(val, destType, "")
		}
		if destFloat.BitWidth < srcFloat.BitWidth {
			return v.ctx.Builder.CreateFPTrunc(val, destType, "")
		}
		return val
	}
	
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

func (v *IRVisitor) VisitSyscallExpression(ctx *parser.SyscallExpressionContext) interface{} {
	exprs := ctx.AllExpression()
	if len(exprs) == 0 {
		v.ctx.Diagnostics.Error("syscall requires at least a syscall number")
		return v.ctx.Builder.ConstInt(types.I64, -1)
	}

	args := make([]ir.Value, len(exprs))
	for i, expr := range exprs {
		val := v.Visit(expr).(ir.Value)
		
		// Auto-cast integers to I64 for the builder
		if types.IsInteger(val.Type()) {
			if val.Type().BitSize() < 64 {
				val = v.ctx.Builder.CreateSExt(val, types.I64, "")
			}
		}
		
		args[i] = val
	}

	return v.ctx.Builder.CreateSyscall(args)
}

func (v *IRVisitor) VisitArgumentList(ctx *parser.ArgumentListContext) interface{} {
	args := make([]ir.Value, 0)
	
	for _, expr := range ctx.AllExpression() {
		arg := v.Visit(expr)
		if arg == nil {
			continue
		}
		argVal, ok := arg.(ir.Value)
		if !ok {
			continue
		}
		args = append(args, argVal)
	}
	
	return args
}

func (v *IRVisitor) VisitLeftHandSide(ctx *parser.LeftHandSideContext) interface{} {
	if ctx.IDENTIFIER() != nil && ctx.DOT() == nil && ctx.STAR() == nil {
		name := ctx.IDENTIFIER().GetText()
		sym, ok := v.ctx.currentScope.Lookup(name)
		if ok {
			return sym.Value
		}
	}
	if ctx.STAR() != nil {
		return v.Visit(ctx.PostfixExpression())
	}
	if ctx.PostfixExpression() != nil {
		return v.Visit(ctx.PostfixExpression())
	}
	return v.ctx.Builder.ConstInt(types.I64, 0)
}