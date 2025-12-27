namespace main

func main() int32 {
    // strlen - C-string length
    let cstr: *byte = cast<*byte>("hello")
    let len = strlen(cstr)
    
    // memchr - find byte in memory
    let buf: *byte = cast<*byte>("hello\nworld")
    let newline = memchr(buf, cast<byte>('\n'), 11)
    
    return 0
}