package idpool

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type mysqlIDRepo struct {
	db *gorm.DB
}

func NewMysqlIDRepo(db *gorm.DB) IDPoolRepo {
	return &mysqlIDRepo{
		db: db,
	}
}

func (r *mysqlIDRepo) GetIDPool(takeAmount uint64) (idPool *IDPool, err error) {
	var tmp IDPool
	err = r.db.First(&tmp, "(end_pos - ?) >= current_pos", takeAmount).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = NoIDAvailable{}
			return
		}
		return
	}

	idPool = &tmp

	return
}

func (r *mysqlIDRepo) CreateNewIDPoolClaim(idPool *IDPool, takeAmount uint64) (res *IDPoolClaim, err error) {
	var newIDPoolClaim IDPoolClaim
	err = r.db.Transaction(func(tx *gorm.DB) error {
		uRes := tx.Model(IDPool{}).
			Where("id = ?", idPool.ID).
			Where("(end_pos - ?) >= current_pos", takeAmount).
			UpdateColumn("current_pos", gorm.Expr("current_pos + ?", takeAmount))
		if uRes.Error != nil {
			return uRes.Error
		}
		if uRes.RowsAffected == 0 {
			return NoIDAvailable{}
		}

		var tmpIDPool IDPool
		sErr := tx.First(&tmpIDPool, "id = ?", idPool.ID).Error
		if sErr != nil {
			return sErr
		}

		newIDPoolClaim = IDPoolClaim{
			StartPos:  tmpIDPool.CurrentPos - takeAmount,
			EndPos:    tmpIDPool.CurrentPos,
			ClaimedAt: uint64(time.Now().Unix()),
		}

		return tx.Create(&newIDPoolClaim).Error
	})
	if err != nil {
		return
	}
	res = &newIDPoolClaim

	return
}
