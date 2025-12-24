// example_v5.arc
namespace main

extern io {
    // printf is variadic, denoted by ...
    func printf(*byte, ...) int32
    func puts(*byte) int32
}

func main() int32 {
    io.puts("=== Testing Arc Loops ===")

    // 1. Standard C-style For Loop
    // for init; condition; post
    io.puts("\n1. Standard Loop (0 to 4):")
    for let i = 0; i < 5; i = i + 1 {
        io.printf("  i = %d\n", i)
    }

    // 2. While-style Loop
    // for condition
    io.puts("\n2. While-style Loop (Countdown 3 to 1):")
    let j = 3
    for j > 0 {
        io.printf("  j = %d\n", j)
        j = j - 1
    }

    // 3. Loop with 'continue'
    // Skips the iteration when k == 2
    io.puts("\n3. Loop with continue (Skipping 2):")
    for let k = 0; k < 4; k = k + 1 {
        if k == 2 {
            io.puts("  (skipping 2)")
            continue
        }
        io.printf("  k = %d\n", k)
    }

    // 4. Infinite Loop with 'break'
    io.puts("\n4. Infinite Loop with break:")
    let counter = 0
    for {
        io.printf("  counter = %d\n", counter)
        counter = counter + 1
        
        if counter >= 3 {
            io.puts("  Breaking out!")
            break
        }
    }

    io.puts("\nDone.")
    return 0
}