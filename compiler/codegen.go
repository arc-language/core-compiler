// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"fmt"
	"os"

	"github.com/arc-language/core-codegen/codegen"
)

// CompileToIR generates textual IR from the module
func (c *Compiler) CompileToIR(outputPath string) error {
	c.logger.Info("Generating textual IR to: %s", outputPath)
	
	if c.context.Module == nil {
		c.logger.Error("No module to compile")
		return fmt.Errorf("no module to compile")
	}

	// Generate IR text
	irText := c.context.Module.String()
	c.logger.Debug("Generated %d bytes of IR text", len(irText))

	// Write to file
	if err := os.WriteFile(outputPath, []byte(irText), 0644); err != nil {
		c.logger.Error("Failed to write IR file '%s': %v", outputPath, err)
		return fmt.Errorf("failed to write IR file: %v", err)
	}

	c.logger.Info("Successfully wrote IR to: %s", outputPath)
	return nil
}

// CompileToObject generates an object file from the module
func (c *Compiler) CompileToObject(outputPath string) error {
	c.logger.Info("Generating object file to: %s", outputPath)
	
	if c.context.Module == nil {
		c.logger.Error("No module to compile")
		return fmt.Errorf("no module to compile")
	}

	// Generate object code
	c.logger.Debug("Calling code generator for module '%s'", c.context.Module.Name)
	objData, err := codegen.GenerateObject(c.context.Module)
	if err != nil {
		c.logger.Error("Code generation failed: %v", err)
		return fmt.Errorf("code generation failed: %v", err)
	}

	c.logger.Debug("Generated %d bytes of object code", len(objData))

	// Write object file
	if err := os.WriteFile(outputPath, objData, 0644); err != nil {
		c.logger.Error("Failed to write object file '%s': %v", outputPath, err)
		return fmt.Errorf("failed to write object file: %v", err)
	}

	c.logger.Info("Successfully wrote object file to: %s", outputPath)
	return nil
}