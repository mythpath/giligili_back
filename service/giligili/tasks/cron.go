package tasks

import (
	"fmt"
	"github.com/robfig/cron"
	"reflect"
	"runtime"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"time"
)

type CronService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	RankService *RankService `inject:"RankService"`

	Cron *cron.Cron
}

func (c *CronService) Init() {
	c.CronJob()
}

// Run 运行
func (c *CronService) Run(job func() error) {
	from := time.Now().UnixNano()
	err := job()
	to := time.Now().UnixNano()
	jobName := runtime.FuncForPC(reflect.ValueOf(job).Pointer()).Name()
	if err != nil {
		fmt.Printf("%s error: %dms\n", jobName, (to-from)/int64(time.Millisecond))
	} else {
		fmt.Printf("%s success: %dms\n", jobName, (to-from)/int64(time.Millisecond))
	}
}

// CronJob 定时任务
func (c *CronService) CronJob() {
	if c.Cron == nil {
		c.Cron = cron.New()
	}

	c.Cron.AddFunc("0 0 0 * * *", func() {
		c.Run(c.RankService.RestartDailyRank)
	})

	c.Cron.Start()

	fmt.Println("CronJob starting...")
}
