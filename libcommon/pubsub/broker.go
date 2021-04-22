package pubsub

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"sync"
	"time"
)

type Broker struct {
	Logger        *logging.LoggerService `inject:"LoggerService"`
	ModelRegistry orm.ModelRegistry      `inject:"DB"`
	DB            orm.DBService          `inject:"DB"`
	topics        map[string]topicInfo
	sync.Mutex
}

type topicInfo struct {
	option      TopicOptions
	subscribers map[string]chans
}

type chans struct {
	eventChan chan *Event
	stopChan  chan struct{}
}

func (b *Broker) Init() error {
	b.topics = make(map[string]topicInfo)
	return nil
}

func (b *Broker) AfterNew() {
	b.ModelRegistry.Put("SubTopic", subTopicDesc())
	b.ModelRegistry.Put("TopicMessage", topicMessageDesc())
}

func (b *Broker) Register(topic string, opt *TopicOptions) {
	b.Lock()
	defer b.Unlock()
	if _, ok := b.topics[topic]; !ok {
		topicOption := defaultTopicOptions()
		if opt != nil {
			replaceOpt(opt, &topicOption)
		}

		b.topics[topic] = topicInfo{
			option:      topicOption,
			subscribers: make(map[string]chans),
		}
	}
}

func (b *Broker) Publish(topic string, message interface{}) error {
	if _, ok := b.topics[topic]; !ok {
		return fmt.Errorf("topic %s not register", topic)
	}

	subscribers, err := b.getSubscriberByTopic(topic)
	if err != nil {
		return err
	}

	msgContent, err := json.Marshal(message)
	if err != nil {
		return err
	}

	tx := b.DB.GetDB().Begin()
	for _, subscriber := range subscribers {
		topicMessage := topicMessage{
			Identifier: subscriber,
			Topic:      topic,
			Message:    string(msgContent),
		}

		if err := tx.Create(&topicMessage).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (b *Broker) getSubscriberByTopic(topic string) ([]string, error) {
	subTopics := []subTopic{}
	err := b.DB.GetDB().Select("identifier").Where("topic = ?", topic).Find(&subTopics).Error
	if err != nil {
		return nil, err
	}

	subscribers := make([]string, 0, len(subTopics))
	for _, subtopic := range subTopics {
		subscribers = append(subscribers, subtopic.Identifier)
	}

	return subscribers, nil
}

func (b *Broker) Subscribe(topic string, identifier string) (<-chan *Event, error) {
	b.Lock()
	defer b.Unlock()

	topicInfo, ok := b.topics[topic]
	if !ok {
		return nil, fmt.Errorf("topic %s not register", topic)
	}

	sub := subTopic{
		Identifier: identifier,
		Topic:      topic,
	}
	if err := b.DB.GetDB().Where("identifier = ? and topic = ?", identifier, topic).Find(&subTopic{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		if err := b.DB.GetDB().Create(&sub).Error; err != nil {
			return nil, err
		}
	}

	subscriberChans, ok := topicInfo.subscribers[identifier]
	if !ok {
		subscriberChans = chans{
			eventChan: make(chan *Event, b.topics[topic].option.getSizeOfChan()),
			stopChan:  make(chan struct{}),
		}
		topicInfo.subscribers[identifier] = subscriberChans
	} else {
		return nil, fmt.Errorf("have subscribe topic %s", topic)
	}

	go func(channel chan<- *Event, stop chan struct{}) {
		delayTime := b.topics[topic].option.getDelayTime()
		ticker := time.NewTicker(delayTime)
		for {
			select {
			case <-ticker.C:
				events, err := b.getMsgByTopic(topic, identifier)
				if err != nil {
					b.Logger.Errorf("get topic message fail! reason:%s", err.Error())
				}
				for _, event := range events {
					channel <- event
				}

			case <-stop:
				return
			}
		}

	}(subscriberChans.eventChan, subscriberChans.stopChan)

	return subscriberChans.eventChan, nil
}

func (b *Broker) getMsgByTopic(topic string, identifier string) ([]*Event, error) {
	msgs := []topicMessage{}
	if err := b.DB.GetDB().Where("topic = ?", topic).
		Where("identifier = ?", identifier).Order("created_at asc").Find(&msgs).Error; err != nil {
		return nil, err
	}

	events := make([]*Event, 0, len(msgs))

	for _, msg := range msgs {
		events = append(events, &Event{
			broker:  b,
			id:      msg.ID,
			topic:   msg.Topic,
			message: msg.Message,
		})
	}

	return events, nil
}

func (b *Broker) Unsubscribe(topic string, identifier string) error {
	b.Lock()
	defer b.Unlock()

	topicInfo, ok := b.topics[topic]
	if !ok {
		return nil
	}

	subscriberChan, ok := topicInfo.subscribers[identifier]
	if !ok {
		return nil
	}

	if err := b.deleteSubscriTopicRecord(topic, identifier); err != nil {
		return err
	}
	select {
	case subscriberChan.stopChan <- struct{}{}:

	}
	close(subscriberChan.eventChan)
	delete(topicInfo.subscribers, identifier)

	return nil
}

func (b *Broker) deleteSubscriTopicRecord(topic string, identifier string) error {
	subTopic := subTopic{}
	if err := b.DB.GetDB().Where("topic = ?", topic).
		Where("identifier = ?", identifier).Find(&subTopic).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	if err := b.DB.GetDB().Where("topic = ?", topic).
		Where("identifier = ?", identifier).Unscoped().Delete(subTopic).Error; err != nil {
		return err
	}
	return nil
}

func (b *Broker) deleteTopicmsg(id uint) error {
	topicMsg := topicMessage{}
	if err := b.DB.GetDB().Find(&topicMsg, "id = ?", id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	if err := b.DB.GetDB().Unscoped().Delete(&topicMsg).Error; err != nil {
		return err
	}
	return nil
}
