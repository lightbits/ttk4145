package lift

import (
    "time"
    "../driver"
    "../logger"
)

const deadline_period = 5 * driver.NumFloors * time.Second
const door_period = 3 * time.Second
var last_passed_floor int

type state_t int
const (
    idle state_t = iota
    doorOpen
    moving
)

func GetLastPassedFloor() int {
    return last_passed_floor
}

func Init(
    completed_floor  chan <- int,
    missed_deadline  chan <- bool,
    floor_reached    <- chan int,
    new_target_floor <- chan int,
    stop_button      <- chan bool,
    obstruction      <- chan bool) {

    deadline_timer := time.NewTimer(deadline_period)
    deadline_timer.Stop()

    door_timer := time.NewTimer(door_period)
    door_timer.Stop()

    state := idle
    last_passed_floor = 0
    target_floor := driver.InvalidFloor

    for {
        select {
        case <- door_timer.C:
            switch (state) {
                case doorOpen:
                    println(logger.Info, "Door timer @ doorOpen")
                    driver.CloseDoor()
                    state = idle
                    completed_floor <- target_floor
                    target_floor = driver.InvalidFloor
                    deadline_timer.Stop()
                case idle:    println(logger.Debug, "Door timer @ idle")
                case moving:  println(logger.Debug, "Door timer @ moving")
            }

        case <- deadline_timer.C:
            missed_deadline <- true

        case floor := <- new_target_floor:
            if target_floor != floor {
                deadline_timer.Reset(deadline_period)
            }
            target_floor = floor
            switch (state) {
                case idle:
                    println(logger.Info, "New order @ idle")
                    if target_floor == driver.InvalidFloor {
                        break
                    } else if target_floor > last_passed_floor {
                        state = moving
                        driver.MotorUp()
                    } else if target_floor < last_passed_floor {
                        state = moving
                        driver.MotorDown()
                    } else {
                        door_timer.Reset(door_period)
                        driver.OpenDoor()
                        driver.MotorStop()
                        state = doorOpen
                    }
                case moving:   println(logger.Debug, "New order @ moving")
                case doorOpen: println(logger.Debug, "New order @ doorOpen")
            }

        case floor := <- floor_reached:
            last_passed_floor = floor
            switch (state) {
                case moving:
                    println(logger.Info, "Reached floor", floor, "@ moving")
                    driver.SetFloorIndicator(floor)
                    if target_floor == driver.InvalidFloor {
                        break
                    } else if target_floor > last_passed_floor {
                        state = moving
                        driver.MotorUp()
                    } else if target_floor < last_passed_floor {
                        state = moving
                        driver.MotorDown()
                    } else {
                        door_timer.Reset(door_period)
                        driver.OpenDoor()
                        driver.MotorStop()
                        state = doorOpen
                    }
                case idle:     println(logger.Info, "Reached floor", floor, "@ idle")
                case doorOpen: println(logger.Info, "Reached floor", floor, "@ doorOpen")
            }

        case <- stop_button: // Ignoring
        case <- obstruction: // Ignoring
        }
    }
}

func println(level logger.Level, args...interface{}) {
    logger.Println(level, "LIFT", args...)
}
