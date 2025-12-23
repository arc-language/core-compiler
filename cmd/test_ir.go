// test_ir.go
package main

import (
    "fmt"
    "github.com/arc-language/core-builder/builder"
    "github.com/arc-language/core-builder/types"
)

func main() {
    // Create a simple module manually
    b := builder.New()
    mod := b.CreateModule("test")
    
    // Create a simple function: fn add(a: i32, b: i32) -> i32
    fn := b.CreateFunction("add", types.I32, []types.Type{types.I32, types.I32}, false)
    fn.Arguments[0].SetName("a")
    fn.Arguments[1].SetName("b")
    
    // Create entry block
    entry := b.CreateBlock("entry")
    b.SetInsertPoint(entry)
    
    // Add the arguments
    result := b.CreateAdd(fn.Arguments[0], fn.Arguments[1], "sum")
    b.CreateRet(result)
    
    // Print the IR
    fmt.Println(mod.String())
}