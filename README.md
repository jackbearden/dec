# Distributed Elevator Control

## Summary
A high performant, distributed elevator control system focused on scalability.

## Features
- Decoupled concurrency
- Fault-tolerant
- Optimized scheduling
- Powerful CLI
- RPC serialized by Protobuf
- Docker for convenience and scale
- 16 x 16 capacity (Floors x Elevator processes)

## Architecture
The client actor is used to communicate with many elevator actors. The client executes a command with serialized data and sends that message over an RPC connection to the other elevator actor(s) in the cluster. After receiving an instruction, the elevator runs the task and replies to the message. Scalability was deeply considered in this architecture, as the current implementation can support over 2 million messages per second. Furthermore, the elevators run on separate processes and communicate over the network to truly decouple the system and components from node failure while also providing a powerful concurrent solution that is ready to scale. When the needs of the system exceed 2 million messages per second, replica managers or an elevator promotion strategy can be implemented to increase throughput further.

The elevator entity was designed using a 16-bit unsigned integer to store and calculate the goals — also, three states: ascending, descending and idle. I found this to be the most straightforward design as it makes updates trivial to schedule while still being incredibly efficient. When the elevator is moving, it continues in that direction until it has reached the limit or no further destinations remain in that orientation. In the event of no further goals, it then switches to the opposite orientation and proceeds to the next goal or goes idle and waits for the next request.

I have made improvements to the scheduler to optimize shorter user wait times, faster destination times, and avoiding unnecessary operating costs. During a pickup request, the scheduler attempts to find nearby cars going the same direction and with the closest proximity to the floor of the requestee. Only when none are available, will an empty elevator be sent. Occasionally, there are times of congestion where no lifts are available for pickup. These requests are put into a priority queue and executed in order immediately after a simulation step has taken place.

## Building
```bash
$ docker-compose build
```

## Running
```bash
$ docker-compose up
$ docker exec -it dec bash
$ cli --elevators=2
```

## Interface
The CLI provides a handful of functions. These can be accessed by typing `help`
```
  - status
  - update [id] [goal] [direction]
  - pickup [floor] [direction]
  - step
  - help
  - exit
```

## Next steps
- It is imperative to encrypt RPC messages when on public and private networks. A straightforward solution would be to secure it over TLS with x509 certificates and keys issued to each actor.
- The interface needs to be locked down to reduce human error and bugs due to incorrect types and other common issues.
- More work could be done on the scheduling to get it even closer to modern day elevators.
