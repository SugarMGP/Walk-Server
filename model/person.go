package model

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
	"walk-server/global"
)

type Person struct {
	OpenId     string `gorm:"primaryKey;size:64;not null;comment:微信OpenID"` // openID
	Name       string `gorm:"size:128;not null;comment:姓名"`
	Gender     int8   `gorm:"not null;comment:性别(1男,2女)"`
	StuId      string `gorm:"size:32;unique;comment:学号"`
	Campus     uint8  `gorm:"not null;comment:校区(1朝晖,2屏峰,3莫干山)"`
	Identity   string `gorm:"size:18;unique;not null;comment:身份证号"`
	Status     uint8  `gorm:"not null;default:0;comment:队伍状态(0未加入,1队员,2队长)"`
	Qq         string `gorm:"size:20;comment:QQ号"`
	Wechat     string `gorm:"size:64;comment:微信号"`
	College    string `gorm:"size:64;not null;comment:学院"`
	Tel        string `gorm:"size:20;unique;not null;comment:联系电话"`
	CreatedOp  uint8  `gorm:"not null;default:3;comment:创建团队次数"`
	JoinOp     uint8  `gorm:"not null;default:5;comment:加入团队次数"`
	TeamId     int    `gorm:"index;default:-1;comment:所属团队ID"`
	Type       uint8  `gorm:"not null;comment:人员类型(1学生,2教职工,3校友)"`
	WalkStatus uint8  `gorm:"not null;default:1;comment:活动状态(1未开始,2进行中,3扫码成功,4放弃,5完成)"`
}

func (p *Person) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

func (p *Person) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// GetPerson 使用加密后的 open ID 获取 person 数据
// encOpenID 是加密后的 openID
// 如果没有找到这个用户就返回 error
func GetPerson(encOpenID string) (*Person, error) {
	// 如果缓存中找到了这个数据 直接返回缓存数据
	var person Person
	if err := global.Rdb.Get(global.Rctx, encOpenID).Scan(&person); err == nil {
		return &person, nil
	}

	// 如果缓存中没有就进数据库查询用户数据
	result := global.DB.Where(&Person{OpenId: encOpenID}).Take(&person)
	if result.RowsAffected == 0 {
		return nil, errors.New("no person")
	} else {
		global.Rdb.Set(global.Rctx, encOpenID, &person, 20*time.Minute)
		return &person, nil
	}
}

// UpdatePerson 更新 person 数据
// encOpenID 加密后的用户 openID
// person 用户数据 (完整的)
func UpdatePerson(encOpenID string, person *Person) {
	// 如果缓存中存在这个数据, 先更新缓存
	if _, err := global.Rdb.Get(global.Rctx, encOpenID).Result(); err == nil {
		global.Rdb.Set(global.Rctx, encOpenID, person, 20*time.Minute)
	}

	// 更新数据库中的数据
	global.DB.Where(&Person{OpenId: encOpenID}).Save(person)
}

func SetPerson(encOpenID string, person *Person) error {
	// 如果缓存中存在这个数据, 先更新缓存
	if _, err := global.Rdb.Get(global.Rctx, encOpenID).Result(); err == nil {
		global.Rdb.Set(global.Rctx, encOpenID, person, 20*time.Minute)
	}

	// 更新数据库中的数据
	return global.DB.Exec("UPDATE people SET open_id = ? WHERE open_id = ?", encOpenID, person.OpenId).Error
}

// TxUpdatePerson 事务中更新
func TxUpdatePerson(tx *gorm.DB, person *Person) error {
	// 如果缓存中存在这个数据, 先更新缓存
	if _, err := global.Rdb.Get(global.Rctx, person.OpenId).Result(); err == nil {
		global.Rdb.Set(global.Rctx, person.OpenId, person, 20*time.Minute)
	} else if !errors.Is(err, redis.Nil) {
		return err
	}

	// 更新数据库中的数据
	if err := tx.Where(&Person{OpenId: person.OpenId}).Save(person).Error; err != nil {
		return err
	}

	return nil
}
