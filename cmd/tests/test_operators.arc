namespace main

func main() int32 {
    let a: int32 = 10
    let b: int32 = 3
    
    // Arithmetic
    let sum = a + b
    let diff = a - b
    let prod = a * b
    let quot = a / b
    let rem = a % b
    
    // Compound assignment
    let x = 10
    x += 5
    x -= 3
    x *= 2
    x /= 4
    x %= 3
    
    // Increment/decrement
    let i = 0
    i++
    ++i
    i--
    --i
    
    // Comparison
    let eq = a == b
    let ne = a != b
    let lt = a < b
    let le = a <= b
    let gt = a > b
    let ge = a >= b
    
    // Logical
    let flag1 = true
    let flag2 = false
    let and_result = flag1 && flag2
    let or_result = flag1 || flag2
    
    // Unary
    let neg = -a
    let not = !flag1
    
    // Pointer arithmetic
    let ptr: *int32 = &a
    let next = ptr + 1
    let prev = ptr - 2
    
    return 0
}