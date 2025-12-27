namespace main

// Basic function
func add(a: int32, b: int32) int32 {
    return a + b
}

// No return
func print_message(msg: string) {
    // implementation
}

// Async function with return
async func fetch_data(url: string) string {
    let response = await http_get(url)
    return response
}

// Async function without return
async func process_items(items: vector<string>) {
    for item in items {
        await process_item(item)
    }
}

// Placeholder async functions for testing
async func http_get(url: string) string {
    return "response"
}

async func process_item(item: string) {
    // implementation
}

func main() int32 {
    let result = add(5, 10)
    print_message("Hello")
    
    return 0
}