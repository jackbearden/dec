package client

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync"

	"dec/internal/queue"
	"dec/messages"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/mailbox"
	"github.com/AsynkronIT/protoactor-go/remote"
	"github.com/olekukonko/tablewriter"
)

const NOT_FOUND uint32 = math.MaxUint32

type Client struct {
	Wg                *sync.WaitGroup
	ElevatorCount     int
	ElevatorPidList   *[]*actor.PID
	ElevatorStatusMap *sync.Map
	ClientActor       *ClientActor
	PickupQueue       *queue.Queue
}

type ClientActor struct {
	Client *Client
	PID    *actor.PID
}

type ElevatorStatus struct {
	Id    uint32
	Floor uint32
	State int32
	Goal  int32
}

type StatusRequestOpt struct {
	BroadcastAll bool
	SinglePID    int
}

type PickupRequestItem struct {
	Floor uint32
	State int32
}

func (ca *ClientActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *messages.StatusResponse:
		// Bit inefficient with memory here
		ca.Client.ElevatorStatusMap.Store(
			msg.Id,
			&ElevatorStatus{
				Id:    msg.Id,
				Floor: msg.Floor,
				Goal:  msg.Goal,
				State: msg.State,
			},
		)

		ca.Client.Wg.Done()
	}
}

func newClientActor(c *Client) actor.Producer {
	return func() actor.Actor {
		return &ClientActor{
			Client: c,
		}
	}
}

func NewClient(bind string, elevatorCount int) *Client {
	remote.Start(bind)
	elevatorPidList := make([]*actor.PID, elevatorCount)
	for i := 0; i < elevatorCount; i++ {
		hostname := "127.0.0.1"
		port := 9000 + i
		binding := fmt.Sprintf("%s:%d", hostname, port)

		elevatorPidList[i] = actor.NewPID(binding, string(i))
	}
	client := &Client{
		Wg:                &sync.WaitGroup{},
		ElevatorCount:     elevatorCount,
		ElevatorPidList:   &elevatorPidList,
		ElevatorStatusMap: &sync.Map{},
		ClientActor:       &ClientActor{},
		PickupQueue:       queue.New(),
	}
	props := actor.FromProducer(newClientActor(client)).
		WithMailbox(mailbox.Bounded(10000))
	client.ClientActor.PID = actor.Spawn(props)
	log.Println("client started")

	return client
}

func (client *Client) SendStatusRequest(opt StatusRequestOpt) {
	msg := &messages.StatusRequest{Sender: client.ClientActor.PID}

	if opt.BroadcastAll == true {
		for _, elevator := range *client.ElevatorPidList {
			client.Wg.Add(1)
			elevator.Tell(msg)
		}
	} else {
		client.Wg.Add(1)
		(*client.ElevatorPidList)[opt.SinglePID].Tell(msg)
	}

	client.Wg.Wait()
}

func (client *Client) SendPickupRequest(floor uint32, state int32) {
	var shortestProximity int32 = math.MaxInt32
	var selectedId uint32 = NOT_FOUND

	// Try to optimize and have nearby elevator pick up
	client.ElevatorStatusMap.Range(func(k, v interface{}) bool {
		e := v.(*ElevatorStatus)
		// See if we are going the same direction
		if state == e.State {
			// Select only those that are in range
			if (state == 1 && e.Floor <= floor) ||
				(state == -1 && e.Floor >= floor) {
				// Get id of closest one to this floor
				prox := int32(math.Abs(float64(floor - e.Floor)))
				// Update it if closer
				if prox < shortestProximity {
					shortestProximity = prox
					selectedId = e.Id
				}

			}
		}

		return true
	})

	// Try to assign to empty car if none found
	if selectedId == NOT_FOUND {
		client.ElevatorStatusMap.Range(func(k, v interface{}) bool {
			e := v.(*ElevatorStatus)

			if e.State == 0 {
				shortestProximity = 0
				selectedId = e.Id
				return false
			}

			return true
		})
	}

	if selectedId != NOT_FOUND {
		msg := &messages.PickupRequest{
			Sender: client.ClientActor.PID,
			Floor:  floor,
			State:  state,
		}
		client.Wg.Add(1)
		(*client.ElevatorPidList)[selectedId].Tell(msg)
		client.Wg.Wait()
	} else {
		// Add to queue when no cars are available
		client.PickupQueue.PushBack(PickupRequestItem{Floor: floor, State: state})
		log.Println("all cars are busy!")
	}
}

func (client *Client) SendUpdateRequest(id int, goal uint32, state int32) {
	msg := &messages.UpdateRequest{
		Sender: client.ClientActor.PID,
		Goal:   goal,
		State:  state,
	}
	client.Wg.Add(1)
	(*client.ElevatorPidList)[id].Tell(msg)
	client.Wg.Wait()
}

func (client *Client) SendStepRequest() {
	msg := &messages.StepRequest{
		Sender: client.ClientActor.PID,
	}
	for _, elevator := range *client.ElevatorPidList {
		client.Wg.Add(1)
		elevator.Tell(msg)
	}
	client.Wg.Wait()

	// Process queue items
	for amt := client.PickupQueue.Len(); amt > 0; amt-- {
		pqi := client.PickupQueue.PopFront().(PickupRequestItem)
		client.SendPickupRequest(pqi.Floor, pqi.State)
	}
}

func (client *Client) PrintCurrentStatus() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Floor", "Goal", "State"})

	client.ElevatorStatusMap.Range(func(k, v interface{}) bool {
		e := v.(*ElevatorStatus)
		table.Append([]string{
			strconv.Itoa(int(e.Id)),
			strconv.Itoa(int(e.Floor)),
			strconv.Itoa(int(e.Goal)),
			strconv.Itoa(int(e.State)),
		})

		return true
	})
	table.Render()
}
