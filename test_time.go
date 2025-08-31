package main

import (
    "fmt"
    "time"
)

func main() {
    now := time.Now()
    fmt.Println("Local time:", now)
    fmt.Println("UTC time:", now.UTC())
    fmt.Println("ISO format (local):", now.Format(time.RFC3339))
    fmt.Println("ISO format (UTC):", now.UTC().Format(time.RFC3339))
}
