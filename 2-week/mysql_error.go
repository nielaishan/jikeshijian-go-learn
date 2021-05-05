package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UserInfoDO struct {
	ID           int64     `gorm:"column:id" json:"id"`
	UserId   string    `gorm:"column:user_id" json:"userId"`
	UserName string    `gorm:"column:user_name" json:"userName"`
}
func (UserInfoDO) TableName() string {
	return "user"
}

func main() {
	fmt.Println(MysqlError())
}

func MysqlError() ([]UserInfoDO, error) {
	db,_ := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	var userIDs []int64
	result := make([]UserInfoDO, 0)
	err := db.Where("user_id in {?}", userIDs).Find(&result).Error
	if err == gorm.ErrRecordNotFound {
		return result, nil
	} else if err != nil {
		// log.warn

	}
	return  result, errors.Wrap(err, "mysql fail")
}