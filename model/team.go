package model

import (
	"errors"
	"time"
	"walk-server/global"
)

type Team struct {
	ID         uint      `gorm:"comment:队伍ID"`
	Name       string    `gorm:"size:64;not null;comment:队伍名称"`
	Num        uint8     `gorm:"not null;default:1;comment:团队人数"`
	Password   string    `gorm:"size:64;not null;comment:团队加入密码"`
	Slogan     string    `gorm:"size:128;comment:团队标语"`
	AllowMatch bool      `gorm:"not null;default:false;comment:是否允许随机匹配"`
	Captain    string    `gorm:"size:64;not null;comment:队长OpenID"`
	Route      uint8     `gorm:"not null;comment:路线(1朝晖,2屏峰半程,3屏峰全程,4莫干山半程,5莫干山全程)"`
	Point      int8      `gorm:"default:0;comment:点位"`
	StartNum   uint      `gorm:"not null;default:0;comment:开始时人数"`
	Status     uint8     `gorm:"not null;default:1;comment:状态(1未开始,2进行中,3未完成,4完成,5扫码成功)"`
	Submit     bool      `gorm:"not null;default:false;comment:是否已提交报名"`
	Code       string    `gorm:"size:128;index;comment:签到二维码绑定码"`
	Time       time.Time `gorm:"comment:队伍状态更新时间"`
	IsLost     bool      `gorm:"not null;default:false;comment:是否失联"`
}

func GetTeamInfo(teamID uint) (*Team, error) {
	team := new(Team)
	result := global.DB.Where("id = ?", teamID).Take(team)

	if result.RowsAffected == 0 {
		return nil, errors.New("no team")
	}
	return team, nil
}

func GetPersonsInTeam(teamID int) (Person, []Person) {
	var persons []Person

	global.DB.Where("team_id = ?", teamID).Order("status DESC").Find(&persons)
	if len(persons) == 0 {
		return Person{}, []Person{}
	}

	// 第一个是队长（status=2），其余是队员
	captain := persons[0]
	members := persons[1:]

	return captain, members
}
