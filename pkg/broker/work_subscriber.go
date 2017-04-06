package broker

import (
	"fmt"
)

type WorkSubscriber interface {
	Subscribe(msgBuffer <-chan WorkMsg)
}

type WorkMsg interface {
	Render() string
}

////////////////////////////////////////////////////////////
// Example stdout subscriber to illustrate how decoupled the
// subscribers actually are from the work engine, and how
// they can perform arbirary processing with the work messages
// as long as the WorkSubscriber interface is implemented
////////////////////////////////////////////////////////////
func SubscriberFactory(sub string) WorkSubscriber {
	// Totally unncessary, but cool nonetheless
	if sub == "etcd" {
		return NewEtcdWorkSubscriber()
	} else {
		return &StdoutWorkSubscriber{}
	}
}

// Example Subscriber
type StdoutWorkSubscriber struct {
	msgBuffer <-chan WorkMsg
}

func (s *StdoutWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	// Always drain the buffer if there's a message waiting.
	// Here we're just forwarding to stdout, but of course, the message
	// destination could be anything (ultimate websockets!)
	// NOTE: DON'T FORGET TO GOROUTINE THIS, OR WILL YOU CHOKE THE MAIN PROCESSOR
	s.msgBuffer = msgBuffer
	go func() {
		for {
			msg := <-msgBuffer
			fmt.Printf(msg.Render())
		}
	}()
}
