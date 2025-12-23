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

	command := os.Args[1]

	switch command {
	case "build":
		handleBuild(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleBuild(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: No input file specified\n\n")
		printUsage()
		os.Exit(1)
	}

	inputFile := args[0]
	outputFile := ""

	// Parse -o flag
	for i := 1; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outputFile = args[i+1]
			break
		}
	}

	if outputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: Output file not specified (use -o)\n\n")
		printUsage()
		os.Exit(1)
	}

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File '%s' does not exist\n", inputFile)
		os.Exit(1)
	}

	// Determine output format from extension
	ext := strings.ToLower(filepath.Ext(outputFile))
	if ext != ".o" && ext != ".ir" {
		fmt.Fprintf(os.Stderr, "Error: Output file must have .o or .ir extension\n")
		os.Exit(1)
	}

	// Extract module name from input file
	moduleName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	fmt.Printf("Compiling %s...\n", inputFile)

	// Create compiler
	comp := compiler.NewCompiler(moduleName)

	// Compile source file
	module, err := comp.CompileFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Module has %d functions, %d globals\n", len(module.Functions), len(module.Globals))

	// Generate output based on extension
	if ext == ".o" {
		err = comp.CompileToObject(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Object generation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Object file written to %s\n", outputFile)

		// Print linking hint
		exeName := strings.TrimSuffix(filepath.Base(outputFile), ".o")
		fmt.Printf("\nTo create executable:\n")
		fmt.Printf("  gcc %s -o %s && ./%s\n", outputFile, exeName, exeName)
	} else {
		err = comp.CompileToIR(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "IR generation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ IR written to %s\n", outputFile)
	}
}

func printUsage() {
	fmt.Println("Arc Language Compiler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  arc build <source-file> -o <output-file>")
	fmt.Println("  arc help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build    Compile an Arc source file")
	fmt.Println("  help     Show this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -o <file>    Output file (.o for object, .ir for IR)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  arc build program.arc -o output.o     # Compile to object file")
	fmt.Println("  arc build program.arc -o output.ir    # Compile to IR")
	fmt.Println("  arc help                              # Show help")
}