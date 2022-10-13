package consume

import (
	"context"
)

type RpcDecreaseCount struct {
	Success  bool
	Msg      string
	MonetCnt int64
}

type RpcQueryResult struct {
	Success  bool
	Msg      string
	MonetCnt int64
}

type RpcInCreaseResult struct {
	Success  bool
	Msg      string
	MonetCnt int64
}

type IConsumeClient interface {
	GetOrderId(args ...int64) (string, error)

	RpcQueryMoney(ctx context.Context, uid uint32, moneyType int32) (*RpcQueryResult, error)

	RpcDecreaseCount(ctx context.Context, uid uint32, moneyType int32, moneyCnt int32, detailString string, optionDetail int64, extraData ...int64) (*RpcDecreaseCount, error)

	RpcIncreaseCount(ctx context.Context, uid uint32, moneyType int32, moneyCnt int32, detailString string, optionDetail int64, extraData ...int64) (*RpcInCreaseResult, error)
}
