package main

import (
    "github.com/bluenviron/mediamtx"
)

func main() {
    config := mediamtx.NewConfig()
    server := mediamtx.NewServer(config)

    if err := server.Start(); err != nil {
        panic(err)
    }
}

