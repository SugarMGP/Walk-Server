package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
	"walk-server/constant"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"
)

type CreateTestTeamData struct {
	Secret string `json:"secret" binding:"required"`
	Num    int    `json:"num" binding:"required"` // 队伍数量
}

func CreateTestTeams(c *gin.Context) {
	var postForm CreateTestTeamData
	if err := c.ShouldBind(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 生成队伍数据（不插入数据库）
		var teams []model.Team
		for i := 0; i < postForm.Num; i++ {
			route := rand.Intn(5) + 1
			teams = append(teams, model.Team{
				Name:       "测试队伍" + strconv.Itoa(i),
				Num:        4,
				Password:   "test",
				Slogan:     "123",
				AllowMatch: false,
				Captain:    "", // 先留空，后面填充
				Route:      uint8(route),
				Point:      0,
				Status:     1,
				StartNum:   0,
				Submit:     false,
				Time:       time.Now(),
			})
		}

		// 2. 批量插入 Team
		if err := tx.Create(&teams).Error; err != nil {
			return err // 回滚事务
		}

		// 3. 生成人员数据（先存到切片）
		var persons []model.Person
		teamCaptains := make(map[uint]string) // team_id -> Captain OpenId

		for i, team := range teams {
			campus := rand.Intn(3) + 1
			for j := 0; j < 4; j++ {
				person := model.Person{
					OpenId:     "test" + strconv.Itoa(i) + "team" + strconv.Itoa(j),
					Name:       "测试队伍" + strconv.Itoa(i) + "队员" + strconv.Itoa(j),
					Gender:     1,
					StuId:      fmt.Sprintf("%011d", rand.Int63n(1e11)),
					Campus:     uint8(campus),
					Identity:   "test" + fmt.Sprintf("%011d", rand.Int63n(1e11)),
					Status:     uint8(ternary(j == 0, 2, 1)), // 第一个人是队长
					Qq:         "123",
					Wechat:     "123",
					College:    "计算机学院",
					Tel:        fmt.Sprintf("%011d", rand.Int63n(1e11)),
					CreatedOp:  1,
					JoinOp:     1,
					TeamId:     int(team.ID), // 赋值 Team ID
					Type:       1,
					WalkStatus: 1,
				}
				persons = append(persons, person)

				// 如果是 Status == 2（队长），存入 map
				if person.Status == 2 {
					teamCaptains[team.ID] = person.OpenId
				}
			}
		}

		// 4. 批量插入 Person
		if err := tx.Create(&persons).Error; err != nil {
			return err // 回滚事务
		}

		// 5. 更新队长信息
		for teamID, captainOpenID := range teamCaptains {
			if err := tx.Model(&model.Team{}).Where("id = ?", teamID).Update("captain", captainOpenID).Error; err != nil {
				return err // 回滚事务
			}
		}

		return nil // 提交事务
	})

	if err != nil {
		utility.ResponseError(c, "创建队伍失败："+err.Error())
		return
	}

	utility.ResponseSuccess(c, nil)
}

type DeleteTestTeamData struct {
	Secret string `json:"secret" binding:"required"`
}

func DeleteTestTeams(c *gin.Context) {
	var postForm DeleteTestTeamData
	if err := c.ShouldBind(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	// 开启事务
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 删除 Team
		if err := tx.Where("name LIKE ?", "测试队伍%").Delete(&model.Team{}).Error; err != nil {
			return err // 事务回滚
		}

		// 2. 删除 Person
		if err := tx.Where("open_id LIKE ?", "test%").Delete(&model.Person{}).Error; err != nil {
			return err // 事务回滚
		}

		return nil // 事务提交
	})

	if err != nil {
		utility.ResponseError(c, "删除失败："+err.Error())
		return
	}

	utility.ResponseSuccess(c, nil)
}

type UpdateTestTeamData struct {
	Secret string `json:"secret" binding:"required"`
}

func UpdateTestTeams(c *gin.Context) {
	var postForm UpdateTestTeamData
	if err := c.ShouldBind(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 更新所有测试人员的 WalkStatus 为 2
		if err := tx.Model(&model.Person{}).
			Where("open_id LIKE ?", "test%").
			Update("walk_status", 2).Error; err != nil {
			return err // 发生错误，回滚事务
		}

		// 2. 查询所有测试队伍
		var teams []model.Team
		if err := tx.Where("name LIKE ?", "测试队伍%").Find(&teams).Error; err != nil {
			return err
		}

		// 3. 更新测试队伍状态和随机 Point
		for _, team := range teams {
			team.Status = 2
			team.Time = time.Now().Add(time.Duration(-1*rand.Intn(60)) * time.Minute).Add(time.Duration(-1*rand.Intn(24)) * time.Hour)
			team.Submit = true
			team.Point = int8(rand.Intn(int(constant.PointMap[team.Route])))

			if err := tx.Save(&team).Error; err != nil {
				return err // 发生错误，回滚事务
			}
		}

		return nil // 事务成功提交
	})

	if err != nil {
		utility.ResponseError(c, "更新失败："+err.Error())
		return
	}

	utility.ResponseSuccess(c, nil)
}

func ternary(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}
