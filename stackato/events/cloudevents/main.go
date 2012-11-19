package main

import (
	"encoding/json"
	"logyard"
	"logyard/log2"
	"logyard/stackato/events"
)

// TODO: share it with systail
type SystailRecord struct {
	UnixTime int64
	Text     string
	Name     string
	NodeID   string
}

func main() {
	parser := events.NewStackatoParser()
	parser.DeleteSamples()

	logyardclient, err := logyard.NewClientGlobal()
	if err != nil {
		log2.Fatal(err)
	}

	sub, err := logyardclient.Recv([]string{"systail"})
	if err != nil {
		log2.Fatal(err)
	}
	log2.Info("Watching the systail stream on this node")
	for message := range sub.Ch {
		var record SystailRecord
		err := json.Unmarshal([]byte(message.Value), &record)
		if err != nil {
			log2.Errorf("failed to parse json: %s; ignoring record: %s",
				err, message.Value)
			continue
		}

		event, err := parser.Parse(record.Name, record.Text)
		if err != nil {
			log2.Errorf(
				"failed to parse event from %s: %s -- source: %s", record.Name, err, record.Text)
			continue
		}
		if event != nil {
			event.NodeID = record.NodeID
			event.UnixTime = record.UnixTime
			data, err := json.Marshal(event)
			if err != nil {
				log2.Fatal(err)
			}
			logyardclient.Send("event."+event.Type, string(data))
		}

	}
}
