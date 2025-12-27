namespace main

func main() int32 {
    let value: int32 = 42
    let ptr: *int32 = &value
    
    // Load (dereference to read)
    let loaded = *ptr
    
    // Store (dereference to write)
    *ptr = 100
    
    // Indexed pointer access
    let buffer = alloca(int32, 10)
    buffer[0] = 1
    buffer[5] = 50
    let val = buffer[5]
    
    return 0
}