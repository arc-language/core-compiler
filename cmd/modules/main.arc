namespace main

import "somefolder/os"

// Linux x86_64 System Call Numbers
const SYS_WRITE = 1
const SYS_EXIT = 60

// File Descriptors
const STDOUT = 1

func main() int32 {
    // 1. Call the imported function from the 'utils' namespace
    // This demonstrates the package loading worked
    utils.Something()

    // 2. Standard main logic
    // String literal decays to a pointer (*i8)
    let msg = "Hello, Direct Syscall!\n"
    let len = 23 

    // syscall(RAX, RDI, RSI, RDX) -> write(fd, buf, count)
    let ret = syscall(SYS_WRITE, STDOUT, msg, len)

    // syscall(RAX, RDI) -> exit(status)
    syscall(SYS_EXIT, 0)
    
    return 0
}