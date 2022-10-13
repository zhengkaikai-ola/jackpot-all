package gots

import (
	"context"
	"encoding/json"
	"errors"
	"gitee.com/hasika/v8go"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/os/glog"
	client "jackpot-ts"
	"jackpot-ts/app/dao"
	"jackpot-ts/app/pb"
	"jackpot-ts/env"
	"strings"
)

type V8Processor struct {
	tsEnv    *env.ServerEnv
	runTime  *Runner
	TaskChan chan *TaskInfo
	//configString string
	currentTx   *gdb.TX
	currentTask *TaskInfo
}

func (t *V8Processor) Run(c context.Context) {
	for {
		select {
		case <-c.Done():
			return
		case t.currentTask = <-t.TaskChan:
			t.ProcessTask(t.currentTask.UID, t.currentTask.AppId)
			t.tsEnv.Iso.TryReleaseValuePtrInC(true)
		}
	}
}

func (t *V8Processor) Init() {
	t.tsEnv.Init(false, true, 9000)

	goEventObject, err := t.tsEnv.CreateEmptyObject()
	defer func() {
		if goEventObject != nil {
			goEventObject.MarkValuePtrCanReleaseInC()
		}
	}()
	if err != nil {
		panic(err)
	}
	err = goEventObject.Set("queryPlayerInfo", v8go.NewFunctionTemplate(t.tsEnv.Iso, func(c *v8go.FunctionCallbackInfo) *v8go.Value {
		uid := c.Args()[0].Integer()
		appId := c.Args()[1].Integer()
		info, infoError := t.QueryPlayerInfo(uid, appId)
		if infoError != nil {
			glog.DefaultLogger().Infof("QueryPlayerInfo Failed,Will Cancel Task")
			return nil
		}
		bs, jsonErr := json.Marshal(info)
		if jsonErr != nil {
			panic(jsonErr)
		}
		ret, v8Err := v8go.NewValue(t.tsEnv.Iso, string(bs))
		if v8Err != nil {
			panic(v8Err)
		}
		defer func() {
			if ret != nil {
				ret.MarkValuePtrCanReleaseInC()
			}
		}()
		return ret
	}).GetFunction(t.tsEnv.Ctx))
	err = goEventObject.Set("queryGlobalInfo", v8go.NewFunctionTemplate(t.tsEnv.Iso, func(c *v8go.FunctionCallbackInfo) *v8go.Value {
		uid := c.Args()[0].Integer()
		appId := c.Args()[1].Integer()
		globalInfo, infoError := t.QueryGlobalInfo(uid, appId)
		if infoError != nil {
			glog.DefaultLogger().Infof("QueryGlobalrInfo Failed,Will Cancel Task")
			return nil
		}
		bs, jsonErr := json.Marshal(globalInfo)
		if jsonErr != nil {
			panic(jsonErr)
		}
		ret, v8Err := v8go.NewValue(t.tsEnv.Iso, string(bs))
		if v8Err != nil {
			panic(v8Err)
		}
		defer func() {
			if ret != nil {
				ret.MarkValuePtrCanReleaseInC()
			}
		}()
		return ret
	}).GetFunction(t.tsEnv.Ctx))
	//err = goEventObject.Set("queryConfig", v8go.NewFunctionTemplate(t.tsEnv.Iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
	//	ret, v8Err := v8go.NewValue(t.tsEnv.Iso, t.configString)
	//	if v8Err != nil {
	//		panic(v8Err)
	//	}
	//	defer func() {
	//		if ret != nil {
	//			ret.MarkValuePtrCanReleaseInC()
	//		}
	//	}()
	//	return ret
	//}).GetFunction(t.tsEnv.Ctx))
	global := t.tsEnv.Ctx.Global()
	err = global.Set("TsCallGo", goEventObject)
	if err != nil {
		panic(err)
	}
	err = global.Set("DefaultJackpotConfigString", t.runTime.ConfigString)
	if err != nil {
		panic(err)
	}
	t.tsEnv.Iso.TryReleaseValuePtrInC(true)
}

var QueryGoDataError = errors.New("Task Failed,Query Go Data Error")

func (t *V8Processor) ProcessTask(uid uint32, appId int32) {
	err := dao.JackpotPlayerInfo.DB.Ctx(context.Background()).Transaction(context.Background(), func(ctx context.Context, tx *gdb.TX) error {
		t.currentTx = tx
		ret := t.tsEnv.GoCallTsEvent("processOneTask", uid, appId)
		defer func() {
			if ret != nil {
				ret.MarkValuePtrCanReleaseInC()
			}
		}()
		//数据错误
		s := ret.String()
		if len(strings.TrimSpace(s)) == 0 {
			glog.DefaultLogger().Infof("Task Failed,Query Go Data Error")
			return QueryGoDataError
		}
		//正常处理
		finalResult := &FinalResult{}
		unmarshalError := json.Unmarshal([]byte(s), finalResult)
		if unmarshalError != nil {
			glog.DefaultLogger().Infof("Unmarshal FinalResult Error,%+v", unmarshalError)
			return unmarshalError
		}
		//更新玩家金币信息
		playerInfo := finalResult.PlayerInfo
		deltaMoney := finalResult.WinTotalMoney - finalResult.ConsumedMoneyThisSpin
		if deltaMoney < 0 {
			rpcDecreaseRet, decreaseError := client.GetRpcConsumeIns().RpcDecreaseCount(context.TODO(), uint32(playerInfo.Uid), 1, int32(-deltaMoney), "JackpotSpin", 9)
			if decreaseError != nil {
				glog.DefaultLogger().Infof("RpcDecreaseCount Error,%+v", decreaseError)
				return decreaseError
			}
			playerInfo.Money = rpcDecreaseRet.MonetCnt
		} else if deltaMoney > 0 {
			rpcIncreaseRet, increaseError := client.GetRpcConsumeIns().RpcIncreaseCount(context.TODO(), uint32(playerInfo.Uid), 1, int32(deltaMoney), "JackpotSpin", 9)
			if increaseError != nil {
				glog.DefaultLogger().Infof("RpcIncreaseCount Error,%+v", increaseError)
				return increaseError
			}
			playerInfo.Money = rpcIncreaseRet.MonetCnt
		}
		//保存数据记录
		saveError := t.SaveSpinResult(finalResult)
		if saveError != nil {
			glog.DefaultLogger().Infof("SaveSpinResult Error,%+v", saveError)
			return saveError
		}
		//发送给客户端
		t.currentTask.ResultChan <- finalResult
		t.currentTx = nil
		t.currentTask = nil
		return nil
	})
	if err != nil {
		glog.DefaultLogger().Infof("Process Task Error")
		return
	}
}

func (t *V8Processor) QueryPlayerInfo(uid int64, appId int64) (*pb.EntityJackpotPlayerInfo, error) {
	return QueryPlayerInfo(uid, appId, t.currentTx)
}

func (t *V8Processor) QueryGlobalInfo(uid int64, appId int64) (*pb.EntityJackpotGlobalInfo, error) {
	return QueryGlobalInfo(t.currentTx, uid, appId)
}

func (t *V8Processor) SaveSpinResult(result *FinalResult) error {
	return SaveFinalResultToDb(result, t.currentTx, t.currentTask)
}

func NewV8Processor(r *Runner) *V8Processor {
	ret := &V8Processor{}
	ret.TaskChan = make(chan *TaskInfo, 100)
	ret.tsEnv = env.NewServerEnv("./jackpot-ts")
	ret.runTime = r
	ret.Init()
	return ret
}
