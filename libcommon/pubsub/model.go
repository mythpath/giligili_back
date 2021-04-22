package pubsub

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
)

type subTopic struct {
	orm.SelfGormModel

	Identifier string `gorm:"unique_index:identifier_topic;not null"`
	Topic      string `gorm:"unique_index:identifier_topic;not null"`
}

func subTopicDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &subTopic{},
		New: func() interface{} {
			return &subTopic{}
		},
		NewSlice: func() interface{} {
			return &[]subTopic{}
		},
	}
}

type topicMessage struct {
	orm.SelfGormModel
	Identifier string
	Topic      string
	Message    string `gorm:"type:text"`
}

func topicMessageDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &topicMessage{},
		New: func() interface{} {
			return &topicMessage{}
		},
		NewSlice: func() interface{} {
			return &[]topicMessage{}
		},
	}
}
