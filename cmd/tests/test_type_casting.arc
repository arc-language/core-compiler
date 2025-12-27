namespace main

func main() int32 {
    let i32_val: int32 = 42
    
    // Basic casting
    let i64_val = cast<int64>(i32_val)
    let f64_val = cast<float64>(i32_val)
    
    // Pointer conversions
    let ptr: *int32 = &i32_val
    let byte_ptr = cast<*byte>(ptr)
    
    // Pointer to integer
    let addr = cast<uint64>(ptr)
    
    // Integer to pointer
    let new_ptr = cast<*int32>(addr)
    
    // Void pointer
    let generic = cast<*void>(ptr)
    
    return 0
}