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
)

type Message struct {
	id     int //routine id
	status int //status of routine
}

type Rout_inf struct {
	id   int
	comm chan bool
}

func Director(to_wait time.Duration, num_routines int) (chan Message, chan Rout_inf) {
	/*Controls the status num_routines routines, makes to_wait duration pauses
	between queries of controlled routines*/
	aux := make(chan Message, num_routines)      //waits for messages from routines
	rout_slice := make([]Rout_inf, num_routines) //maintains a list of routines
	report := make(chan Rout_inf, num_routines)  //waits for registration from routines
	go func() {
		for { //runs forever
		l:
			for { //waits to_wait time, meanwhile allowing registration of new go-routines
				timeout := time.After(to_wait)
				select {
				case a_new_routine := <-report: //new routine registers
					log.Println("Append", a_new_routine)
					rout_slice = append(rout_slice, a_new_routine) //add it to the list	
				case <-timeout:
					break l //enough waiting and checking for others, time to update
				}
				time.Sleep(to_wait / 10)
			} //apparently it is not a good idea to wait for it in parallel
			//log.Println(rout_slice)
			for _, control := range rout_slice { //sends all registered go-routines a signal to update themselves
				log.Println(control)
				select {
				case control.comm <- true:
				default:
					aux <- Message{control.id, S_UPDATE_FAIL} //report that we were unable to send update signal
				} //non-blocking request of status update sent to "control" channel
			} //OPEN ISSUE: can we actually run these select statements in parallel?
		}
	}() //no deleting yetticker
	go func() {
		for {
			message := <-aux //record the message
			log.Println("Routine", message.id, "reported status", message.status)
			//at the moment processing of the message is very simple
		}
	}()
	return aux, report
}

func Routine(id int, output chan Message, registrator chan Rout_inf) {
	input := make(chan bool, 1)        //we will receive our input here
	registrator <- Rout_inf{id, input} //first a routine has to register to a director
	for {
		<-input
		output <- Message{id, S_OK} //all is fine
	}
}

func main() {
	message_channel, registration_channel := Director(upd_delay, 10)
	for i := 0; i < 10; i++ {
		go Routine(i, message_channel, registration_channel)
	}
	time.Sleep(time.Minute)
}
