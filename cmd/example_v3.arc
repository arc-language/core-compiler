// example_v4.arc
namespace main

extern io {
    func printf(*byte, ...) int32
    func fopen(*byte, *byte) *byte
    func fclose(*byte) int32
    
    // fputs is NOT variadic, so it is safer for now
    func fputs(*byte, *byte) int32
}

func main() int32 {
    let filename = "test.txt"
    let file = io.fopen(filename, "w")

    // Check for NULL pointer (requires the cast fix above)
    if cast<int64>(file) == 0 {
        io.printf("Error: Could not open file!\n")
        return 1
    }

    // Use fputs (arguments: string, file_pointer)
    io.fputs("Hello from Arc (using fputs)!\n", file)
    
    io.fclose(file)
    io.printf("Success writing to %s\n", filename)
    return 0
}