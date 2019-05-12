package main

import (
	"fmt"
	"time"
	"encoding/json"
	"log"
	"github.com/beanstalkd/go-beanstalk"
)


// beanstalk config struct
type BeanstalkConfig struct {
	Uri			string	`json:"uri"`
	Tube			string	`json:"tube"`
	ReplyTubePrefix		string	`json:"reply_tube_prefix"`
	ReconnectTimeout	int	`json:"reconnect_timeout"`
	ReserveTimeout		int	`json:"reserve_timeout"`
	PublishTimeout		int	`json:"publish_timeout"`
}

func beanstalkdPublish(config BeanstalkConfig, tube string, body []byte) error {

	amqpURI := config.Uri
	c, err := beanstalk.Dial("tcp", amqpURI)

	if err != nil {
		log.Printf("Publish/callback: unable connect to beanstalkd broker:%s", err)
		return nil
	}

	mytube := &beanstalk.Tube{Conn: c, Name: tube}
	id, err := mytube.Put([]byte(body), 1, 0, time.Duration(config.PublishTimeout)*time.Second)
	if err != nil {
		fmt.Printf("\nPublish err: %d\n",err)
		return err
	}
	fmt.Printf("\nPublish id: %d\n",id)

	return nil
}

func beanstalkdLoop(config BeanstalkConfig) error {
	for {
		beanstalkdConsume(config)
		log.Printf("broker disconnected, sleep and retry:%d\n", config.ReconnectTimeout)
		time.Sleep(time.Duration(config.ReconnectTimeout) * time.Second)
	}
	return nil
}

func WakeOnJob(ch chan bool, config BeanstalkConfig, id uint64, body []byte) {

	fmt.Printf("\nwake up and delete job id: %d\n",id)
	comment := Comment{}
	fmt.Printf("\nWI: %d\n",id)
	comment.JobID = id
	response := fmt.Sprintf("%v", comment.JobID)
	fmt.Printf("response %s\n", response)
	//callback
	log.Printf("recv msg: %s", string(body))
	err := json.Unmarshal(body, &comment)
	if err != nil {
			log.Printf("json decode error %s", err)
	}
	callbackQueueName := fmt.Sprintf("%s%d",config.ReplyTubePrefix,comment.JobID)
	fmt.Printf("callback queue name: %s\n",callbackQueueName)
	err, cbsdTask := DoProcess(&comment)
	if err != nil {
		fmt.Println("doprocess error:", err)
		panic(err)
	}
	b, err := json.Marshal(cbsdTask)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("FINE: %s\n",b)
	err = beanstalkdPublish(config,callbackQueueName,b)
	ch <- true
}


func beanstalkdConsume(config BeanstalkConfig) error {

	amqpURI := config.Uri
	tube := config.Tube

	c, err := beanstalk.Dial("tcp", amqpURI)

	if err != nil {
		log.Printf("Unable connect to beanstalkd broker:%s", err)
		return nil
	}

	log.Printf("Subscribe tube: %s, reserve timeout: %d", tube, config.ReserveTimeout)

	c.TubeSet = *beanstalk.NewTubeSet(c, tube)

	for {
		// The BS library does not understand the network/BS problems and hangs forever.
		// ping in backround?
		id, body, err := c.Reserve(time.Duration(config.ReserveTimeout) * time.Second)

		if err != nil {
			fmt.Printf("\nid: %d, res: %s\n",id, err.Error())
		}

		if id == 0 {
			continue
		}
		c.Delete(id)

		ch := make(chan bool)
		go WakeOnJob(ch, config, id, body)
	}

	return nil
}
