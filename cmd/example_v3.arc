// example_v3.arc
namespace main

extern io {
    // puts and fputs are NOT variadic. They are safe to use.
    func puts(*byte) int32
    func fputs(*byte, *byte) int32
    
    func fopen(*byte, *byte) *byte
    func fclose(*byte) int32
}

func main() int32 {
    io.puts("Start: Opening file...")

    let filename = "test_output.txt"
    let mode = "w"
    
    let file = io.fopen(filename, mode)

    if cast<int64>(file) == 0 {
        io.puts("Error: fopen returned NULL!")
        return 1
    }

    io.puts("File opened successfully.")
    
    // Write to file
    io.fputs("Hello from Arc Language!\n", file)
    io.fputs("This is a text file test.\n", file)
    
    io.fclose(file)
    io.puts("Done. Check test_output.txt")

    return 0
}