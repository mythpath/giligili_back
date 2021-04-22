package test

import (
	"flag"
	"fmt"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
	"selfText/giligili_back/libcommon/pubsub"
	"testing"
	"time"
)

const (
	Topic      = "Test-topic"
	Concurrent = true
)

type TopicMeeage struct {
	Id      int
	Message string
}

func TestPubsub(t *testing.T) {
	configPath := flag.String("c", "./config.json", "configuration file")

	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService(*configPath)
	}))

	container.Add(&mysql.MySQLService{}, "DB", nil)
	container.Add(&logging.LoggerService{}, "LoggerService", nil)
	container.Add(&pubsub.Broker{}, "BrokerService", nil)
	container.Add(&Publisher{}, "Publisher", nil)
	container.Add(&Subscriber{}, "Subscriber", nil)

	container.Build()
	time.Sleep(65 * time.Second)
	defer container.Dispose()
}

type Publisher struct {
	Broker *pubsub.Broker `inject:"BrokerService"`
}

func (p *Publisher) Init() error {
	fmt.Printf("Publisher init...\n")
	p.Broker.Register(Topic, &pubsub.TopicOptions{DelayTime: time.Second})
	go p.PulishMsg()
	return nil
}

func (p *Publisher) PulishMsg() {
	num := 0
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := p.Broker.Publish(Topic, TopicMeeage{Id: num, Message: Topic}); err != nil {
				fmt.Printf("[error]: pushlish message fail! reason:%s\n", err.Error())
			}

			fmt.Println("publish msg:", num)
			num++
		}

		if num >= 10 {
			fmt.Println("publish out ...")
			return
		}
	}
}

type Subscriber struct {
	Broker    *pubsub.Broker `inject:"BrokerService"`
	Publisher *Publisher     `inject:"Publisher"`
}

func (s *Subscriber) Init() error {
	identfier := pubsub.Uniqueidentifier("test", Topic)
	if !Concurrent {
		eventChan, err := s.Broker.Subscribe(Topic, identfier)
		if err != nil {
			fmt.Printf("[error]: subscribe topic : %s fail!\n", Topic)
			return err
		}
		go s.DealMsg(eventChan)
		time.Sleep(1 * time.Minute)

		if err := s.Broker.Unsubscribe(Topic, identfier); err != nil {
			fmt.Printf("[error]: unsubscribe topic : %s fail! reason:%s\n", Topic, err.Error())
			return err
		}
	} else {
		for i := 0; i < 3; i++ {
			go func() {

				eventChan, err := s.Broker.Subscribe(Topic, identfier)
				if err != nil {
					fmt.Printf("[error]: subscribe topic : %s fail! reason:%s\n", Topic, err.Error())
				} else {
					go s.DealMsg(eventChan)
				}
			}()
		}

		time.Sleep(1 * time.Minute)
		for i := 0; i < 3; i++ {
			go func() {
				if err := s.Broker.Unsubscribe(Topic, identfier); err != nil {
					fmt.Printf("[error]: unsubscribe topic : %s fail! reason:%s\n", Topic, err.Error())
				}
			}()
		}

	}

	return nil
}

func (s *Subscriber) DealMsg(eventChan <-chan *pubsub.Event) {
	topicMsg := TopicMeeage{}
	for event := range eventChan {
		if err := event.GetEventContent(&topicMsg); err != nil {
			fmt.Printf("[error]: get topic message fail! reason:%s\n", err.Error())
		}

		fmt.Println("receive:", topicMsg)
		if err := event.Ack(); err != nil {
			fmt.Println("ack event fail! reason:", err.Error())
		}
	}

	fmt.Println("deal message out...")
}
