// ==========================================================================
// This is auto-generated by gf cli tool. DO NOT EDIT THIS FILE MANUALLY.
// ==========================================================================

package internal

// MockMoney is the golang structure for table mock_money.
type MockMoney struct {
	UID       int64 `orm:"uid,primary"        json:"uid"`        //
	MoneyType int32 `orm:"money_type,primary" json:"money_type"` //
	MoneyCnt  int64 `orm:"money_cnt"          json:"money_cnt"`  //
}
