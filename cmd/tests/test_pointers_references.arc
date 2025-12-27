namespace main

func main() int32 {
    let value: int32 = 42
    
    // Basic pointer
    let ptr: *int32 = &value
    
    // Void pointer (opaque)
    let handle: *void = cast<*void>(ptr)
    
    // Basic reference
    let ref: &int32 = value
    
    // Address-of and dereference
    let x = *ptr
    *ptr = 100
    
    return 0
}