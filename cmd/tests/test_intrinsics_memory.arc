namespace main

func main() int32 {
    let buf1 = alloca(byte, 1024)
    let buf2 = alloca(byte, 1024)
    
    // memset - zero buffer
    memset(buf1, 0, 1024)
    
    // memcpy - copy non-overlapping
    memcpy(buf2, buf1, 1024)
    
    // memmove - copy with potential overlap
    memmove(buf1, buf1 + 10, 500)
    
    // memcmp - compare buffers
    let diff = memcmp(buf1, buf2, 1024)
    
    return 0
}