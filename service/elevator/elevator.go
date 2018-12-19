package elevator

import (
	"log"
	"math/bits"

	"dec/messages"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/mailbox"
	"github.com/AsynkronIT/protoactor-go/remote"
)

const (
	ASCENDING  = 1
	DESCENDING = -1
	IDLE       = 0
)

type Elevator struct {
	Id                uint
	BitVector         uint16
	Floor             uint16
	State             int
	LockedPickupFloor uint16
	LockedDirection   int
}

func (e *Elevator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.StatusRequest:
		msg.Sender.Tell(e.newStatusResponse())
	case *messages.UpdateRequest:
		e.Update(int(msg.Goal), int(msg.State))
		msg.Sender.Tell(e.newStatusResponse())
	case *messages.PickupRequest:
		e.Pickup(uint16(msg.Floor), int(msg.State))
		msg.Sender.Tell(e.newStatusResponse())
	case *messages.StepRequest:
		e.Step()
		msg.Sender.Tell(e.newStatusResponse())
	}
}

func NewElevator(id uint) *Elevator {
	return &Elevator{
		Id:        id,
		BitVector: 0,
		Floor:     (1 << 0),
		State:     IDLE,
	}
}

func newElevatorActor(id uint) actor.Producer {
	return func() actor.Actor {
		return NewElevator(id)
	}
}

func NewElevatorService(bind string, id uint) {
	remote.Start(bind)
	props := actor.FromProducer(newElevatorActor(id)).
		WithMailbox(mailbox.Bounded(10000))
	actor.SpawnNamed(props, string(id))

	log.Println("elevator", id, "ready")
}

func (e *Elevator) newStatusResponse() *messages.StatusResponse {
	return &messages.StatusResponse{
		Id:    uint32(e.Id),
		Floor: uint32(e.GetCurrentFloor()),
		Goal:  int32(e.FindNextGoal()),
		State: int32(e.State),
	}
}

func (e *Elevator) Pickup(pickupFloor uint16, direction int) {
	if e.State == IDLE {
		e.State = e.GetPickupDirection(pickupFloor)
		e.LockedPickupFloor = (1 << pickupFloor)
		e.LockedDirection = direction
	}
	if e.GetCurrentFloor() != pickupFloor {
		e.SetBit(pickupFloor)
	}
}

func (e *Elevator) IsLocked() bool {
	return e.LockedPickupFloor != 0
}

func (e *Elevator) GetPickupDirection(pickupFloor uint16) int {
	if e.GetCurrentFloor() <= pickupFloor {
		return 1
	}

	return -1
}

func (e *Elevator) FindNextGoal() int {
	var mask uint16
	var lsb, msb int
	var floorBit uint16 = e.GetCurrentFloor()

	switch e.State {
	case ASCENDING:
		mask = ^((1 << (floorBit + 1)) - 1)
		lsb = LSB16(e.BitVector & mask)

		return lsb
	case DESCENDING:
		mask = ((1 << floorBit) - 1)
		msb = MSB16(e.BitVector & mask)

		return msb
	}

	return -1
}

func (e *Elevator) GetCurrentFloor() uint16 {
	return uint16(bits.TrailingZeros16(e.Floor))
}

func (e *Elevator) SetBit(n uint16) {
	e.BitVector |= (1 << n)
}

func (e *Elevator) UnsetBit(n uint16) {
	e.BitVector &= ^(1 << n)
}

func (e *Elevator) Move() {
	switch e.State {
	case ASCENDING:
		e.MoveUp()
	case DESCENDING:
		e.MoveDown()
	}

	if e.HasGoalAtCurrentFloor() {
		e.UnsetBit(e.GetCurrentFloor())
	}

	return
}

func (e *Elevator) MoveUp() {
	n := MSB16(e.Floor)

	e.Floor = (1 << uint16(n+1))

	return
}

func (e *Elevator) MoveDown() {
	n := LSB16(e.Floor)

	e.Floor = (1 << uint16(n-1))

	return
}

func (e *Elevator) Step() {
	// No action required when idling
	if e.State == IDLE {
		return
	}

	// Move towards the next goal
	if e.HasGoals() {
		if e.FindNextGoal() != -1 {
			e.Move()
		} else if e.FindNextGoal() == -1 {
			// Toggle state if we have goals but not current direction
			e.ToggleState()
		}
	}

	// Go idle if no more goals
	if !e.HasGoals() {
		e.State = IDLE
	}

	// Check if the pickup req arrived at the intended floor
	if e.IsLocked() && e.LockedPickupFloor == e.Floor {
		e.State = e.LockedDirection
		e.LockedPickupFloor = 0
		e.LockedDirection = 0
	}
}

func (e *Elevator) ToggleState() {
	if e.State == ASCENDING {
		e.State = DESCENDING
	} else {
		e.State = ASCENDING
	}
}

func (e *Elevator) HasGoals() bool {
	return (bits.OnesCount16(e.BitVector) > 0)
}

func (e *Elevator) HasGoalAtCurrentFloor() bool {
	return (e.BitVector & e.Floor) > 0
}

func (e *Elevator) Update(goal int, state int) {
	if e.State == IDLE {
		e.State = state
	}
	e.SetBit(uint16(goal))
}

func (e *Elevator) Status() []int {
	return []int{int(e.GetCurrentFloor()), int(e.FindNextGoal()), e.State}
}
