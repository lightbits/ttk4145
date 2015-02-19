package main

import (
    "time"
    "log"
)

type network_message struct {
    Protocol     uint32
    Length       uint32
    UserData     []byte
    EndDelimiter uint32
}

// The reason we use a channel for OUTGOING messages
// is because the network might be busy reading a packet,
// but we don't want to block?
func NetworkInit(OutgoingUpdate chan client_update,
                 IncomingUpdate chan master_update) {

    SendChannel := make(chan network_message)
    RecvChannel := make(chan network_message)
    go FakeNetwork(SendChannel, RecvChannel)

    for {
        select {
        case Request := <- OutgoingUpdate:

            // TODO: Send request to master over UDP
        case Packet := <- RecvChannel:
            // Parse packet, verify protocol
            // acceptance test

            // Dummy code
            OrderA := order{
                FromFloor: 0,
                ToFloor: 1,
                Type: order_up,
                TakenBy: lift_id{0xabad1dea, 0xbeef},
            }

            OrderB := order{
                FromFloor: 1,
                ToFloor: 2,
                Type: order_down,
                TakenBy: lift_id{0xaabababa, 0xbeef},
            }

            PendingOrders := []order{OrderA, OrderB}
            Update := master_update{PendingOrders}
            IncomingUpdate <- Update
        }
    }
}
