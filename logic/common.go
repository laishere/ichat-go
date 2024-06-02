package logic

import "gorm.io/gorm"

func commitOrRollback(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	} else {
		tx.Commit()
	}
}

func rollbackWhenPanic(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}
