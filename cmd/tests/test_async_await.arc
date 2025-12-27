namespace main

async func fetch_data(url: string) string {
    return "data"
}

async func task1() int32 {
    return 1
}

async func task2() int32 {
    return 2
}

async func check_status() bool {
    return true
}

async func main_async() {
    // Await async function call
    let data = await fetch_data("https://api.example.com")
    
    // Multiple awaits
    let result1 = await task1()
    let result2 = await task2()
    
    // Await in expressions
    if await check_status() {
        // do something
    }
}

func main() int32 {
    return 0
}