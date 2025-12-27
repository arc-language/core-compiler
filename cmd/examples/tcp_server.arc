namespace main

extern c {
    // Socket API
    func socket(int32, int32, int32) int32
    func setsockopt(int32, int32, int32, *byte, int32) int32
    func bind(int32, *byte, int32) int32
    func listen(int32, int32) int32
    func accept(int32, *byte, *int32) int32
    func recv(int32, *byte, int64, int32) int64
    func send(int32, *byte, int64, int32) int64
    func close(int32) int32
    
    // Memory
    func memset(*byte, int32, int64) *byte
    
    // I/O
    func puts(*byte) int32
    func printf(*byte, ...) int32
    func perror(*byte) void
}

func htons(n: uint16) uint16 {
    let high = n / 256
    let low = n * 256
    return low + high
}

func main() int32 {
    // Constants
    let AF_INET = 2
    let SOCK_STREAM = 1
    let SOL_SOCKET = 1
    let SO_REUSEADDR = 2
    let PORT = 8080
    
    c.puts("Server: Creating socket...")
    let server_fd = c.socket(AF_INET, SOCK_STREAM, 0)
    if server_fd < 0 {
        c.perror("Socket failed")
        return 1
    }

    // Enable SO_REUSEADDR (allows restarting server quickly)
    let opt_val = 1
    let opt_ptr = alloca(int32, 1)
    *opt_ptr = opt_val
    
    // Cast *int32 to *byte for the API
    let opt_void = cast<*byte>(opt_ptr)
    c.setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, opt_void, 4)

    // --- Prepare Address Struct ---
    let addr = alloca(uint8, 16)
    c.memset(addr, 0, 16)
    
    // Family = AF_INET (offset 0)
    let ptr_family = cast<*int16>(addr)
    *ptr_family = cast<int16>(AF_INET)
    
    // Port = 8080 (offset 2)
    let port_u16 = cast<uint16>(PORT)
    let net_port = htons(port_u16)
    
    let addr_int = cast<int64>(addr)
    let port_ptr = cast<*uint16>(addr_int + 2)
    *port_ptr = net_port
    
    // IP = INADDR_ANY (0) is already set by memset

    // --- Bind & Listen ---
    if c.bind(server_fd, addr, 16) < 0 {
        c.perror("Bind failed")
        return 1
    }

    if c.listen(server_fd, 5) < 0 {
        c.perror("Listen failed")
        return 1
    }

    c.printf("Server: Listening on port %d...\n", PORT)

    // --- Accept Loop ---
    let client_len = alloca(int32, 1)
    let client_addr = alloca(uint8, 16)
    
    for {
        *client_len = 16
        c.memset(client_addr, 0, 16)
        
        let client_fd = c.accept(server_fd, client_addr, client_len)
        if client_fd < 0 {
            c.perror("Accept failed")
            continue
        }
        
        c.puts("Server: Client connected!")
        
        // Handle client (echo)
        let buf = alloca(uint8, 1024)
        
        // Loop to read all data from this client
        for {
            c.memset(buf, 0, 1024)
            let n = c.recv(client_fd, buf, 1024, 0)
            
            if n <= 0 {
                // Client disconnected or error
                break
            }
            
            c.printf("  Received: %s", buf)
            c.send(client_fd, buf, cast<int32>(n), 0)
        }
        
        c.puts("Server: Client disconnected.")
        c.close(client_fd)
    }
    
    c.close(server_fd)
    return 0
}