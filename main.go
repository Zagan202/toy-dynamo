// main.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson          lelawson
// Pete Wilcox              pcwilcox
// Annie Shen				ashen7
// Victoria Tran            vilatran
//
// This is the main source file for HW2. It sets up some initialization variables by
// reading the environment, then sets up the two interfaces the application uses. If
// the app is launched as a 'leader', then it will use a kvs object from kvs.go for
// its back end data store. If it is launched as a 'follower' then it will use a
// forwarder object from forward.go as its back end data store. Whichever data store
// is used is passed as an initialization member to an App object from app.go. The
// App object implements the RESTful API front end and communicates with its data
// store in order to satisfy client requests.
//

package main

import (
	"io"
	"log"
	"os"
)

// Versioning info defined via linker flags at compile time
var branch string // Git branch
var hash string   // Shortened commit hash
var build string  // Number of commits in the branch

// MultiLogOutput controls logging output to stdout and to a log file
var MultiLogOutput io.Writer

func main() {
	// Create a logfile
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	// Create a stream that writes to console and the logfile
	MultiLogOutput = io.MultiWriter(os.Stdout, logFile)
	// Set some logging flags and setup the logger
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetOutput(MultiLogOutput)

	// Print version info to the log
	version := branch + "." + hash + "." + build
	log.Println("Running version " + version)

	// IP_PORT is defined at runtime in the docker command
	myIP = os.Getenv("IP_PORT")

	log.Println("My IP is " + myIP)

	// VIEW is defined at runtime in the docker command as a string
	str := os.Getenv("VIEW")
	log.Println("My view is: " + str)

	// docker run -p 8082:8080 --net=mynet --ip=10.0.0.2 -e VIEW="10.0.0.2:8080,10.0.0.3:8080,10.0.0.4:8080" -e IP_PORT="10.0.0.2:8080" -e S="3" REPLICA_1
	s := os.Getenv("S")
	log.Println("I belong in shard#" + s)

	// Create a viewList and load the view into it
	MyView := NewView(myIP, str)

	// Create a shardList and create the seperation of shard ID to servers
	// MyShard := NewShard(myIP, MyView)

	// Make a KVS to use as the db
	k := NewKVS()

	// The App object is the front end and has references to the KVS and viewList
	a := App{db: k, view: *MyView}

	log.Println("Starting server...")

	// The gossip object controls communicating with other servers and has references to the viewlist and the kvs
	gossip := GossipVals{
		view: MyView,
		kvs:  k,
	}
	// Start the heartbeat loop
	go gossip.GossipHeartbeat() // goroutines

	// Start the servers with references to the REST app and the gossip module
	server(a, gossip)
}
