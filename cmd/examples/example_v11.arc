namespace main

extern c {
    func printf(*byte, ...) int32
}

func main() int32 {
    // ------------------------------------------
    // 1. Test Range Loop (for x in start..end)
    // ------------------------------------------
    c.printf("--- Test 1: Range Loop (1..5) ---\n")
    for x in 1..5 {
        c.printf("x = %d\n", x)
    }

    // ------------------------------------------
    // 2. Test Logic inside Range Loop
    // ------------------------------------------
    c.printf("\n--- Test 2: Logic Match ---\n")
    let target = 30
    // Scanning a range to find a specific number
    for i in 25..35 {
        if i == target {
            c.printf("Found target value: %d\n", i)
        }
    }

    // ------------------------------------------
    // 3. Test C-Style Loop (init; cond; step)
    // ------------------------------------------
    c.printf("\n--- Test 3: C-Style Loop ---\n")
    for let j = 0; j < 3; j = j + 1 {
        c.printf("j = %d\n", j)
    }

    // ------------------------------------------
    // 4. Test While-Style Loop (cond only)
    // ------------------------------------------
    c.printf("\n--- Test 4: While-Style Loop ---\n")
    let k = 3
    for k > 0 {
        c.printf("k = %d\n", k)
        k = k - 1
    }

    return 0
}