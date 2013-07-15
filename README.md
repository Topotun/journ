Two main functions, Director and Routine

"Director" runs in parallel two routines.
One routine has two phases: registration and sending confirmation requests, which run one after another in an infinite loop.
  Registration allows new routines to send their input channels to director. Confirmation requests are then sent to registered routines through given channels.
Another routine accepts messages from registered routines and immediately prints them (this behaviour can be altered of course, e.g. by adding a parameter function)
"Director" returns reference to registration and message channel, which are respectively input and output for its go-routines.

"Routine" runs an infinite loop that sends a confirmation message upon receiving a request from director.

This architecture can be improved with altering "Director" so that it only requests updates from routines only by demand from "Agent", which
simulates actions of a human observer. Moreover, a "Transmitter" can be added, which will forward channel messages via network.

