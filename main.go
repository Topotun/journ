// main.go
package main

import (
	"log"
	"time"
)

const (
	S_UPDATE_FAIL = 1
	S_OK          = 0
	upd_delay     = time.Second
	A_CHECKALL    = 1
	A_CHECKONE    = 2
)

type Message struct {
	id     int //routine id
	status int //status of routine
}

type Routine_info struct {
	id        int       //routine id
	comm_chan chan bool //communication channel
}

func Director(to_wait time.Duration, num_routines int) (chan Message, chan Message, chan Routine_info) {
	/*Controls the status of num_routines routines, makes to_wait duration pauses
	between queries of controlled routines*/
	aux := make(chan Message, num_routines)             //waits for messages from routines
	rout_slice := make([]Routine_info, 0, num_routines) //maintains a list of routines
	rout_map := make(map[int]int)                       //maintain a map of ids->slice indices of routines
	report := make(chan Routine_info, num_routines)     //waits for registration from routines
	command := make(chan Message, 1)

	go func() { //runs in parallel while we supply links to its communication channels to other routines
		for { //runs forever
			timeout := time.After(to_wait)
		loop:
			for { //waits to_wait time, meanwhile allowing registration of new go-routines
				time.Sleep(to_wait / 10)
				if len(rout_slice) >= num_routines {
					break loop
				}
				select {
				case a_new_routine := <-report: //new routine registers
					log.Println("Append", a_new_routine)
					rout_slice = append(rout_slice, a_new_routine) //add it to the list	
					rout_map[a_new_routine.id] = len(rout_slice) - 1
				case <-timeout:
					break loop //enough waiting and checking for others, time to update
				}
			} //apparently it is not a good idea to wait for it in parallel
			//log.Println(rout_slice)
			select {
			case action := <-command: //receive a command from agent and proceed with its execution
				switch action.status {
				case A_CHECKALL:
					for _, control := range rout_slice { //sends all registered go-routines a signal to update themselves
						select {
						case control.comm_chan <- true:
						default:
							aux <- Message{control.id, S_UPDATE_FAIL} //report that we were unable to send update signal
						} //non-blocking request of status update 
					} //OPEN ISSUE: can we actually run these select statements in parallel?				
				case A_CHECKONE:
					index, ok := rout_map[action.id] //obtain index of a given routine in routines slice or return false if none is recorded
					if true != ok {
						log.Println("(Director) Error, unable to find requested routine in routines map, message from agent:", action)
						break //breaks from "select" statement, no label is required
					}
					if index < 0 || index > num_routines-1 {
						log.Println("(Director) Error, requested to ping routine with invalid id, message from agent:", action)
						break //breaks from "select" statement, no label is required
					}
					control := rout_slice[index] // 
					select {
					case control.comm_chan <- true:
					default:
						aux <- Message{control.id, S_UPDATE_FAIL} //report that we were unable to send update signal
					} //non-blocking request of status update 
				default:
					log.Println("(Director) Error, unknown type of action is requested, message:", action)
				}
			default:
			}
		}
	}()
	go func() {
		for {
			message := <-aux //record the message
			log.Println("Routine", message.id, "reported status", message.status)
			//at the moment processing of the message is very simple
		}
	}()
	return command, aux, report
}

func Agent(command chan Message) {
	//supplies commands to Director
	for i := 0; i < 1000; i++ {
		command <- Message{0, A_CHECKALL}
		command <- Message{i % 10, A_CHECKONE}
		time.Sleep(time.Second)
	}

}

func Routine(id int, output chan Message, registrator chan Routine_info) {
	input := make(chan bool, 1)            //we will receive our input here
	registrator <- Routine_info{id, input} //first a routine has to register to a director
	for {
		<-input
		output <- Message{id, S_OK} //all is fine
	}
}

func main() {
	agent_channel, message_channel, registration_channel := Director(upd_delay, 10)
	for i := 0; i < 10; i++ {
		go Routine(i, message_channel, registration_channel)
	}
	go Agent(agent_channel)
	time.Sleep(time.Minute)
}
