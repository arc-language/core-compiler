namespace main

// Linux x86_64 System Call Numbers
const SYS_WRITE = 1
const SYS_EXIT = 60

// File Descriptors
const STDOUT = 1

func main() int32 {
    // String literal decays to a pointer (*i8)
    let msg = "Hello, Direct Syscall!\n"
    let len = 23 // Length of the string above

    // 1. Write to specific file descriptor (stdout)
    // syscall(RAX, RDI, RSI, RDX) -> write(fd, buf, count)
    let ret = syscall(SYS_WRITE, STDOUT, msg, len)

    // 2. Exit the process nicely
    // syscall(RAX, RDI) -> exit(status)
    syscall(SYS_EXIT, 0)
    
    // Unreachable code (process will exit above)
    return 0
}