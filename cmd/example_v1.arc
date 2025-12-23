// example_v1.arc
namespace main

func add(a: int64, b: int64) int64 {
    return a + b
}

func main() int64 {
    let x: int64 = 100
    let y: int64 = 42
    return add(x, y)
}