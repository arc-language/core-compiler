namespace main

import "io"

func main() int32 {
    io.printf("Hello, %s!\n", "World")
    io.printf("Number: %d\n", 42)
    io.printf("Unsigned: %u\n", 100)
    io.printf("Hex: %x\n", 255)
    io.printf("HEX: %X\n", 255)
    io.printf("Char: %c\n", 'A')
    io.printf("Percent: %%\n")

    return 0
}