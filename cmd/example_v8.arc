// example_v4.arc
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

func main() int32 {

    let client = Client{}

    return 0
}