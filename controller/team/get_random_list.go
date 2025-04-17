package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type GetRandomListData struct {
	Route int `json:"route" binding:"required"`
}

func addTeamData(teamList []gin.H, teamResultSet *[]model.Team) []gin.H {
	for _, team := range *teamResultSet {
		teamList = append(teamList, gin.H{
			"id":     team.ID,
			"name":   team.Name,
			"num":    team.Num,
			"slogan": team.Slogan,
			"route":  team.Route,
		})
	}

	return teamList
}

func GetRandomList(context *gin.Context) {
	// 解析请求数据
	var getRandomListData GetRandomListData
	err := context.ShouldBindJSON(&getRandomListData)
	if err != nil { // 参数发送错误
		utility.ResponseError(context, "参数错误")
		return
	}

	// 获取列表
	var teams []model.Team
	var teamList []gin.H

	// 先查找 3 人以下的团队
	global.DB.Model(&model.Team{}).
		Where("route = ? AND num <= 3 AND allow_match = 1", getRandomListData.Route).
		Order("RAND()").
		Limit(3).
		Find(&teams)
	teamNum1 := len(teams)
	teamList = addTeamData(teamList, &teams)

	// 查找 4 人团队
	teams = teams[:0]
	global.DB.Model(&model.Team{}).
		Where("route = ? AND num = 4 AND allow_match = 1", getRandomListData.Route).
		Order("RAND()").
		Limit(4 - teamNum1).
		Find(&teams)
	teamNum2 := len(teams)
	teamList = addTeamData(teamList, &teams)

	// 查找 5 人团队
	teams = teams[:0]
	global.DB.Model(&model.Team{}).
		Where("route = ? AND num = 5 AND allow_match = 1", getRandomListData.Route).
		Order("RAND()").
		Limit(5 - teamNum2 - teamNum1).
		Find(&teams)
	teamNum3 := len(teams)
	teamList = addTeamData(teamList, &teams)

	if teamNum1+teamNum2+teamNum3 == 0 { // 没有查询结果
		utility.ResponseError(context, "No result")
	} else {
		utility.ResponseSuccess(context, gin.H{
			"teams": teamList,
		})
	}
}
