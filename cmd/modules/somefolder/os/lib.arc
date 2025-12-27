namespace utils

// We can define internal constants here too
const SYS_WRITE = 1
const STDOUT = 1

// This function is exported to the 'utils' namespace
func Something() void {
    let msg = "[utils] Something() was called!\n"
    // Length: 32
    syscall(SYS_WRITE, STDOUT, msg, 32)
}