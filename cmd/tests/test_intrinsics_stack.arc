namespace main

func main() int32 {
    // Allocate single item on stack
    let ptr = alloca(int32)
    *ptr = 42
    
    // Allocate buffer on stack
    let buffer = alloca(byte, 1024)
    
    // Use with memset
    memset(buffer, 0, 1024)
    
    return 0
}