package main

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var stdout = log.New(os.Stdout, "", 0)
var delegate SimpleDelegate

// region: Step1: Implement thebasic interfaces

// SimpleDelegate handles incoming messages, see NotifyMsg
type SimpleDelegate struct {
	// SharedVariableValue keeps avalue shared accross all members of the cluster
	SharedVariableValue atomic.Value

	// Broadcasts is the object we use to broadcast a message
	Broadcasts *memberlist.TransmitLimitedQueue
}

func (delegate *SimpleDelegate) NotifyMsg(message []byte) {
	value := string(message)
	stdout.Printf("GOT MESSAGE: %s", value)
	delegate.SharedVariableValue.Store(value)
}

func (delegate *SimpleDelegate) NodeMeta(limit int) []byte {
	return []byte("")
}

func (delegate *SimpleDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return delegate.Broadcasts.GetBroadcasts(overhead, limit)
}

func (delegate *SimpleDelegate) LocalState(join bool) []byte {
	// see https://github.com/hashicorp/memberlist/issues/184
	return []byte("")
}

func (delegate *SimpleDelegate) MergeRemoteState(buf []byte, join bool) {
	stdout.Printf("MergeRemoteState?")
}

// Message implements the memberlist.Broadcast interface, to serialize and enqueue a message
type Message struct {
	Value string
}

// Message Returns a byte form of the message
func (m *Message) Message() []byte {
	return []byte(m.Value)
}

// Invalidates checks if enqueuing the current broadcast
// invalidates a previous broadcast
func (m *Message) Invalidates(b memberlist.Broadcast) bool {
	return false
}

// Finished is invoked when the message will no longer
// be broadcast, either due to invalidation or to the
// transmit limit being reached
func (m *Message) Finished() {

}

// endregion

func main() {

	// this is the list of members in the cluster
	var list *memberlist.Memberlist

	// Setup the broadcast queue
	brodcastQueue := new(memberlist.TransmitLimitedQueue)
	// This numeber is the multiplier for each retransmission: if you send a mesage to a node, this node will send themessage to 5 other node.
	// Higher the number, higher the TCP traffic, lower the time to converge to a common state
	brodcastQueue.RetransmitMult = 5
	brodcastQueue.NumNodes = func() int { return len(list.Members()) }

	// Configure the delegate
	delegate = SimpleDelegate{
		SharedVariableValue: atomic.Value{},
		Broadcasts:          brodcastQueue,
	}

	// Setup the library
	// DefaultLANConfig is a sane fedault configuration for a nodes in a LAN
	config := memberlist.DefaultLANConfig()
	config.Delegate = &delegate

	list, err := memberlist.Create(config)
	PanicIfErr(err)

	// Join another node of the cluster.
	// If this is the first node it can't join a cluster - there is no cluster
	joinServer := os.Getenv("NODE")
	if joinServer != "localhost" {
		stdout.Printf("joining node %s", joinServer)
		n, err := list.Join([]string{joinServer})
		if err != nil {
			stdout.Panicf("error joining node: %v", err)
		}
		stdout.Printf("found %v nodes", n)
	}

	// Asynchronously poll the member of the cluster.
	// No need to have this routine, it's just for debug
	go func() {
		for {
			for _, member := range list.Members() {
				stdout.Printf("Members: %s %s\n", member.Name, member.Addr)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	// Finally, start the web server
	http.HandleFunc("/", requestHandler)
	err = http.ListenAndServe(":3333", nil)
	PanicIfErr(err)

}

func requestHandler(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	valueToSet := query.Get("set")

	// if there is a query param, set thevalue; otherwise, read the value
	if valueToSet == "" {
		lastValue := GetValue()
		_, _ = writer.Write([]byte(fmt.Sprintf("last value: %s", lastValue)))
		return
	}

	stdout.Printf("setting value to %s", valueToSet)
	SetValue(valueToSet)
	_, _ = writer.Write([]byte(fmt.Sprintf("SET last value: %s", valueToSet)))

}

func SetValue(valueToSet string) {
	stdout.Printf("setting value %s", valueToSet)
	delegate.Broadcasts.QueueBroadcast(&Message{Value: valueToSet})
}

func GetValue() string {
	v := delegate.SharedVariableValue.Load()

	if v == nil {
		return ""
	}

	return v.(string)
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
