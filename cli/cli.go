package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"

	. "dec/client"

	"github.com/chzyer/readline"
	proto "github.com/gogo/protobuf/proto"
)

var flagBind = flag.String("bind", "127.0.0.1:8999", "Bind to address")
var flagElevators = flag.Int("elevators", 16, "Amount of elevators to connect to")
var client *Client

// Reference imports to suppress errors if they are not otherwise used
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Function constructor - constructs new function for listing given directory
var completer = readline.NewPrefixCompleter(
	readline.PcItem("status"),
	readline.PcItem("update"),
	readline.PcItem("pickup"),
	readline.PcItem("step"),
	readline.PcItem("help"),
	readline.PcItem("exit"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// Block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func main() {
	flag.Parse()

	logo := `
        __
   ____/ /__  _____
  / __  / _ \/ ___/
 / /_/ /  __/ /__
 \__,_/\___/\___/
`
	log.Println(logo)

	// setup
	client = NewClient(*flagBind, *flagElevators)
	client.SendStatusRequest(StatusRequestOpt{BroadcastAll: true})

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})

	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		switch {
		case line == "status":
			statusRequestOpt := StatusRequestOpt{BroadcastAll: true}

			client.SendStatusRequest(statusRequestOpt)
			client.PrintCurrentStatus()
		case strings.HasPrefix(line, "update "):
			parts := strings.SplitN(line, " ", 4)

			if len(parts) != 4 {
				fmt.Printf("Wrong number of arguments for `update`. expected: ID Goal Direction\n")
			} else {
				i, err := strconv.Atoi(parts[1])
				if err != nil {
					panic(err)
				}
				id := int(i)

				g, err := strconv.Atoi(parts[2])
				if err != nil {
					panic(err)
				}
				goal := uint32(g)

				d, err := strconv.Atoi(parts[3])
				if err != nil {
					panic(err)
				}
				direction := int32(d)

				client.SendUpdateRequest(id, goal, direction)
				client.PrintCurrentStatus()
			}
		case strings.HasPrefix(line, "pickup "):
			parts := strings.SplitN(line, " ", 3)

			if len(parts) != 3 {
				fmt.Printf("Wrong number of arguments for `pickup`. expected: Floor Direction\n")
			} else {
				f, err := strconv.Atoi(parts[1])
				if err != nil {
					panic(err)
				}
				floor := uint32(f)

				d, err := strconv.Atoi(parts[2])
				if err != nil {
					panic(err)
				}
				direction := int32(d)

				client.SendPickupRequest(floor, direction)
				client.PrintCurrentStatus()
			}
		case line == "help":
			helpText := `
 commands:
  - status
  - update [id] [goal] [direction]
  - pickup [floor] [direction]
  - step
  - help
  - exit
`
			fmt.Println(helpText)
		case line == "step":
			client.SendStepRequest()
			client.PrintCurrentStatus()
		case line == "exit":
			goto exit
		case line == "":
		default:
			log.Println("Invalid command :", strconv.Quote(line))
		}
	}
exit:
}
