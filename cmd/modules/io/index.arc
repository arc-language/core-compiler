// std/io.arc
namespace io

// ============================================================================
// Syscall Numbers (Linux x86_64)
// ============================================================================

const SYS_WRITE: int32 = 1

// ============================================================================
// File Descriptors
// ============================================================================

const STDOUT: int32 = 1

// ============================================================================
// Core Write Function
// ============================================================================

// Write bytes to file descriptor
func write(fd: int32, data: *byte, len: usize) isize {
    return cast<isize>(syscall(SYS_WRITE, fd, cast<uint64>(data), cast<uint64>(len)))
}

// ============================================================================
// Printf Helpers
// ============================================================================

// Write single character to fd
func write_char(fd: int32, ch: char) {
    let buf = alloca(byte, 1)
    buf[0] = cast<byte>(ch)
    write(fd, buf, 1)
}

// Helper to get digit character for base conversion
func get_digit_char(digit: uint64) byte {
    if digit < 10 {
        return cast<byte>('0' + cast<char>(digit))
    }
    return cast<byte>('a' + cast<char>(digit - 10))
}

// Convert unsigned integer to string
func uint_to_str(value: uint64, buffer: *byte, base: uint32) usize {
    if value == 0 {
        buffer[0] = '0'
        return 1
    }
    
    let temp = alloca(byte, 64)
    let pos: usize = 0
    
    let v = value
    for v > 0 {
        let digit = v % cast<uint64>(base)
        temp[pos] = get_digit_char(digit)
        v = v / cast<uint64>(base)
        pos++
    }
    
    // Reverse into output buffer
    for let i: usize = 0; i < pos; i++ {
        buffer[i] = temp[pos - i - 1]
    }
    
    return pos
}

// Convert signed integer to string
func int_to_str(value: int64, buffer: *byte) usize {
    let pos: usize = 0
    
    if value < 0 {
        buffer[0] = '-'
        pos = 1
        value = -value
    }
    
    let uval = cast<uint64>(value)
    let len = uint_to_str(uval, buffer + pos, 10)
    return pos + len
}

// Write string to file descriptor
func write_string(fd: int32, s: string) {
    let ptr = cast<*byte>(s)
    let len = *cast<*usize>(cast<uint64>(&s) + 8)
    write(fd, ptr, len)
}

// ============================================================================
// Printf Implementation
// ============================================================================

// Core formatting function
func vfprintf(fd: int32, fmt: string, args: *void) {
    let fmt_ptr = cast<*byte>(fmt)
    let fmt_len = *cast<*usize>(cast<uint64>(&fmt) + 8)
    
    let buffer = alloca(byte, 64)
    let i: usize = 0
    
    for i < fmt_len {
        let ch = fmt_ptr[i]
        
        if ch == '%' {
            i++
            if i >= fmt_len {
                break
            }
            
            let spec = fmt_ptr[i]
            
            if spec == 'd' {
                let val = va_arg<int32>(args)
                let len = int_to_str(cast<int64>(val), buffer)
                write(fd, buffer, len)
                
            } else if spec == 'u' {
                let val = va_arg<uint32>(args)
                let len = uint_to_str(cast<uint64>(val), buffer, 10)
                write(fd, buffer, len)
                
            } else if spec == 'x' {
                let val = va_arg<uint32>(args)
                let len = uint_to_str(cast<uint64>(val), buffer, 16)
                write(fd, buffer, len)
                
            } else if spec == 'X' {
                let val = va_arg<uint32>(args)
                let len = uint_to_str(cast<uint64>(val), buffer, 16)
                for let j: usize = 0; j < len; j++ {
                    if buffer[j] >= 'a' && buffer[j] <= 'f' {
                        buffer[j] = buffer[j] - 32
                    }
                }
                write(fd, buffer, len)
                
            } else if spec == 's' {
                let str = va_arg<string>(args)
                write_string(fd, str)
                
            } else if spec == 'c' {
                let c = va_arg<char>(args)
                write_char(fd, c)
                
            } else if spec == '%' {
                write_char(fd, '%')
                
            } else {
                write_char(fd, '%')
                write_char(fd, cast<char>(spec))
            }
            
        } else {
            write_char(fd, cast<char>(ch))
        }
        
        i++
    }
}

// Printf - formatted output to stdout
func printf(fmt: string, ...) {
    let args = va_start(fmt)
    defer va_end(args)
    
    vfprintf(STDOUT, fmt, args)
}