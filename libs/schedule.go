package Libs

import (
	"Letsgo2/Lconfig"
	"time"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"fmt"
	"errors"
)

//Scheduler 调度器接口
type Scheduler interface {
	GetSchedule() *ScheduleModel
	GetDebugInfo() *DebugModel
	InitSchedule()
}

//LSchedule 结构体
type LSchedule struct {
}

//NewSchedule 返回一个LSchedule类型指针
func NewSchedule() *LSchedule {
	return &LSchedule{}
}

//Init LSchedule初始化
func (c *LSchedule) Init() {

}

//Run 运行多协程model函数调度器
func (c *LSchedule) Run(ser Scheduler) ([]BaseReturnData, error) {
	sch := ser.GetSchedule()
	debug := ser.GetDebugInfo()
	ser.InitSchedule()

	var all_data []BaseReturnData
	var all_seqid []string
	AllRecvData := make(map[string]BaseReturnData)

	var NeedSchSum int
	for k := range sch.FuncDescs {
		seqid := c.GenUniqID()
		all_seqid = append(all_seqid, seqid)

		debug.Add(fmt.Sprintln("Schedule seqid",seqid))

		//worker run
		go c.ScheduleWorker(sch,debug,k,seqid)
		NeedSchSum++
	}

	
	for NeedSchSum > 0 {
		select {
			case i,ok := <- sch.DataCH:
				if ok {
					//fmt.Println("Schedule channel receive data:",i)
					AllRecvData[i.SEQID] = i.RET
					debug.Add(fmt.Sprintln("get func return",i.SEQID))
				} else {
					return all_data, errors.New(fmt.Sprintln("[error]Schedule channel closed before reading"))
				}
			case <-time.After(Lconfig.SCHEDULE_SELECT_TIMEOUT):
				return all_data, errors.New(fmt.Sprintln("[error]Schedule channel timeout after ",Lconfig.SCHEDULE_SELECT_TIMEOUT," second"))
		}
		NeedSchSum--
	}
	
	//排序
	for _,seqid := range all_seqid {
		all_data = append(all_data, AllRecvData[seqid])
	}
	return all_data, nil
}

//ScheduleWorker 调度器工人
func (c *LSchedule) ScheduleWorker(sch *ScheduleModel, debug *DebugModel, index int, seqid string) {
	start := time.Now()
	//运行函数
	ret := sch.FuncDescs[index].ModelFunc(sch.FuncDescs[index].Args...)
	end := time.Since(start)
	debug.Add(fmt.Sprintln("Worker Time Cost",end,"seqid",seqid))
	sch.DataCH <- ScheduleChan{SEQID:seqid, RET:ret}
}

//GenUniqID 生成唯一id
func (c *LSchedule) GenUniqID() string {
	un := time.Now().UnixNano()
	md5Ctx := md5.New()
    md5Ctx.Write([]byte(strconv.FormatInt(un,10) + strconv.Itoa(RandNum(1000))))
    cipherStr := hex.EncodeToString(md5Ctx.Sum(nil))
	return cipherStr
}
