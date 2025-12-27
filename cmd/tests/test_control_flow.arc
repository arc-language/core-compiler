namespace main

func main() int32 {
    let value: int32 = 5
    
    // If-else
    if value > 10 {
        // branch 1
    } else if value > 5 {
        // branch 2
    } else {
        // branch 3
    }
    
    // Defer
    let ptr = alloca(byte, 64)
    defer memset(ptr, 0, 64)
    
    // Early return
    if value < 0 {
        return -1
    }
    
    return 0
}