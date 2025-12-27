namespace main

extern libc {
    // Maps Arc 'printf' to C symbol 'printf'
    func printf "printf" (*byte, ...) int32
    
    // Maps Arc 'sleep' to C symbol 'usleep'
    func sleep "usleep" (int32) int32
    
    // Direct mapping
    func usleep(int32) int32
    
    func malloc(usize) *void
    func free(*void)
}

func main() int32 {
    let ptr = malloc(1024)
    defer free(ptr)
    
    sleep(1000000)
    
    return 0
}