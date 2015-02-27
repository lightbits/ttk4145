package main

import (
    "fmt"
    "time"
    "./network"
)

const CLIENT_UPDATE_INTERVAL = 1 * time.Second

func main() {
    outgoing := make(chan network.ClientUpdate)
    incoming := make(chan network.MasterUpdate)
    go network.InitClient(outgoing, incoming)

    ticker := time.NewTicker(CLIENT_UPDATE_INTERVAL)

    for {
        select {
        case <- ticker.C:
            fmt.Println("Client send update")
            outgoing <- network.ClientUpdate{Request: "Hello master!"}

        case update := <- incoming:
            fmt.Println("Master said:", update.ActiveOrders)
        }
    }
}
