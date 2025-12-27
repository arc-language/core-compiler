namespace main

import "std/io"

struct Point {
    x: int32
    y: int32
    
    func distance_squared(self p: Point) int32 {
        return p.x * p.x + p.y * p.y
    }
    
    mutating move(self p: *Point, dx: int32, dy: int32) {
        p.x += dx
        p.y += dy
    }
}

class Connection {
    host: string
    port: int32
    connected: bool
    
    func connect(self c: *Connection) bool {
        c.connected = true
        return true
    }
    
    deinit(self c: *Connection) {
        // cleanup
    }
}

func process_data(data: vector<int32>) int32 {
    let sum: int32 = 0
    for value in data {
        sum += value
    }
    return sum
}

func main() int32 {
    // Test variables and types
    let count: int32 = 0
    const max: int32 = 100
    
    // Test collections
    let numbers: vector<int32> = {1, 2, 3, 4, 5}
    let config: map<string, int32> = {"timeout": 30, "retries": 3}
    
    // Test structs
    let point = Point{x: 10, y: 20}
    let dist = point.distance_squared()
    point.move(5, -3)
    
    // Test classes
    let conn = Connection{host: "localhost", port: 8080, connected: false}
    conn.connect()
    
    // Test control flow
    for let i = 0; i < 10; i++ {
        if i % 2 == 0 {
            continue
        }
        count++
    }
    
    // Test function call
    let total = process_data(numbers)
    
    // Test intrinsics
    let buffer = alloca(byte, 256)
    memset(buffer, 0, 256)
    
    return 0
}