namespace main

func main() int32 {
    // Empty vector
    let v1: vector<int32> = {}
    
    // Initialized vector
    let v2: vector<int32> = {1, 2, 3, 4, 5}
    
    // Vector with inference
    let v3 = {10, 20, 30}
    
    // Empty map
    let m1: map<string, int32> = {}
    
    // Initialized map
    let m2: map<string, int32> = {"alice": 100, "bob": 95}
    
    // Map with inference
    let m3 = {"host": "localhost", "port": "8080"}
    
    return 0
}