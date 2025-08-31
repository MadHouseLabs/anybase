package main

import (
    "encoding/json"
    "fmt"
    "time"
)

type TestStruct struct {
    CreatedAt time.Time `json:"created_at"`
}

func main() {
    ts := TestStruct{
        CreatedAt: time.Now(),
    }
    
    data, _ := json.Marshal(ts)
    fmt.Println("JSON output:", string(data))
    fmt.Println("Local time:", ts.CreatedAt)
    fmt.Println("UTC time:", ts.CreatedAt.UTC())
}
