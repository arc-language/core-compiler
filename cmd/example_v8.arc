// example_v8.arc
namespace main

extern c {
    // Socket API
    func printf(*byte, ...) int32
}

class Client {
    port: int32
    
    func connect(self c: *Client, host: string) bool {
        return true
    }
    
    deinit(self c: *Client) {
        // cleanup when ref count hits 0
    }
}

func newClient() *Client {
    let d = alloca(Client)
    d.port = 1000
    // initialize fields
    return d
}

func main() int32 {
    let client = newClient()

    c.printf("%d\n", client.port)

    return 0
}