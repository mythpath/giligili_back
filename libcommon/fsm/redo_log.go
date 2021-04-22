package fsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"selfText/giligili_back/libcommon/orm"

	"github.com/jinzhu/gorm"
)

// RedoLog track the transition in the persitence system
type RedoLogger interface {
	// Redo returns all logs hasn't been applied
	Redo() (chan Event, error)
	//Commit the log in the tail,returns log id
	Commit(from, to, fail StateFrame, Data interface{}, action string, status Status, groupID uint) (Event, error)
	//FSMPersist persist a fsm for giligili
	FSMPersist(obj string, objID uint) (*FSMPersisteModel, error)
	//DestroyFSM delete fsm for giligili when it done
	DestroyFSM(groupID uint) error
	//Apply delete a event which is done
	Apply(eventID uint) error
	//RecoverFSM recover the fsm not delete to fsm manager
	RecoverFSM() ([]*FSMPersisteModel, error)

	//RecoverEventForGroup return redolog when event not finish for group
	RecoverEventForGroup(group Group) *ReDoLogModel
}

// RedoPersistService  persist event for fsm
type RedoPersistService struct {
	ModelRegistry orm.ModelRegistry `inject:"DB"`
	DB            orm.DBService     `inject:"DB"`
	redoLogChan   chan Event
}

func (p *RedoPersistService) AfterNew() {
	p.ModelRegistry.Put("ReDoLogModel", ReDoLogModelDesc())
	p.ModelRegistry.Put("FSMPersisteModel", FSMPersisteModelDesc())

}

func (p *RedoPersistService) Init() error {
	p.redoLogChan = make(chan Event, 10)

	return nil
}

func (p *RedoPersistService) Redo() (chan Event, error) {

	if p.redoLogChan == nil {
		return nil, fmt.Errorf("error redo log chan")
	}
	return p.redoLogChan, nil
}

func (p *RedoPersistService) Commit(from, to, fail StateFrame, data interface{}, action string, status Status, groupID uint) (Event, error) {

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	presentState := &event{
		from:   from,
		to:     to,
		fail:   fail,
		data:   data,
		action: action,
		status: status,
	}

	redomodel := &ReDoLogModel{
		From:               presentState.from.State,
		FromFlag:           presentState.from.Final,
		To:                 presentState.to.State,
		ToFlag:             presentState.to.Final,
		Fail:               presentState.fail.State,
		FailFlag:           presentState.fail.Final,
		Action:             presentState.action,
		Data:               string(dataBytes),
		FSMPersisteModelID: groupID,
		Status:             status,
	}
	if err := p.DB.GetDB().Create(redomodel).Error; err != nil {
		return nil, fmt.Errorf("error save log")
	}
	presentState.id = redomodel.ID

	if err := p.DB.GetDB().Model(&FSMPersisteModel{}).Where("id = ?", groupID).Updates(map[string]interface{}{"state": from.State, "state_flag": from.Final}).Error; err != nil {
		return nil, fmt.Errorf("error save fsm", err)
	}

	return presentState, nil
}

func (p *RedoPersistService) Apply(eventID uint) error {

	if err := p.DB.GetDB().Where("id = ?", eventID).Delete(&ReDoLogModel{}); err != nil {
		return fmt.Errorf("delete event failed: ", err)
	}

	return nil
}

func (p *RedoPersistService) FSMPersist(obj string, objID uint) (*FSMPersisteModel, error) {
	fsmPersistModel := &FSMPersisteModel{
		Obj:   obj,
		ObjID: objID,
	}
	if err := p.DB.GetDB().Create(fsmPersistModel).Error; err != nil {
		return nil, err
	}
	return fsmPersistModel, nil
}

func (p *RedoPersistService) RecoverFSM() ([]*FSMPersisteModel, error) {
	fsmPersistes := []*FSMPersisteModel{}
	if errors.Is(p.DB.GetDB().Find(&fsmPersistes).Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return fsmPersistes, nil
}

func (p *RedoPersistService) RecoverEventForGroup(group Group) *ReDoLogModel {
	eventPersist := &ReDoLogModel{}
	if err := p.DB.GetDB().Where("fsm_persiste_model_id = ? and alias = ?", group.GetID(), Doing).Find(eventPersist).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	p.DB.GetDB().Delete(eventPersist)
	p.DB.GetDB().Where("fsm_persiste_model_id = ? and alias = ?", group.GetID(), Done).Delete(&ReDoLogModel{})

	return eventPersist

}

func (p *RedoPersistService) DestroyFSM(groupID uint) error {
	if err := p.DB.GetDB().Where("id = ?", groupID).Delete(&FSMPersisteModel{}).Error; err != nil {
		return fmt.Errorf("destroy fsm failed: ", err)
	}
	return nil
}

type ReDoLogModel struct {
	orm.SelfGormModel

	From     State `json:"from"`
	FromFlag bool  `json:"fromFlag"`

	To     State `json:"to"`
	ToFlag bool  `json:"fromFlag"`

	Fail     State `json:"fail"`
	FailFlag bool  `json:"fromFlag"`

	Action             string `json:"action"`
	Data               string `json:"data" gorm:"type:text"`
	Status             Status `json:"alias"`
	FSMPersisteModelID uint   `json:"fsmPersisteModelID" gorm:"index`
}

func ReDoLogModelDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &ReDoLogModel{},
		New: func() interface{} {
			return &ReDoLogModel{}
		},
		NewSlice: func() interface{} {
			return &[]ReDoLogModel{}
		},
	}
}

type FSMPersisteModel struct {
	orm.SelfGormModel
	Obj          string          `json:"obj"`
	ObjID        uint            `json:"objID"`
	State        State           `json:"state"`
	StateFlag    bool            `json:"stateFlag"`
	RedoLogModel []*ReDoLogModel `json:"redoLogModel"`
}

func FSMPersisteModelDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &FSMPersisteModel{},
		New: func() interface{} {
			return &FSMPersisteModel{}
		},
		NewSlice: func() interface{} {
			return &[]FSMPersisteModel{}
		},
	}
}
