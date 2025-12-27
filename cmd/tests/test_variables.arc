namespace main

func main() int32 {
    // Mutable with type
    let x: int32 = 42
    x = 100
    
    // Mutable with inference
    let y = 42
    y = 100
    
    // Constants with type
    const c1: int32 = 42
    
    // Constants with inference
    const c2 = 42
    
    return 0
}