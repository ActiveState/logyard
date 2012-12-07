package main

import (
	"encoding/json"
	"fmt"
	"github.com/srid/log"
	"logyard"
	"logyard/stackato/events"
	"time"
)

// Make relevant cloud events available in application logs. Heroku style.
func MonitorCloudEvents() {
	// TODO: add more events; will require modifying the log
	// invokation to include the required app id
	filters := []string{
		"event.dea_start",
		"event.dea_ready",
		"event.dea_stop",
		"event.stager_start",
		"event.stager_end",
		"event.cc_app_update",
	}

	c, err := logyard.NewClientGlobal()
	if err != nil {
		log.Fatal(err)
	}
	ss, err := c.Recv(filters)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Listening for app relevant cloud events...")
	for msg := range ss.Ch {
		var event events.Event
		err := json.Unmarshal([]byte(msg.Value), &event)
		if err != nil {
			log.Fatal(err) // not expected at all
		}

		switch msg.Key {
		case "event.dea_start", "event.dea_ready", "event.dea_stop":
			appid := int(event.Info["app_id"].(float64))
			name := event.Info["app_name"].(string)
			index := int(event.Info["instance"].(float64))
			source := "stackato.dea"
			PublishAppLog(c, appid, name, index, source, &event)
		case "event.stager_start", "event.stager_end":
			appid := int(event.Info["app_id"].(float64))
			name := event.Info["app_name"].(string)
			PublishAppLog(c, appid, name, -1, "stackato.stager", &event)
		case "event.cc_app_update":
			appid := int(event.Info["app_id"].(float64))
			name := event.Info["app_name"].(string)
			PublishAppLog(c, appid, name, -1, "stackato.controller", &event)
		}
	}
	log.Warn("Finished listening for app relevant cloud events.")

	err = ss.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func PublishAppLog(
	client *logyard.Client, app_id int, app_name string,
	index int, source string, event *events.Event) {

	m := AppLogMessage{
		Text:          event.Desc,
		LogFilename:   "",
		UnixTime:      event.UnixTime,
		HumanTime:     time.Unix(event.UnixTime, 0).Format("2006-01-02T15:04:05-07:00"), // heroku-format
		Source:        source,
		InstanceIndex: index,
		AppID:         app_id,
		AppName:       app_name,
	}
	data, err := json.Marshal(m)
	if err != nil {
		log.Errorf("cannot encode %+v into JSON; %s. Skipping this message", m, err)
		return
	}
	key := fmt.Sprintf("apptail.%d", app_id)
	err = client.Send(key, string(data))
	if err != nil {
		log.Fatal(err)
	}
}
