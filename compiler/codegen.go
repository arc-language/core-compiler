// Package compiler provides the Arc language compiler implementation.
package compiler

import (
	"fmt"
	"os"

	"github.com/arc-language/core-codegen/codegen"
)

// CompileToIR generates textual IR from the module
func (c *Compiler) CompileToIR(outputPath string) error {
	if c.context.Module == nil {
		return fmt.Errorf("no module to compile")
	}

	// Generate IR text
	irText := c.context.Module.String()

	// Write to file
	if err := os.WriteFile(outputPath, []byte(irText), 0644); err != nil {
		return fmt.Errorf("failed to write IR file: %v", err)
	}

	return nil
}

// CompileToObject generates an object file from the module
func (c *Compiler) CompileToObject(outputPath string) error {
	if c.context.Module == nil {
		return fmt.Errorf("no module to compile")
	}

	// Generate object code
	objData, err := codegen.GenerateObject(c.context.Module)
	if err != nil {
		return fmt.Errorf("code generation failed: %v", err)
	}

	// Write object file
	if err := os.WriteFile(outputPath, objData, 0644); err != nil {
		return fmt.Errorf("failed to write object file: %v", err)
	}

	return nil
}