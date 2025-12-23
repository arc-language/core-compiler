// example_v2.arc
namespace main


extern io {
    func printf (*byte, ...) int32
}

func main() int32 {

    let y = 42

    io.printf("%d\n", y)

    return 0
}