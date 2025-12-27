namespace main

func main() int32 {
    // Fixed-width integers (signed)
    let i8: int8 = -128
    let i16: int16 = -32768
    let i32: int32 = -500
    let i64: int64 = -9223372036854775808
    
    // Fixed-width integers (unsigned)
    let u8: uint8 = 255
    let u16: uint16 = 65535
    let u32: uint32 = 4294967295
    let u64: uint64 = 10000
    
    // Architecture dependent
    let len: usize = 100
    let offset: isize = -4
    
    // Floating point
    let f32: float32 = 3.14
    let f64: float64 = 2.71828
    
    // Aliases
    let b: byte = 255
    let flag: bool = true
    let r: char = 'a'
    
    // String (composite)
    let s: string = "hello"
    
    return 0
}