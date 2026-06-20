package db

import "gorm.io/gorm"

const (
	advisoryClassCheer int32 = 1
)

func AdvisoryLockCheer(tx *gorm.DB, accountID uint) error {
	return tx.Exec("SELECT pg_advisory_xact_lock(?, ?)", advisoryClassCheer, int32(accountID)).Error
}

func AdvisoryLockRelationship(tx *gorm.DB, a, b uint) error {
	lo, hi := a, b
	if lo > hi {
		lo, hi = hi, lo
	}
	key := int64(uint32(lo))<<32 | int64(uint32(hi))
	return tx.Exec("SELECT pg_advisory_xact_lock(?)", key).Error
}
