package local

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/database/gdb"
	"jackpot-ts/app/dao"
	"jackpot-ts/app/pb"
	"jackpot-ts/consume"
	"strings"
)

type ConsumeClient struct {
}

func (c *ConsumeClient) EnsureMoneyInfoExist(tx *gdb.TX, uid uint32, moneyType int32) (*pb.EntityMockMoney, error) {
	moneyInfo, err := dao.MockMoney.TX(tx).LockUpdate().Where(dao.MockMoney.Columns.UID, uid).Where(dao.MockMoney.Columns.MoneyType, moneyType).One()
	if err != nil {
		return nil, err
	}
	if moneyInfo == nil {
		moneyInfo = &pb.EntityMockMoney{
			Uid:       int64(uid),
			MoneyType: moneyType,
			MoneyCnt:  0,
		}
		_, insertError := dao.MockMoney.TX(tx).Insert(moneyInfo)
		if insertError != nil {
			return nil, insertError
		}
	}
	return moneyInfo, nil
}

func (c *ConsumeClient) GetOrderId(args ...int64) (string, error) {
	builder := strings.Builder{}
	builder.WriteString("laya")
	for index := 0; index < len(args); index++ {
		builder.WriteString("_%d")
	}
	return fmt.Sprintf(builder.String(), args), nil
}

func (c *ConsumeClient) RpcQueryMoney(ctx context.Context, uid uint32, moneyType int32) (*consume.RpcQueryResult, error) {
	var outMoneyInfo *pb.EntityMockMoney
	txError := dao.MockMoney.Transaction(context.Background(), func(ctx context.Context, tx *gdb.TX) error {
		moneyInfo, infoError := c.EnsureMoneyInfoExist(tx, uid, moneyType)
		if infoError != nil {
			return infoError
		}
		outMoneyInfo = moneyInfo
		return nil
	})
	if outMoneyInfo != nil && txError == nil {
		return &consume.RpcQueryResult{
			Success:  true,
			Msg:      "",
			MonetCnt: outMoneyInfo.MoneyCnt,
		}, nil
	}
	return nil, txError
}

func (c *ConsumeClient) RpcDecreaseCount(ctx context.Context, uid uint32, moneyType int32, moneyCnt int32, detailString string, optionDetail int64, extraData ...int64) (*consume.RpcDecreaseCount, error) {
	var outMoneyInfo *pb.EntityMockMoney
	txError := dao.MockMoney.Transaction(context.Background(), func(ctx context.Context, tx *gdb.TX) error {
		moneyInfo, infoError := c.EnsureMoneyInfoExist(tx, uid, moneyType)
		if infoError != nil {
			return infoError
		}
		if moneyInfo.MoneyCnt < int64(moneyCnt) {
			return errors.New("No Enough Money")
		}
		moneyInfo.MoneyCnt -= int64(moneyCnt)
		_, updateError := dao.MockMoney.TX(tx).Where(dao.MockMoney.Columns.UID, uid).Update(map[string]interface{}{
			dao.MockMoney.Columns.MoneyType: moneyType,
			dao.MockMoney.Columns.MoneyCnt:  moneyInfo.MoneyCnt,
		})
		if updateError != nil {
			return updateError
		}
		outMoneyInfo = moneyInfo
		return nil
	})
	if outMoneyInfo != nil && txError == nil {
		return &consume.RpcDecreaseCount{
			Success:  true,
			Msg:      "",
			MonetCnt: outMoneyInfo.MoneyCnt,
		}, nil
	}
	return nil, txError
}

func (c *ConsumeClient) RpcIncreaseCount(ctx context.Context, uid uint32, moneyType int32, moneyCnt int32, detailString string, optionDetail int64, extraData ...int64) (*consume.RpcInCreaseResult, error) {
	var outMoneyInfo *pb.EntityMockMoney
	txError := dao.MockMoney.Transaction(context.Background(), func(ctx context.Context, tx *gdb.TX) error {
		moneyInfo, infoError := c.EnsureMoneyInfoExist(tx, uid, moneyType)
		if infoError != nil {
			return infoError
		}
		moneyInfo.MoneyCnt += int64(moneyCnt)
		_, updateError := dao.MockMoney.TX(tx).Where(dao.MockMoney.Columns.UID, uid).Update(map[string]interface{}{
			dao.MockMoney.Columns.MoneyType: moneyType,
			dao.MockMoney.Columns.MoneyCnt:  moneyInfo.MoneyCnt,
		})
		if updateError != nil {
			return updateError
		}
		outMoneyInfo = moneyInfo
		return nil
	})
	if outMoneyInfo != nil && txError == nil {
		return &consume.RpcInCreaseResult{
			Success:  true,
			Msg:      "",
			MonetCnt: outMoneyInfo.MoneyCnt,
		}, nil
	}
	return nil, txError
}

func NewConsumeClient() *ConsumeClient {
	return &ConsumeClient{}
}
