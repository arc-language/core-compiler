namespace io

// ---------------------------------------------------------
// Linux x86_64 Syscall Constants
// ---------------------------------------------------------
// rax = 1 (sys_write)
const SYS_WRITE: int64 = 1

// File Descriptors
const STDOUT: int32 = 1
const STDERR: int32 = 2

// ---------------------------------------------------------
// Core Functions
// ---------------------------------------------------------

// Write raw bytes to a file descriptor
// Returns number of bytes written
func write_raw(fd: int32, data: *byte, len: usize) usize {
    // Linux x86_64: syscall(rax=1, rdi=fd, rsi=buf, rdx=count)
    let res = syscall(SYS_WRITE, fd, data, len)
    return cast<usize>(res)
}

// Print a string to standard output
func print(msg: string) {
    // Convert high-level string to raw pointer for the kernel
    let ptr = cast<*byte>(msg)
    // Access internal length property of the string type
    let len = msg.len
    
    write_raw(STDOUT, ptr, len)
}

// Print a string to standard output with a trailing newline
func println(msg: string) {
    print(msg)
    print("\n")
}