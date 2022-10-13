package client

import (
	"jackpot-ts/consume"
	"jackpot-ts/local"
	"sync"
)

var rpcConsumeIns consume.IConsumeClient
var rpcConsumeOnce sync.Once // 单例执行

func GetRpcConsumeIns() consume.IConsumeClient {
	rpcConsumeOnce.Do(func() {
		rpcConsumeIns = newConsumeClient()
	})
	return rpcConsumeIns
}

func newConsumeClient() consume.IConsumeClient {
	return local.NewConsumeClient()
}
