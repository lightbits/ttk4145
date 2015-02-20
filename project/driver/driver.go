package driver

import (
    "log"
    "time"
)

type ButtonType int
const (
    ButtonUp ButtonType = iota
    ButtonDown
    ButtonOut
)
type ButtonEvent struct {
    Floor int
    Type  ButtonType
}

type ReachedFloorEvent struct {
    FloorIndex int
}

type StopButtonEvent struct {
    IsPressed bool
}

type ObstructionEvent struct {
    IsObstructed bool
}

type MotorDirection int
const (
    MotorDirectionUp   = 1
    MotorDirectionStop = 0
    MotorDirectionDown = -1
)

type edge_detect int
const (
    edge_detect_rising = iota
    edge_detect_falling
    edge_detect_both
)

type io_event struct {
    bit    int
    is_set bool
}

type bit_state struct {
    channel int
    was_set bool
}

func poll(bits [N_FLOORS]bit_state, event chan io_event, edge edge_detect) {
    for {
        for i := 0; i < len(bits); i++ {
            is_set := io_read_bit(bits[i].channel) == 1

            is_rising := !bits[i].was_set && is_set
            is_falling := bits[i].was_set && !is_set

            if ((edge == edge_detect_rising && is_rising) ||
                (edge == edge_detect_falling && is_falling) ||
                (edge == edge_detect_both && (is_rising || is_falling))) {
                event <- io_event{i, is_set}
            }
            bits[i].was_set = is_set
        }

        // If we hammer the io module too much it actually fails to read.
        // So we sleep a little bit.
        time.Sleep(1 * time.Millisecond)
    }
}

func MotorUp() {
    io_clear_bit(MOTORDIR)
    io_write_analog(MOTOR, 2800)
}

func MotorDown() {
    io_set_bit(MOTORDIR)
    io_write_analog(MOTOR, 2800)
}

func MotorStop() {
    io_write_analog(MOTOR, 0)
}

func SetButtonLamp(btn ButtonType, floor int, set bool) {
    lights := up_lights
    if floor >= N_FLOORS {
        log.Fatal("Tried to set light on non-existent floor")
    }

    switch btn {
    case ButtonUp:
        lights = up_lights
        if floor >= N_FLOORS - 1 {
            log.Fatal("Tried to set light on non-existent floor")
        }
    case ButtonDown:
        lights = down_lights
        if floor == 0 {
            log.Fatal("Tried to set light on non-existent floor")
        }
    case ButtonOut:
        lights = out_lights
    }

    if set {
        io_set_bit(lights[floor])
    } else {
        io_clear_bit(lights[floor])
    }
}

func SetDoorOpenLamp(on bool) {
    if on {
        io_set_bit(LIGHT_DOOR_OPEN);
    } else {
        io_clear_bit(LIGHT_DOOR_OPEN);
    }
}

func SetStopLamp(on bool) {
    if on {
        io_set_bit(LIGHT_STOP)
    } else {
        io_clear_bit(LIGHT_STOP);
    }
}

func SetFloorIndicator(floor int) {
    if (floor & 0x02 != 0) {
        io_set_bit(LIGHT_FLOOR_IND1);
    } else {
        io_clear_bit(LIGHT_FLOOR_IND1);
    }

    if (floor & 0x01 != 0) {
        io_set_bit(LIGHT_FLOOR_IND2);

    } else {
        io_clear_bit(LIGHT_FLOOR_IND2);
    }
}

func Init(button_pressed chan ButtonEvent,
          floor_reached  chan ReachedFloorEvent,
          stop_pressed   chan StopButtonEvent,
          obstruction    chan ObstructionEvent) {

    if (!io_init()) {
        log.Fatal("Failed to initialize driver")
    }

    // Zero all floor button lamps
    for i := 0; i < N_FLOORS; i++ {
        if i != 0 {
            SetButtonLamp(ButtonDown, i, false)
        }
        if i != N_FLOORS - 1 {
            SetButtonLamp(ButtonUp, i, false)
        }
        SetButtonLamp(ButtonOut, i, false)
    }

    // Clear stop lamp, door open lamp, and set floor indicator to ground floor
    SetDoorOpenLamp(false)
    SetStopLamp(false)
    SetFloorIndicator(0)

    // Drive to bottom floor
    MotorDown()
    for io_read_bit(SENSOR_FLOOR1) != 1 { }
    MotorStop()

    up_ch   := make(chan io_event)
    down_ch := make(chan io_event)
    out_ch  := make(chan io_event)
    flr_ch  := make(chan io_event)
    stp_ch  := make(chan io_event)
    obs_ch  := make(chan io_event)

    if (button_pressed != nil) {
        go poll(up_buttons, up_ch, edge_detect_rising)
        go poll(down_buttons, down_ch, edge_detect_rising)
        go poll(out_buttons, out_ch, edge_detect_rising)
    }

    if (floor_reached != nil) {
        go poll(floor_sensors, flr_ch, edge_detect_rising)
    }

    if (stop_pressed != nil) {
        go poll(stop_buttons, stp_ch, edge_detect_rising)
    }

    if (obstruction != nil) {
        go poll(obstructions, obs_ch, edge_detect_both)
    }

    for {
        select {
        case e := <-up_ch:
            button_pressed <- ButtonEvent{e.bit, ButtonUp}
        case e := <-down_ch:
            button_pressed <- ButtonEvent{e.bit, ButtonDown}
        case e := <-out_ch:
            button_pressed <- ButtonEvent{e.bit, ButtonOut}
        case e := <-flr_ch:
            floor_reached <- ReachedFloorEvent{e.bit}
        case e := <- stp_ch:
            stop_pressed <- StopButtonEvent{e.is_set}
        case e := <- obs_ch:
            obstruction <- ObstructionEvent{e.is_set}
        }
    }
}