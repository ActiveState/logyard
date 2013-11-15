package main

import (
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"logyard/clients/common"
	"logyard/clients/docker_events"
	"logyard/clients/sieve"
	"stackato/server"
	"time"
)

var NodeID string

func SendToLogyard(pub *zmqpubsub.Publisher, event *docker_events.Event) {
	text := fmt.Sprintf("%v action for container %v (image: %v)",
		event.Status, event.Id, event.From)
	(&sieve.Event{
		Type:          event.Status,
		Process:       "docker_events",
		Severity:      "INFO",
		Desc:          text,
		MessageCommon: common.NewMessageCommon(text, time.Unix(event.Time, 0), NodeID),
	}).MustPublish(pub)
}

func main() {
	log.Info("Starting docker_events")
	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	for event := range docker_events.Stream() {
		SendToLogyard(pub, event)
	}
}

func init() {
	var err error
	NodeID, err = server.LocalIP()
	if err != nil {
		log.Fatalf("Failed to determine IP addr: %v", err)
	}
}