/*
Copyrigtt 2017 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    tttp://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/satori/go.uuid"
	"github.com/urfave/cli"

	"github.com/fission/fission"
	"github.com/fission/fission/mqtrigger/messageQueue"
)

func mqtCreate(c *cli.Context) error {
	client := getClient(c.GlobalString("server"))

	mqtName := c.String("name")
	if len(mqtName) == 0 {
		mqtName = uuid.NewV4().String()
	}
	fnName := c.String("function")
	if len(fnName) == 0 {
		fatal("Need a function name to create a trigger, use --function")
	}
	fnUid := c.String("uid")
	mqType := c.String("mqtype")
	switch mqType {
	case "":
		mqType = messageQueue.NATS
	case messageQueue.NATS:
		mqType = messageQueue.NATS
	default:
		fatal("Unknown message queue type, currently only \"nats-streaming\" is supported")
	}

	// TODO: check topic availability
	topic := c.String("topic")
	if len(topic) == 0 {
		fatal("Listen topic cannot be empty")
	}
	respTopic := c.String("resptopic")

	if topic == respTopic {
		fatal("Listen topic should not equal to response topic")
	}

	checkMQTopicAvailability(mqType, topic, respTopic)

	fnMeta := fission.Metadata{
		Name: fnName,
		Uid:  fnUid,
	}

	mqt := fission.MessageQueueTrigger{
		Metadata: fission.Metadata{
			Name: mqtName,
		},
		Function:         fnMeta,
		MessageQueueType: mqType,
		Topic:            topic,
		ResponseTopic:    respTopic,
	}

	_, err := client.MessageQueueTriggerCreate(&mqt)
	checkErr(err, "create message queue trigger")

	fmt.Printf("trigger '%s' created\n", mqtName)
	return err
}

func mqtGet(c *cli.Context) error {
	return nil
}

func mqtUpdate(c *cli.Context) error {
	client := getClient(c.GlobalString("server"))
	mqtName := c.String("name")
	if len(mqtName) == 0 {
		fatal("Need name of trigger, use --name")
	}
	topic := c.String("topic")
	respTopic := c.String("resptopic")

	mqt, err := client.MessageQueueTriggerGet(&fission.Metadata{Name: mqtName})
	checkErr(err, "get Time trigger")

	checkMQTopicAvailability(mqt.MessageQueueType, topic, respTopic)

	mqt.Topic = topic
	mqt.ResponseTopic = respTopic

	_, err = client.MessageQueueTriggerUpdate(mqt)
	checkErr(err, "update Time trigger")

	fmt.Printf("trigger '%v' updated\n", mqtName)
	return nil
}

func mqtDelete(c *cli.Context) error {
	client := getClient(c.GlobalString("server"))
	mqtName := c.String("name")
	if len(mqtName) == 0 {
		fatal("Need name of trigger to delete, use --name")
	}

	err := client.MessageQueueTriggerDelete(&fission.Metadata{Name: mqtName})
	checkErr(err, "delete trigger")

	fmt.Printf("trigger '%v' deleted\n", mqtName)
	return nil
}

func mqtList(c *cli.Context) error {
	client := getClient(c.GlobalString("server"))

	mqts, err := client.MessageQueueTriggerList(c.String("mqtype"))
	checkErr(err, "list message queue triggers")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
		"NAME", "FUNCTION_NAME", "FUNCTION_UID", "MESSAGE_QUEUE_TYPE", "TOPIC", "RESPONSE_TOPIC")
	for _, mqt := range mqts {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
			mqt.Metadata.Name, mqt.Function.Name, mqt.Function.Uid, mqt.MessageQueueType, mqt.Topic, mqt.ResponseTopic)
	}
	w.Flush()

	return nil
}

func checkMQTopicAvailability(mqType string, topics ...string) {
	for _, t := range topics {
		if len(t) > 0 && !messageQueue.IsTopicValid(mqType, t) {
			fatal(fmt.Sprintf("Invalid topic for %s: %s", mqType, t))
		}
	}
}
