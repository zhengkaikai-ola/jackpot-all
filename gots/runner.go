package gots

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/os/glog"
	client "jackpot-ts"
	"jackpot-ts/app/dao"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TaskInfo struct {
	UID        uint32
	AppId      int32
	ResultChan chan *FinalResult
}

type Runner struct {
	Processors   map[int]*V8Processor
	cancel       context.CancelFunc
	ConfigString string
}

func NewTaskPool(processorNum int) *Runner {
	ret := &Runner{}
	ret.Init(processorNum)
	return ret
}

func (r *Runner) Init(processorNum int) {
	configBs, readError := os.ReadFile("./config.json")
	if readError != nil {
		panic(readError)
	}
	r.ConfigString = string(configBs)
	r.Processors = map[int]*V8Processor{}
	ctx, cancel := context.WithCancel(context.TODO())
	r.cancel = cancel
	for index := 0; index < processorNum; index++ {
		p := NewV8Processor(r)
		r.Processors[index] = p
		go p.Run(ctx)
	}

}

func (r *Runner) Stop() {
	r.cancel()
}

func (r *Runner) AddTask(uid uint32, appid int32) chan *FinalResult {
	pn := uint32(len(r.Processors))
	index := uid % pn
	processor := r.Processors[int(index)]
	ch := make(chan *FinalResult, 1)
	processor.TaskChan <- &TaskInfo{
		UID:        uid,
		AppId:      appid,
		ResultChan: ch,
	}
	return ch
}

type ErrorInfo struct {
	Error string
}

func (r *Runner) Spin(ctx *gin.Context) {
	uid, err := QueryPathArg(ctx, "uid")
	if err != nil {
		return
	}
	appId, err := QueryPathArg(ctx, "appid")
	if err != nil {
		return
	}
	retChan := r.AddTask(uint32(uid), int32(appId))
	after := time.After(time.Second * 3)
	select {
	case <-after:
		ctx.IndentedJSON(http.StatusRequestTimeout, &ErrorInfo{Error: "Time Out"})
	case ret := <-retChan:
		if ret == nil {
			ctx.IndentedJSON(http.StatusInternalServerError, &ErrorInfo{Error: "Internal Server Error"})
		} else {
			if len(ret.FailReason) == 0 {
				ctx.IndentedJSON(http.StatusOK, ret)
			} else {
				ctx.IndentedJSON(http.StatusInternalServerError, &ErrorInfo{Error: ret.FailReason})
			}
		}
	}
}

func QueryPathArg(ctx *gin.Context, argName string) (int64, error) {
	uidString := ctx.Query(argName)
	uid, err := strconv.ParseInt(uidString, 10, 64)
	if err != nil {
		glog.DefaultLogger().Infof("Parse Arg %s Error,%+v", argName, err)
		ctx.IndentedJSON(http.StatusBadRequest, &ErrorInfo{Error: err.Error()})
		return 0, err
	}
	return uid, nil
}

func (r *Runner) ChangeSpinLevel(ctx *gin.Context) {
	uid, err := QueryPathArg(ctx, "uid")
	if err != nil {
		return
	}
	appId, err := QueryPathArg(ctx, "appid")
	if err != nil {
		return
	}
	level, err := QueryPathArg(ctx, "level")
	if err != nil {
		return
	}
	//只有一个
	transActionError := dao.JackpotPlayerInfo.DB.Ctx(context.Background()).Transaction(context.Background(), func(ctx context.Context, tx *gdb.TX) error {
		return ChangeSpinLevel(tx, uid, appId, level)
	})
	if transActionError != nil {
		glog.DefaultLogger().Infof("Change Spin Level Failed ,%+v", transActionError)
	}
}

func (r *Runner) QueryConfig(c *gin.Context) {
	c.String(http.StatusOK, r.ConfigString)
}

func (r *Runner) QueryBroadCast(c *gin.Context) {
	time, err := QueryPathArg(c, "time")
	if err != nil {
		return
	}
	appId, err := QueryPathArg(c, "appid")
	if err != nil {
		return
	}
	cnt, err := QueryPathArg(c, "cnt")
	if err != nil {
		return
	}
	m := dao.JackpotSpinRecord.Ctx(context.Background()).Where(dao.JackpotSpinRecord.Columns.Appid, appId)
	if time > 0 {
		m = m.And(dao.JackpotSpinRecord.Columns.SpinTime+">=?", time)
	}
	if cnt <= 0 {
		cnt = 1
	}
	if cnt > 20 {
		cnt = 20
	}
	m = m.Order(dao.JackpotSpinRecord.Columns.SpinTime + " desc")
	records, queryError := m.All()
	if queryError != nil {
		return
	}
	c.IndentedJSON(200, records)
}

func (r *Runner) ExchangeMoney(c *gin.Context) {
	uid, err := QueryPathArg(c, "uid")
	if err != nil {
		return
	}
	appId, err := QueryPathArg(c, "appid")
	if err != nil {
		return
	}
	//moneyType, err := QueryPathArg(c, "moneyType")
	//if err != nil {
	//	return
	//}
	moneyCnt, err := QueryPathArg(c, "moneyCnt")
	if err != nil {
		return
	}
	transActionError := dao.JackpotPlayerInfo.Ctx(context.Background()).Transaction(context.Background(), func(ctx context.Context, tx *gdb.TX) error {
		pl, playerError := EnsurePlayerExist(tx, uid, appId)
		if playerError != nil {
			glog.DefaultLogger().Infof("EnsurePlayerExist Error,%v", playerError)
			return playerError
		}
		pl.Money += moneyCnt
		_, updateError := dao.JackpotPlayerInfo.Update(pl)
		if updateError != nil {
			glog.DefaultLogger().Infof("Update Player Money Error,%v", updateError)
			return updateError
		}
		return nil
	})
	if transActionError != nil {
		glog.DefaultLogger().Infof("Update Player Money TransAction Error,%v", transActionError)
		c.JSON(http.StatusInternalServerError, &ErrorInfo{
			Error: "InternalError",
		})
	}
}

func (r *Runner) AddMoney(c *gin.Context) {
	uid, err := QueryPathArg(c, "uid")
	if err != nil {
		return
	}
	appId, err := QueryPathArg(c, "appid")
	if err != nil {
		return
	}
	glog.DefaultLogger().Infof("Add Money,User %d ,AppId %d, Cnt 10000", uid, appId)
	ret, err := client.GetRpcConsumeIns().RpcIncreaseCount(context.TODO(), uint32(uid), 1, 1000000000, "模拟充值", 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ErrorInfo{
			Error: "InternalError",
		})
		return
	}
	c.IndentedJSON(http.StatusOK, ret)
}

func (r *Runner) QueryMoney(c *gin.Context) {
	uid, err := QueryPathArg(c, "uid")
	if err != nil {
		return
	}

	ret, err := client.GetRpcConsumeIns().RpcQueryMoney(context.TODO(), uint32(uid), 1)
	if err != nil {
		if err != nil {
			c.JSON(http.StatusInternalServerError, &ErrorInfo{
				Error: "InternalError",
			})
			return
		}
	}
	c.JSON(http.StatusOK, ret.MonetCnt)
}
