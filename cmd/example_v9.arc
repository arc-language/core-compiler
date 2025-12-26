// example_v9.arc
namespace main // same as golang package but named namespace

extern c {
    // Socket API
    func printf(*byte, ...) int32
}

struct Client {
    port: int32
    
    func connect(self s: *Client, host: string) bool {
        return true
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
    client.port = 50

    c.printf("%d\n", client.port)

    return 0
}