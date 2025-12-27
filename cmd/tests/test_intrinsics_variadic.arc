namespace main

func print_values(first: int32, ...) {
    let args = va_start(first)
    defer va_end(args)
    
    let val1 = va_arg<int32>(args)
    let val2 = va_arg<float64>(args)
}

func main() int32 {
    print_values(1, 42, 3.14)
    return 0
}