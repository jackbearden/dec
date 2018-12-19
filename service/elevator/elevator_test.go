package elevator

import (
	_ "fmt"
	"testing"
)

func TestPickup(t *testing.T) {
	e := NewElevator(0)
	// a pickup request comes in at foor 2 going up
	e.Pickup(2, 1)
	status := e.Status()
	if status[0] != 0 ||
		status[1] != 2 ||
		status[2] != 1 {
		t.Error("Expected {0, 2, 1}, got ", status)
	}
	// we lock the car from other pickup requests
	if !e.IsLocked() {
		t.Error("Expected true, got ", e.IsLocked())
	}
	e.Step()
	status = e.Status()
	if status[0] != 1 ||
		status[1] != 2 ||
		status[2] != 1 {
		t.Error("Expected {1, 2, 1}, got ", status)
	}
	// step up to our goal
	// since the up button was pressed originally,
	// it is retained for this step
	e.Step()
	status = e.Status()
	if status[0] != 2 ||
		status[1] != -1 ||
		status[2] != 1 {
		t.Error("Expected {2, -1, 1}, got ", status)
	}
	// since no action was taken, the car goes idle
	e.Step()
	status = e.Status()
	if status[0] != 2 ||
		status[1] != -1 ||
		status[2] != 0 {
		t.Error("Expected {2, -1, 0}, got ", status)
	}
	if e.IsLocked() {
		t.Error("Expected false, got ", e.IsLocked())
	}
}

func TestStep(t *testing.T) {
	e := NewElevator(0)
	status := e.Status()
	// idle status with no goals
	if status[0] != 0 ||
		status[1] != -1 ||
		status[2] != 0 {
		t.Error("Expected {0, -1, 0}, got ", status)
	}
	// button pressed
	e.Update(2, ASCENDING)
	status = e.Status()
	if status[0] != 0 ||
		status[1] != 2 ||
		status[2] != 1 {
		t.Error("Expected {0, 2, 1}, got ", status)
	}
	// we step up to floor 1 as our goal is 2
	e.Step()
	status = e.Status()
	if status[0] != 1 ||
		status[1] != 2 ||
		status[2] != 1 {
		t.Error("Expected {1, 2, 1}, got ", status)
	}
	// we step up to floor 2 and arrive at our goal
	e.Step()
	status = e.Status()
	if status[0] != 2 ||
		status[1] != -1 ||
		status[2] != 0 {
		t.Error("Expected {2, -1, 0}, got ", status)
	}
	// lobby is pressed but floor 3 is also pressed accidently
	e.Update(0, DESCENDING)
	e.SetBit(3)
	status = e.Status()
	if status[0] != 2 ||
		status[1] != 0 ||
		status[2] != -1 {
		t.Error("Expected {2, 0, -1}, got ", status)
	}
	// we step down to floor 1
	e.Step()
	status = e.Status()
	if status[0] != 1 ||
		status[1] != 0 ||
		status[2] != -1 {
		t.Error("Expected {1, 0, -1}, got ", status)
	}
	// we step down to floor 0 and arrive at our goal
	// as we still have a goal, we do not switch to an idle state
	// and remain in descending mode until the next step
	e.Step()
	status = e.Status()
	if status[0] != 0 ||
		status[1] != -1 ||
		status[2] != -1 {
		t.Error("Expected {0, -1, -1}, got ", status)
	}

	// no more decending goals, so state changes to up again
	e.Step()
	status = e.Status()
	if status[0] != 0 ||
		status[1] != 3 ||
		status[2] != 1 {
		t.Error("Expected {0, 3, 1}, got ", status)
	}
}

func TestLSB(t *testing.T) {
	e := NewElevator(0)

	lsb := LSB16(e.BitVector)
	if lsb != -1 {
		t.Error("Expected -1, got ", lsb)
	}

	e.SetBit(3)
	e.SetBit(4)
	e.SetBit(6)

	lsb = LSB16(e.BitVector)
	if lsb != 3 {
		t.Error("Expected 3, got ", lsb)
	}
}

func TestMSB(t *testing.T) {
	e := NewElevator(0)
	msb := MSB16(e.BitVector)

	if msb != -1 {
		t.Error("Expected -1, got ", msb)
	}

	e.SetBit(3)
	e.SetBit(4)
	e.SetBit(6)

	msb = MSB16(e.BitVector)
	if msb != 6 {
		t.Error("Expected 6, got ", msb)
	}
}

func TestSetBit(t *testing.T) {
	e := NewElevator(0)

	e.SetBit(3)
	e.SetBit(4)
	e.SetBit(7)

	if e.BitVector != 152 {
		t.Error("Expected 152, got ", e.BitVector)
	}
}

func TestHasGoals(t *testing.T) {
	e := NewElevator(0)
	if e.HasGoals() {
		t.Error("Expected false, got ", e.BitVector)
	}

	e.SetBit(3)
	e.SetBit(4)
	e.SetBit(7)

	if !e.HasGoals() {
		t.Error("Expected true, got ", e.BitVector)
	}

	e.UnsetBit(3)
	e.UnsetBit(4)
	e.UnsetBit(7)

	if e.HasGoals() {
		t.Error("Expected false, got ", e.BitVector)
	}
}

func TestMoveUp(t *testing.T) {
	e := NewElevator(0)
	if e.Floor != 1 {
		t.Error("Expected 1, got ", e.Floor)
	}

	e.MoveUp()
	if e.Floor != 2 {
		t.Error("Expected 2, got ", e.Floor)
	}

	e.MoveUp()
	if e.Floor != 4 {
		t.Error("Expected 4, got ", e.Floor)
	}
}

func TestMoveDown(t *testing.T) {
	e := NewElevator(0)
	if e.Floor != 1 {
		t.Error("Expected 1, got ", e.Floor)
	}

	e.MoveUp()
	if e.Floor != 2 {
		t.Error("Expected 2, got ", e.Floor)
	}

	e.MoveUp()
	if e.Floor != 4 {
		t.Error("Expected 4, got ", e.Floor)
	}

	e.MoveDown()
	if e.Floor != 2 {
		t.Error("Expected 2, got ", e.Floor)
	}

	e.MoveDown()
	if e.Floor != 1 {
		t.Error("Expected 1, got ", e.Floor)
	}
}

func TestFindNextDescGoal(t *testing.T) {
	e := NewElevator(0)

	e.BitVector = 199 // 1100 0111
	e.Floor = (1 << 4)
	e.State = DESCENDING
	n := e.FindNextGoal()
	if n != 2 {
		t.Error("Expected 2, got ", n)
	}
	e.UnsetBit(2)
	n = e.FindNextGoal()
	if n != 1 {
		t.Error("Expected 1, got ", n)
	}
	e.UnsetBit(1)
	n = e.FindNextGoal()
	if n != 0 {
		t.Error("Expected 0, got ", n)
	}
}

func TestFindNextAscGoal(t *testing.T) {
	e := NewElevator(0)

	e.SetBit(0)
	e.SetBit(3)
	e.Floor = (1 << 3)
	e.SetBit(5)
	e.SetBit(6)
	e.State = ASCENDING
	n := e.FindNextGoal()
	if n != 5 {
		t.Error("Expected 5, got ", n)
	}
	e.UnsetBit(5)
	n = e.FindNextGoal()
	if n != 6 {
		t.Error("Expected 6, got ", n)
	}
	n = e.FindNextGoal()
	if n != 6 {
		t.Error("Expected 6, got ", n)
	}
}

func TestUnsetBit(t *testing.T) {
	e := NewElevator(0)

	e.SetBit(3)
	e.SetBit(4)
	e.SetBit(7)

	e.UnsetBit(3)
	e.UnsetBit(4)
	e.UnsetBit(7)

	if e.BitVector != 0 {
		t.Error("Expected 0, got ", e.BitVector)
	}
}
