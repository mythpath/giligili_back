package fsm_test

import (
	"context"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/fsm"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
	"flag"
	"log"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestGroup(t *testing.T) {
	configPath := flag.String("c", "config.json", "configuration file")

	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService(*configPath)
	}))

	container.Add(&mysql.MySQLService{}, "DB", nil)
	container.Add(&fsm.RedoPersistService{}, "redo", nil)
	container.Add(&Cluster{}, "cluster", nil)
	container.Add(&StopChannel{}, "stop-channel", nil)
	container.Build()
	defer container.Dispose()

	time.Sleep(5 * time.Second)
	obj := container.GetByName("cluster")
	cluster, _ := obj.(*Cluster)
	ctx := context.Background()
	cluster.Create(ctx, cluster.group)
	stopObj := container.GetByName("stop-channel")
	stop, _ := stopObj.(*StopChannel)
	select {
	case done := <-stop.Channel:
		log.Print(done)
		return
	}

}

func TestCrashGroup(t *testing.T) {
	configPath := flag.String("c", "config.json", "configuration file")

	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService(*configPath)
	}))

	container.Add(&mysql.MySQLService{}, "DB", nil)
	container.Add(&fsm.RedoPersistService{}, "redo", nil)
	container.Add(&Cluster{}, "cluster", nil)
	container.Add(&StopChannel{}, "stop-channel", nil)
	container.Build()
	defer container.Dispose()

	time.Sleep(5 * time.Second)
	obj := container.GetByName("cluster")
	cluster, _ := obj.(*Cluster)
	ctx := context.Background()
	cluster.Create(ctx, cluster.group)
	stopObj := container.GetByName("stop-channel")
	stop, _ := stopObj.(*StopChannel)

	log.Print(unix.Getppid())

	// crash mocking point
	time.Sleep(5 * time.Microsecond)
	cmd := exec.Command("/bin/sh", "-c", "...........")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)

	select {
	case done := <-stop.Channel:
		log.Print(done)
		return
	}
}

type StopChannel struct {
	Channel chan interface{}
}

func (p *StopChannel) Init() error {

	p.Channel = make(chan interface{})
	return nil
}

type Cluster struct {
	Redo        *fsm.RedoPersistService `inject:"redo"`
	StopChannel *StopChannel            `inject:"stop-channel"`
	group       fsm.Group
}

func (p *Cluster) Init() error {
	state := map[string]fsm.StateFrame{"Init": fsm.StateFrame{0, false}, "Created": fsm.StateFrame{1, false}, "Createfailed": fsm.StateFrame{2, true}, "Startfailed": fsm.StateFrame{3, true}, "Running": fsm.StateFrame{4, true}}
	var trasitions []fsm.Transition

	createTrasition := fsm.Transition{
		From:   state["Init"],
		To:     state["Created"],
		Fail:   state["Createfailed"],
		Action: "create",
	}

	startTrasition := fsm.Transition{
		From:   state["Created"],
		To:     state["Running"],
		Fail:   state["Startfailed"],
		Action: "start",
	}

	mockStartTrasition := fsm.Transition{
		From:   state["Init"],
		To:     state["Init"],
		Fail:   state["Startfailed"],
		Action: "start",
	}

	trasitions = append(trasitions, createTrasition)
	trasitions = append(trasitions, startTrasition)
	trasitions = append(trasitions, mockStartTrasition)
	group, err := fsm.NewGroup(state, trasitions, p.Redo, 0, fsm.StateFrame{fsm.Init, false})
	p.group = group
	if err != nil {
		log.Print(err)
		return err
	}
	go p.Loop()
	return nil
}

func (p *Cluster) Loop() {
	for {

		select {
		case event := <-p.group.WatchEvents():
			switch event.From().State {

			case 0:
				//do create
				if event.Status() == fsm.Done {
					break
				}
				state, data := p.create()

				if state.State == 1 {
					event.Complete(data)
					if err := p.group.Submit(context.Background(), "start", data); err != nil {
						log.Print("start error is ", err)
					}

				} else if state.State == 2 {
					event.Fail(data)
				}

			case 1:

				if event.Status() == fsm.Done {
					break
				}
				redata := event.Data()
				state, data := p.start(redata)
				if state.State == 4 {
					event.Complete(data)
				} else if state.State == 3 {
					event.Fail(data)
				}
			case 2:
				log.Print("error create")
			case 3:
				log.Print("error start")
			case 4:
				log.Print("running now")

				p.StopChannel.Channel <- "done"
			default:
				log.Print("error states is :", event.From())
			}
		default:
		}
	}
}

func (p *Cluster) create() (fsm.StateFrame, interface{}) {
	log.Print("***create****")
	return fsm.StateFrame{1, false}, "createResultData"
}
func (p *Cluster) start(data interface{}) (fsm.StateFrame, interface{}) {
	log.Print("recive data is :", data.(string))
	//json.unmarshal(recive)
	//do
	//json.marshal(resultdata)
	log.Print("***start****")
	return fsm.StateFrame{4, true}, "startResultData"
}
func (p *Cluster) Create(ctx context.Context, group fsm.Group) {

	if err := group.Submit(ctx, "create", "createData"); err != nil {
		log.Print("create error is ,", err)
	}
}
