namespace main

// Basic struct
struct Point {
    x: int32
    y: int32
}

// Struct with inline methods
struct Rectangle {
    width: int32
    height: int32
    
    func area(self r: Rectangle) int32 {
        return r.width * r.height
    }
    
    func resize(self r: *Rectangle, w: int32, h: int32) {
        r.width = w
        r.height = h
    }
}

// Struct with mutating methods
struct Counter {
    count: int32
    
    mutating increment(self c: *Counter) {
        c.count++
    }
    
    mutating add(self c: *Counter, value: int32) {
        c.count += value
    }
    
    func get_count(self c: Counter) int32 {
        return c.count
    }
}

// Flat methods (declared outside)
struct Circle {
    radius: int32
}

func diameter(self c: Circle) int32 {
    return c.radius * 2
}

mutating set_radius(self c: *Circle, r: int32) {
    c.radius = r
}

func main() int32 {
    // Initialization
    let p1: Point = Point{x: 10, y: 20}
    let p2 = Point{x: 5, y: 15}
    let p3: Point = Point{}
    
    // Field access
    let x = p1.x
    p1.y = 30
    
    // Method calls
    let rect = Rectangle{width: 10, height: 20}
    let area = rect.area()
    rect.resize(15, 25)
    
    // Mutating methods
    let counter = Counter{count: 0}
    counter.increment()
    counter.add(5)
    let value = counter.get_count()
    
    // Flat methods
    let circle = Circle{radius: 5}
    let d = circle.diameter()
    circle.set_radius(10)
    
    return 0
}