namespace main

// Basic class
class Client {
    name: string
    port: int32
}

// Class with inline methods
class Connection {
    host: string
    port: int32
    
    func connect(self c: *Connection, timeout: int32) bool {
        return true
    }
    
    async func fetch_data(self c: *Connection) string {
        return "data"
    }
    
    deinit(self c: *Connection) {
        // cleanup when ref count hits 0
    }
}

// Class with flat methods
class Server {
    port: int32
}

func start(self s: *Server) bool {
    return true
}

func stop(self s: *Server) {
    // implementation
}

deinit(self s: *Server) {
    // cleanup
}

func main() int32 {
    let client = Client{name: "test", port: 8080}
    let conn = Connection{host: "localhost", port: 8080}
    let success = conn.connect(30)
    
    let server = Server{port: 9000}
    server.start()
    server.stop()
    
    return 0
}