namespace main

extern c {
    func printf(*byte, ...) int32
}


func main() int32 {

    for x in 1..10 {

        c.printf("%d\n", x)
    }
    
    return 0
}