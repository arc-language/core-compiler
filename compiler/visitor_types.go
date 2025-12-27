package compiler

import (
	"github.com/arc-language/core-builder/types"
	"github.com/arc-language/core-parser"
)

// registerStructType registers a struct type in pass 1
func (v *IRVisitor) registerStructType(ctx *parser.StructDeclContext) {
	name := ctx.IDENTIFIER().GetText()
	
	// Check if already registered
	if _, ok := v.ctx.GetType(name); ok {
		v.logger.Debug("Struct type '%s' already registered", name)
		return
	}
	
	v.logger.Debug("Registering struct type: %s", name)
	
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
			v.logger.Debug("  Field '%s' at index %d, type: %v", fieldName, fieldIndex, fieldType)
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
	
	v.logger.Info("Registering class type: %s", name)
	
	// Check if already registered
	if _, ok := v.ctx.GetType(name); ok {
		v.logger.Debug("Class '%s' already registered", name)
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
			
			v.logger.Debug("  Field '%s' at index %d, type: %v", fieldName, fieldIndex, fieldType)
			
			fieldTypes = append(fieldTypes, fieldType)
			fieldMap[fieldName] = fieldIndex
			fieldIndex++
		}
	}
	
	// Register mapping in context
	v.ctx.ClassFieldIndices[name] = fieldMap

	// Create struct type for the class
	structType := types.NewStruct(name, fieldTypes, false)
	
	v.ctx.RegisterClass(name, structType)
	v.logger.Debug("Registered class '%s' with %d fields", name, len(fieldTypes))
}

func (v *IRVisitor) VisitStructDecl(ctx *parser.StructDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	v.logger.Debug("Processing struct declaration: %s", name)
	
	// Type already registered in pass 1
	// Now compile methods
	for _, member := range ctx.AllStructMember() {
		if member.FunctionDecl() != nil {
			v.Visit(member.FunctionDecl())
		}
	}
	
	return nil
}

func (v *IRVisitor) VisitClassDecl(ctx *parser.ClassDeclContext) interface{} {
	name := ctx.IDENTIFIER().GetText()
	v.logger.Info("Processing class declaration: %s", name)
	
	// Type already registered in pass 1
	// Now compile methods
	for i, member := range ctx.AllClassMember() {
		v.logger.Debug("Processing class member %d/%d", i+1, len(ctx.AllClassMember()))
		if member.FunctionDecl() != nil {
			v.Visit(member.FunctionDecl())
		} else if member.DeinitDecl() != nil {
			v.Visit(member.DeinitDecl())
		} else if member.ClassField() != nil {
			// Fields are handled in registerClassType
			v.logger.Debug("Skipping field (already registered)")
		}
	}
	
	v.logger.Info("Completed class declaration: %s", name)
	return nil
}

func (v *IRVisitor) VisitClassField(ctx *parser.ClassFieldContext) interface{} {
	// Field definitions are handled in registerClassType
	v.logger.Debug("VisitClassField called for: %s (should not process here)", ctx.IDENTIFIER().GetText())
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
	v.ctx.Logger.Warning("deinit is not yet implemented")
	return nil
}