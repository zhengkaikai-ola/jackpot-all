package gots

import (
	"context"
	"errors"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/os/glog"
	client "jackpot-ts"
	"jackpot-ts/app/dao"
	"jackpot-ts/app/pb"
	"time"
)

func SaveFinalResultToDb(result *FinalResult, tx *gdb.TX, task *TaskInfo) error {
	if len(result.FailReason) != 0 {
		return nil
	}
	playerInfo := result.PlayerInfo
	//保存全局池子信息
	_, updateGlobalError := dao.JackpotGlobalInfo.TX(tx).
		Where(dao.JackpotGlobalInfo.Columns.PoolID, result.GlobalInfo.PoolId).
		Update(map[string]interface{}{
			dao.JackpotGlobalInfo.Columns.Money: result.GlobalInfo.Money,
		})
	if updateGlobalError != nil {
		glog.DefaultLogger().Infof("Update Global Info Error %+v", updateGlobalError)
		return updateGlobalError
	}
	//保存玩家信息

	_, updatePlayerError := dao.JackpotPlayerInfo.TX(tx).
		Where(dao.JackpotPlayerInfo.Columns.UID, task.UID).
		Where(dao.JackpotPlayerInfo.Columns.Appid, task.AppId).
		Update(map[string]interface{}{
			dao.JackpotPlayerInfo.Columns.TodaySpinCnt:  playerInfo.TodaySpinCnt,
			dao.JackpotPlayerInfo.Columns.Money:         playerInfo.Money,
			dao.JackpotPlayerInfo.Columns.RecordLevel:   playerInfo.RecordLevel,
			dao.JackpotPlayerInfo.Columns.LastSpinTime:  playerInfo.LastSpinTime,
			dao.JackpotPlayerInfo.Columns.TodaySpinCnt:  playerInfo.TodaySpinCnt,
			dao.JackpotPlayerInfo.Columns.BonusProgress: playerInfo.BonusProgress,
			dao.JackpotPlayerInfo.Columns.BonusCnt:      playerInfo.BonusCnt,
		})
	if updatePlayerError != nil {
		glog.DefaultLogger().Infof("Update Player Info Error %+v", updatePlayerError)
		return updatePlayerError
	}
	//保存玩家记录信息
	record := result.PlayerProgress
	record.Uid = playerInfo.Uid
	record.Appid = playerInfo.Appid
	record.ResultOrder = int32(0)
	record.SpinTime = playerInfo.LastSpinTime
	_, insertError := dao.JackpotSpinRecord.TX(tx).Insert(record)
	if insertError != nil {
		glog.DefaultLogger().Infof("Insert Player Spin Record Error %+v", insertError)
		return insertError
	}
	return nil
}

func NeedResetTodaySpinCnt(spinTime int64) bool {
	if spinTime == 0 {
		return true
	}
	nowYear, nowMon, nowDay := time.Now().Date()
	lastYear, lastMon, lastDay := time.Unix(0, spinTime*1000_000).Date()
	if nowYear > lastYear {
		return true
	}
	if nowYear == lastYear && nowMon > lastMon {
		return true
	}
	if nowYear == lastYear && nowMon == lastMon && nowDay > lastDay {
		return true
	}
	return false
}

func QueryPlayerInfo(uid int64, appId int64, tx *gdb.TX) (*pb.EntityJackpotPlayerInfo, error) {
	modelInfo, err := EnsurePlayerExist(tx, uid, appId)
	if err != nil {
		glog.DefaultLogger().Infof("Ensure Player Exist Error %+v", err)
		return nil, err
	}
	if NeedResetTodaySpinCnt(modelInfo.LastSpinTime) {
		modelInfo.TodaySpinCnt = 0
	}

	return modelInfo, nil
}

func QueryGlobalInfo(tx *gdb.TX, uid int64, id int64) (*pb.EntityJackpotGlobalInfo, error) {
	poolId := 1
	findResults, findError := dao.JackpotGlobalInfo.TX(tx).LockUpdate().Where(dao.JackpotGlobalInfo.Columns.PoolID, poolId).One()
	if findError != nil {
		return nil, findError
	}
	var globalInfo *pb.EntityJackpotGlobalInfo = nil
	if findResults == nil {
		globalInfo = &pb.EntityJackpotGlobalInfo{
			PoolId: int32(poolId),
			Money:  0,
		}
		_, insertError := dao.JackpotGlobalInfo.TX(tx).Insert(globalInfo)
		if insertError != nil {
			glog.DefaultLogger().Infof("Insert Global Info Error %+v", insertError)
			return nil, insertError
		}
	} else {
		globalInfo = findResults
	}
	return globalInfo, nil
}

func EnsurePlayerExist(tx *gdb.TX, uid int64, appId int64) (*pb.EntityJackpotPlayerInfo, error) {
	findResults, findError := dao.JackpotPlayerInfo.TX(tx).
		Where(dao.JackpotPlayerInfo.Columns.UID, uid).
		Where(dao.JackpotPlayerInfo.Columns.Appid, appId).
		One()
	if findError != nil {
		//查找失败无法处理
		glog.DefaultLogger().Infof("Find Player Info Error ,%+v", findError)
		return nil, findError
	}
	queryMoneyRet, rpcQueryError := client.GetRpcConsumeIns().RpcQueryMoney(context.TODO(), uint32(uid), 1)
	if rpcQueryError != nil {
		return nil, rpcQueryError
	}
	if !queryMoneyRet.Success {
		return nil, errors.New(queryMoneyRet.Msg)
	}
	var player *pb.EntityJackpotPlayerInfo
	if findResults == nil {
		//查找没有错误,但是没能找到记录,表明需要插入记录
		player = &pb.EntityJackpotPlayerInfo{
			Uid:           uid,
			Money:         queryMoneyRet.MonetCnt,
			RecordLevel:   1,
			LastSpinTime:  0,
			TodaySpinCnt:  0,
			BonusProgress: 0,
			BonusCnt:      0,
			Appid:         int32(appId),
		}
		_, insertError := dao.JackpotPlayerInfo.TX(tx).Insert(player)
		if insertError != nil {
			glog.DefaultLogger().Infof("Insert Player Info Error ,%+v", insertError)
			return nil, insertError
		}
	} else {
		player = findResults
	}
	player.Money = queryMoneyRet.MonetCnt
	return player, nil
}

func ChangeSpinLevel(tx *gdb.TX, uid int64, appId int64, level int64) error {
	_, playerError := EnsurePlayerExist(tx, uid, appId)
	if playerError != nil {
		return playerError
	}
	_, updateError := dao.JackpotPlayerInfo.TX(tx).
		Where(dao.JackpotPlayerInfo.Columns.UID, uid).
		Where(dao.JackpotPlayerInfo.Columns.Appid, appId).
		Update(map[string]interface{}{
			dao.JackpotPlayerInfo.Columns.RecordLevel: level,
		})
	if updateError != nil {
		glog.DefaultLogger().Infof("Update Record Level Error ,%+v", updateError)
	}
	return updateError
}
