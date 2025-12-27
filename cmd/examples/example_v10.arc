namespace main

extern c {
    func socket(int32, int32, int32) int32
    func connect(int32, *byte, int32) int32
    func send(int32, *byte, int64, int32) int64
    func recv(int32, *byte, int64, int32) int64
    func close(int32) int32
    func printf(*byte, ...) int32
    func memset(*byte, int32, int64) *byte
    func inet_addr(*byte) uint32
}

func htons(n: uint16) uint16 {
    let high = n / 256
    let low = n * 256
    return low + high
}

class TcpClient {
    fd: int32
    connected: bool
    
    func connect(self s: *TcpClient, host: string, port: uint16) bool {
        let AF_INET = 2
        let SOCK_STREAM = 1
        
        // Create socket
        s.fd = c.socket(AF_INET, SOCK_STREAM, 0)
        if s.fd < 0 {
            c.printf("Failed to create socket\n")
            return false
        }
        
        // Build sockaddr_in
        let addr = alloca(uint8, 16)
        c.memset(addr, 0, 16)
        
        // Set family
        let ptr_family = cast<*int16>(addr)
        *ptr_family = cast<int16>(AF_INET)
        
        // Set port
        let addr_int = cast<int64>(addr)
        let port_ptr = cast<*uint16>(addr_int + 2)
        *port_ptr = htons(port)
        
        // Set IP
        let ip_ptr = cast<*uint32>(addr_int + 4)
        *ip_ptr = c.inet_addr(host)
        
        // Connect
        let res = c.connect(s.fd, addr, 16)
        if res < 0 {
            c.printf("Failed to connect to %s:%d\n", host, port)
            c.close(s.fd)
            return false
        }
        
        s.connected = true
        return true
    }
    
    func send(self s: *TcpClient, data: string, len: int64) int64 {
        if !s.connected {
            return -1
        }
        return c.send(s.fd, data, len, 0)
    }
    
    func recv(self s: *TcpClient, buffer: *byte, len: int64) int64 {
        if !s.connected {
            return -1
        }
        return c.recv(s.fd, buffer, len, 0)
    }
    
    deinit(self s: *TcpClient) {
        if s.connected {
            c.close(s.fd)
            s.connected = false
        }
    }
}

func newTcpClient() *TcpClient {
    let client = alloca(TcpClient)
    client.fd = -1
    client.connected = false
    return client
}

func main() int32 {
    let client = newTcpClient()
    
    if client.connect("127.0.0.1", 8080) {
        c.printf("Connected!\n")
        
        client.send("hello", 5)
        c.printf("Sent message\n")
        
        let buf = alloca(uint8, 64)
        c.memset(buf, 0, 64)
        let n = client.recv(buf, 63)
        
        if n > 0 {
            c.printf("Received: %s\n", buf)
        }
    }
    
    return 0
}