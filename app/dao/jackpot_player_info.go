// ============================================================================
// This is auto-generated by gf cli tool only once. Fill this file as you wish.
// ============================================================================

package dao

import (
	"jackpot-ts/app/dao/internal"
)

// jackpotPlayerInfoDao is the manager for logic model data accessing
// and custom defined data operations functions management. You can define
// methods on it to extend its functionality as you wish.
type jackpotPlayerInfoDao struct {
	internal.JackpotPlayerInfoDao
}

var (
	// JackpotPlayerInfo is globally public accessible object for table jackpot_player_info operations.
	JackpotPlayerInfo = jackpotPlayerInfoDao{
		internal.JackpotPlayerInfo,
	}
)

// Fill with you ideas below.