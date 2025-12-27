namespace main

func main() int32 {
    // C-style for loop
    for let i = 0; i < 10; i = i + 1 {
        // loop body
    }
    
    // With increment operator
    for let j = 0; j < 10; j++ {
        // loop body
    }
    
    // While-style for loop
    let k = 5
    for k > 0 {
        k--
    }
    
    // Infinite loop with break
    let counter = 0
    for {
        counter++
        if counter >= 10 {
            break
        }
    }
    
    // For-in with vector
    let items: vector<int32> = {1, 2, 3, 4, 5}
    for item in items {
        // use item
    }
    
    // For-in with map
    let scores: map<string, int32> = {"alice": 100, "bob": 95}
    for key, value in scores {
        // use key and value
    }
    
    // For-in with range
    for i in 0..10 {
        // i goes from 0 to 9
    }
    
    // Break and continue
    for let m = 0; m < 10; m++ {
        if m == 3 {
            continue
        }
        if m == 7 {
            break
        }
    }
    
    return 0
}