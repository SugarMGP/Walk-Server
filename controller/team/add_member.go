package team

import (
	"gorm.io/gorm"
	"log"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func AddMember(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status == 0 {
		utility.ResponseError(context, "请先加入团队")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "只有队长可以添加队员")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", team.ID).Result()
	if teamSubmitted && team.Num >= 6 {
		utility.ResponseError(context, "队伍人数不能超过6")
		return
	}

	// 读取 Get 参数
	var newMember model.Person
	newMemberStuID := context.Query("stuid")
	if newMemberStuID == "" {
		utility.ResponseError(context, "请输入学号")
		return
	}
	if newMemberStuID == person.StuId {
		utility.ResponseError(context, "不能添加自己")
		return
	}

	// 查找新添加的用户
	result := global.DB.Where(&model.Person{StuId: newMemberStuID}).Take(&newMember)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "没有这个用户")
		return
	}
	if newMember.Status != 0 { // 如果在一个团队中
		utility.ResponseError(context, "该用户在其他队伍中")
		return
	}
	if person.Type == 1 && newMember.Type == 2 {
		utility.ResponseError(context, "该用户为教师，无法加入学生队伍")
		return
	}

	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 队伍成员数量加一
		if err := tx.Model(&team).Update("num", team.Num+1).Error; err != nil {
			return err
		}

		// 更新添加的人的状态
		newMember.Status = 1
		newMember.TeamId = int(team.ID)
		if err := model.TxUpdatePerson(tx, &newMember); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Println(err)
		utility.ResponseError(context, "服务异常，请重试")
		return
	}

	// 通知
	utility.SendMessage("你被"+person.Name+"添加至团队"+team.Name, nil, &newMember)

	utility.SendMessage("你添加了成员"+newMember.Name, nil, person)

	utility.ResponseSuccess(context, nil)
}
