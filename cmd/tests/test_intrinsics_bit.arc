namespace main

func main() int32 {
    // bit_cast - reinterpret bits
    let f: float32 = 1.0
    let bits = bit_cast<uint32>(f)
    
    let i: int32 = -1
    let u = bit_cast<uint32>(i)
    
    return 0
}