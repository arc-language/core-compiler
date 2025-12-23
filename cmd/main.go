package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arc-language/core-compiler/compiler"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	inputFile := os.Args[1]
	
	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist\n", inputFile)
		os.Exit(1)
	}
	
	// Determine output file
	outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + ".ir"
	if len(os.Args) >= 4 && os.Args[2] == "-o" {
		outputFile = os.Args[3]
	}
	
	// Extract module name from file
	moduleName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	
	fmt.Printf("Compiling %s...\n", inputFile)
	
	// Create compiler
	comp := compiler.NewCompiler(moduleName)
	
	// Compile file
	module, err := comp.CompileFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		os.Exit(1)
	}
	
	// Write IR to file
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	
	_, err = f.WriteString(module.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write IR: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✓ Compilation successful\n")
	fmt.Printf("✓ IR written to %s\n", outputFile)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: arc <source-file> [-o <output-file>]\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  arc program.arc              # Generates program.ir\n")
	fmt.Fprintf(os.Stderr, "  arc program.arc -o out.ir    # Generates out.ir\n")
}