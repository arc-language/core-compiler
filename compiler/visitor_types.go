package compiler

import (
	"fmt"
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// registerStructType registers a struct type in pass 1
func (v *IRVisitor) registerStructType(ctx *parser.StructDeclContext) {
	name := ctx.IDENTIFIER().GetText()
	
	// Check if already registered
	if _, ok := v.ctx.GetType(name); ok {
		return
	}
	
	// Create field map
	fieldMap := make(map[string]int)
	fieldTypes := make([]types.Type, 0)
	
	fieldIndex := 0
	for _, member := range ctx.AllStructMember() {
		if member.StructField() != nil {
			field := member.StructField()
			fieldName := field.IDENTIFIER().GetText()
			fieldType := v.resolveType(field.Type_())
			
			fieldTypes = append(fieldTypes, fieldType)
			fieldMap[fieldName] = fieldIndex
			fieldIndex++
		}
	}
	
	// Register mapping in context
	v.ctx.StructFieldIndices[name] = fieldMap

	structType := types.NewStruct(name, fieldTypes, false)
	v.ctx.RegisterType(name, structType)
}

// registerClassType registers a class type in pass 1
func (v *IRVisitor) registerClassType(ctx *parser.ClassDeclContext) {
	name := ctx.IDENTIFIER().GetText()
	
	fmt.Printf("DEBUG: Registering class type: %s\n", name)
	
	// Check if already registered
	if _, ok := v.ctx.GetType(name); ok {
		fmt.Printf("DEBUG: Class '%s' already registered\n", name)
		return
	}
	
	// Create field map
	fieldMap := make(map[string]int)
	fieldTypes := make([]types.Type, 0)
	
	fieldIndex := 0
	for _, member := range ctx.AllClassMember() {
		if member.ClassField() != nil {
			field := member.ClassField()
			fieldName := field.IDENTIFIER().GetText()
			fieldType := v.resolveType(field.Type_())
			
			fmt.Printf("DEBUG: Class '%s' field '%s' at index %d, type: %v\n", name, fieldName, fieldIndex, fieldType)
			
			fieldTypes = append(fieldTypes, fieldType)
			fieldMap[fieldName] = fieldIndex
			fieldIndex++
		}
	}
	
	// Register mapping in context
	v.ctx.ClassFieldIndices[name] = fieldMap
	fmt.Printf("DEBUG: Registered ClassFieldIndices[%s] = %v\n", name, fieldMap)

	// Create struct type for the class - ENSURE NAME IS SET
	structType := types.NewStruct(name, fieldTypes, false)
	fmt.Printf("DEBUG: Created struct type with name: %s\n", structType.Name)
	
	v.ctx.RegisterClass(name, structType)
	fmt.Printf("DEBUG: Registered class '%s' in context\n", name)
}

func (v *IRVisitor) VisitStructDecl(ctx *parser.StructDeclContext) interface{} {
	// Type already registered in pass 1
	// Now compile methods
	for _, member := range ctx.AllStructMember() {
		if member.FunctionDecl() != nil {
			v.Visit(member.FunctionDecl())
		}
	}
	
	return nil
}

// visitor_types.go - Enhanced VisitClassDecl

func (v *IRVisitor) VisitClassDecl(ctx *parser.ClassDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	fmt.Printf("DEBUG VisitClassDecl: Processing class %s\n", name)
	
	// Type already registered in pass 1
	// Now compile methods
	for i, member := range ctx.AllClassMember() {
		fmt.Printf("DEBUG VisitClassDecl: Processing member %d of %d\n", i, len(ctx.AllClassMember()))
		if member.FunctionDecl() != nil {
			fmt.Printf("DEBUG VisitClassDecl: Member %d is a function\n", i)
			v.Visit(member.FunctionDecl())
		} else if member.DeinitDecl() != nil {
			fmt.Printf("DEBUG VisitClassDecl: Member %d is deinit\n", i)
			v.Visit(member.DeinitDecl())
		} else if member.ClassField() != nil {
			fmt.Printf("DEBUG VisitClassDecl: Member %d is a field (skipping)\n", i)
		}
	}
	
	fmt.Printf("DEBUG VisitClassDecl: Completed class %s\n", name)
	return nil
}

func (v *IRVisitor) VisitClassField(ctx *parser.ClassFieldContext) interface{} {
	// Field definitions are handled in registerClassType
	fmt.Printf("DEBUG VisitClassField: Field %s (should not process here)\n", ctx.IDENTIFIER().GetText())
	return nil
}

func (v *IRVisitor) VisitClassMember(ctx *parser.ClassMemberContext) interface{} {
	if ctx.ClassField() != nil {
		return v.Visit(ctx.ClassField())
	}
	if ctx.FunctionDecl() != nil {
		return v.Visit(ctx.FunctionDecl())
	}
	if ctx.DeinitDecl() != nil {
		return v.Visit(ctx.DeinitDecl())
	}
	return nil
}

func (v *IRVisitor) VisitDeinitDecl(ctx *parser.DeinitDeclContext) interface{} {
	// TODO: Implement deinit as a special destructor function
	// This will be called when reference count reaches zero
	v.ctx.Diagnostics.Warning("deinit is not yet implemented")
	return nil
}