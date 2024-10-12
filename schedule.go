package letsgo

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/time2k/letsgo-ng/config"
)

// Scheduler 调度器接口
type Scheduler interface {
	GetBuilder() *ScheduleBuilder
	GetDebugInfo() *DebugInfo
	InitSchedule()
}

// Schedule 结构体
type Schedule struct {
}

// newSchedule 返回一个Schedule类型指针
func newSchedule() *Schedule {
	return &Schedule{}
}

// Init Schedule初始化
func (c *Schedule) Init() {

}

// Run 运行多协程model函数调度器
func (c *Schedule) Run(ser Scheduler) ([]BaseReturnData, error) {
	sch := ser.GetBuilder()
	debug := ser.GetDebugInfo()
	ser.InitSchedule()

	var AllData []BaseReturnData
	AllRecvData := make(map[string]BaseReturnData)

	var NeedSchSum int
	for k := range sch.FuncDescs {
		seqid := c.GenUniqID()
		sch.FuncDescs[k].SEQID = seqid

		debug.Add(fmt.Sprintln("Schedule seqid", seqid))

		//worker run
		go c.ScheduleWorker(sch, debug, k, seqid)
		NeedSchSum++
	}

	for NeedSchSum > 0 {
		select {
		case i, ok := <-sch.DataCH:
			if ok {
				//fmt.Println("Schedule channel receive data:",i)
				AllRecvData[i.SEQID] = i.RET
				debug.Add(fmt.Sprintln("get func return", i.SEQID))
			} else {
				return AllData, errors.New(fmt.Sprintln("[error]Schedule channel closed before reading"))
			}
		case <-time.After(config.SCHEDULE_SELECT_TIMEOUT):
			return AllData, errors.New(fmt.Sprintln("[error]Schedule channel timeout after ", config.SCHEDULE_SELECT_TIMEOUT, " second"))
		}
		NeedSchSum--
	}

	//排序
	for _, eachsch := range sch.FuncDescs {
		AllData = append(AllData, AllRecvData[eachsch.SEQID])
	}
	return AllData, nil
}

// ScheduleWorker 调度器工人
func (c *Schedule) ScheduleWorker(sch *ScheduleBuilder, debug *DebugInfo, index int, seqid string) {
	defer PanicFunc()
	start := time.Now()
	//运行函数
	ret := sch.FuncDescs[index].ModelFunc(sch.FuncDescs[index].CommP, sch.FuncDescs[index].Args...)
	end := time.Since(start)
	debug.Add(fmt.Sprintln("Worker Time Cost", end, "seqid", seqid, "args", sch.FuncDescs[index].Args))
	sch.DataCH <- ScheduleChan{SEQID: seqid, RET: ret}
}

// GenUniqID 生成唯一id
func (c *Schedule) GenUniqID() string {
	un := time.Now().UnixNano()
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(strconv.FormatInt(un, 10) + strconv.Itoa(RandNum(1000))))
	cipherStr := hex.EncodeToString(md5Ctx.Sum(nil))
	return cipherStr
}
