package model

import (
	"apiserver-gin/pkg/time"

	validator "gopkg.in/go-playground/validator.v9"
)

type User struct {
	BaseModel
	Name     string        `gorm:"column:name;type:varchar(32);not null" json:"name" validate:"min=1,max=32"`
	Password string        `gorm:"column:password;type:char(64);not null" json:"-" validate:"min=6,max=32"` // 密码json化时要忽略避免泄露，用不到时sql中不要查询该字段
	Birthday time.JsonTime `gorm:"column:birthday;type:datetime" json:"birthday"`
}

func (User) TableName() string {
	return "user"
}

func (user *User) Validate() error {
	validate := validator.New()
	return validate.Struct(user)
}
