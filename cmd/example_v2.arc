// example_v1.arc
namespace main



extern io {
    func printf (*byte, ...) -> int32
}


func add(a: int32, b: int32) int32 {

    return a + b
}

func someFunc() int32 {
 
    let m: int32 = 100

    if m == 50 {
        return 200
    } else if m == 100 {
        return 100
    }

    return m
}

func main() int32 {

    let x = someFunc()
    let y = 42


    return 0
}