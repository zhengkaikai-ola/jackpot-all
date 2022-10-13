package gots

import (
	"jackpot-ts/app/pb"
)

type FinalResult struct {
	PlayerInfo            *pb.EntityJackpotPlayerInfo
	GlobalInfo            *pb.EntityJackpotGlobalInfo
	PlayerProgress        *pb.EntityJackpotSpinRecord
	FailReason            string
	ConsumedMoneyThisSpin int64
	WinTotalMoney         int64
}
