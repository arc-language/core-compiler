namespace main

struct TestStruct {
    a: int32
    b: int64
    c: float64
}

func main() int32 {
    // Sizeof
    let sz1 = sizeof<int32>
    let sz2 = sizeof<float64>
    let sz3 = sizeof<TestStruct>
    
    // Alignof
    let align1 = alignof<int32>
    let align2 = alignof<float64>
    let align3 = alignof<TestStruct>
    
    return 0
}