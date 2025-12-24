// example_tcp.arc
namespace main

extern c {
    // Socket API
    func socket(int32, int32, int32) int32
    func connect(int32, *byte, int32) int32
    func send(int32, *byte, int64, int32) int64
    func recv(int32, *byte, int64, int32) int64
    func close(int32) int32
    
    // Network helpers
    func inet_addr(*byte) uint32
    
    // Memory helpers
    func memset(*byte, int32, int64) *byte
    
    // I/O
    func puts(*byte) int32
    func printf(*byte, ...) int32
    func perror(*byte) void
}

func htons(n: uint16) uint16 {
    // Manually swap bytes since we might not have bitwise operators yet
    // high part -> low part (div 256)
    // low part -> high part (mul 256)
    let high = n / 256
    let low = n * 256
    return low + high
}

func main() int32 {
    let AF_INET = 2
    let SOCK_STREAM = 1
    
    c.puts("Client: Creating socket...")
    let fd = c.socket(AF_INET, SOCK_STREAM, 0)
    
    if fd < 0 {
        c.perror("Socket creation failed")
        return 1
    }

    // Allocate memory for sockaddr_in struct (16 bytes)
    let addr = alloca(uint8, 16)
    c.memset(addr, 0, 16)
    
    // --- Set family (AF_INET = 2) at offset 0 ---
    let ptr_family = cast<*int16>(addr)
    *ptr_family = cast<int16>(AF_INET)
    
    // --- Set port (8080) at offset 2 ---
    let port = cast<uint16>(8080)
    let net_port = htons(port)
    
    // Pointer arithmetic: addr + 2 bytes
    let addr_int = cast<int64>(addr)
    let port_ptr_int = addr_int + 2
    let port_ptr = cast<*uint16>(port_ptr_int)
    *port_ptr = net_port
    
    // --- Set IP (127.0.0.1) at offset 4 ---
    let ip_str = "127.0.0.1"
    let ip_val = c.inet_addr(ip_str)
    
    // Pointer arithmetic: addr + 4 bytes
    let ip_ptr_int = addr_int + 4
    let ip_ptr = cast<*uint32>(ip_ptr_int)
    *ip_ptr = ip_val
    
    // --- Connect ---
    c.printf("Client: Connecting to %s:8080...\n", ip_str)
    let res = c.connect(fd, addr, 16)
    if res < 0 {
        c.perror("Connect failed (make sure a server is running on port 8080)")
        c.close(fd)
        return 1
    }
    
    c.puts("Client: Connected!")
    
    // --- Send ---
    let msg = "hello"
    c.send(fd, msg, 5, 0)
    c.puts("Client: Sent 'hello'")
    
    // --- Recv ---
    let buf = alloca(uint8, 64)
    c.memset(buf, 0, 64)
    
    let n = c.recv(fd, buf, 63, 0)
    if n > 0 {
        c.printf("Client: Server replied with '%s'\n", buf)
    }
    
    c.close(fd)
    return 0
}