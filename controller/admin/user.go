package admin

import (
	"errors"
	"fmt"
	"sort"
	"time"
	"walk-server/constant"
	"walk-server/global"
	"walk-server/middleware"
	"walk-server/model"
	"walk-server/service/adminService"
	"walk-server/service/teamService"
	"walk-server/service/userService"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type UserStatusForm struct {
	UserID string `json:"user_id" binding:"required"`
	Status int    `json:"status" binding:"required,oneof=1 2"`
}

type UserStatusList struct {
	List []UserStatusForm `json:"list" binding:"required"`
}

// UserStatus handles user status updates
func UserStatus(c *gin.Context) {
	var postForm UserStatusList
	if err := c.ShouldBindJSON(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	// 获取管理员信息
	user, _ := adminService.GetAdminByJWT(c)

	// 批量获取用户和队伍信息
	users, teams, err := getUsersAndTeams(postForm.List)
	if err != nil {
		utility.ResponseError(c, err.Error())
		return
	}

	// 验证用户权限
	for _, person := range users {
		team, exists := teams[person.TeamId]
		if !exists {
			utility.ResponseError(c, "队伍信息获取失败")
			return
		}

		// 管理员只能管理自己所在的校区
		if !middleware.CheckRoute(user, &team) {
			utility.ResponseError(c, "该队伍为其他路线")
			return
		}

		// 验证毅行状态
		if person.WalkStatus == 5 {
			utility.ResponseError(c, "成员已结束毅行")
			return
		}
	}

	// 更新用户状态
	for _, form := range postForm.List {
		person := users[form.UserID]
		if form.Status == 1 {
			person.WalkStatus = 3
		} else {
			person.WalkStatus = 4
		}
		userService.Update(*person)
	}

	// 检查队伍是否已经没人在行
	for _, user := range users {
		num := 0
		team, exists := teams[user.TeamId]
		if !exists {
			continue
		}
		persons, err := userService.GetUsersByTeamID(team.ID)
		if err != nil {
			utility.ResponseError(c, "获取队伍成员失败")
			return
		}
		for _, person := range persons {
			if person.WalkStatus != 4 {
				num++
			}
		}
		if num == 0 {
			team.Status = 3
			teamService.Update(team)
		}
	}

	utility.ResponseSuccess(c, nil)
}

// getUsersAndTeams retrieves user and team data for the given user IDs
func getUsersAndTeams(forms []UserStatusForm) (map[string]*model.Person, map[int]model.Team, error) {
	userMap := make(map[string]*model.Person)
	teamMap := make(map[int]model.Team)

	for _, form := range forms {
		person, err := model.GetPerson(form.UserID)
		if err != nil {
			return nil, nil, errors.New("扫码错误，查找用户失败，请再次核对")
		}
		userMap[form.UserID] = person

		if _, exists := teamMap[person.TeamId]; !exists {
			var team model.Team
			if err := global.DB.Where("id = ?", person.TeamId).Take(&team).Error; err != nil {
				return nil, nil, errors.New("队伍信息获取失败")
			}
			teamMap[person.TeamId] = team
		}
	}

	return userMap, teamMap, nil
}

type GetTimeoutUsersData struct {
	Minute int    `form:"minute" binding:"required"` // 超时时间
	Route  uint8  `form:"route" binding:"required"`  // 路线
	Type   uint8  `form:"type"`                      // 类型
	Secret string `form:"secret" binding:"required"` // 密钥
}

type User struct {
	Name       string    `json:"name"`
	Gender     int8      `json:"gender"` // 1 男，2 女
	StuId      string    `json:"stu_id"`
	Campus     uint8     `json:"campus"`  // 1 朝晖，2 屏峰，3 莫干山
	College    string    `json:"college"` // 学院
	Tel        string    `json:"tel"`
	Type       uint8     `json:"type"` // 1 学生， 2 教职工
	Time       time.Time `json:"time"`
	Point      int8      `json:"point"`
	TeamID     uint      `json:"team_id"`
	TeamName   string    `json:"team_name"`
	Status     uint8     `json:"status"`      // 1 队员，2 队长
	WalkStatus uint8     `json:"walk_status"` // 1 未开始，2 进行中，3 扫码成功，4 放弃，5 完成
	Location   string    `json:"location"`
}

type PointUsers struct {
	Point    int8   `json:"point"`
	Location string `json:"location"`
	Users    []User `json:"users"`
}

func GetTimeoutUsers(c *gin.Context) {
	var postForm GetTimeoutUsersData
	if err := c.ShouldBindQuery(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	// 获取超时队伍
	teamMap, err := adminService.GetTimeoutTeams(postForm.Minute, postForm.Route)
	if err != nil {
		utility.ResponseError(c, "获取失败，请稍后重试")
		return
	}

	// 获取未到队伍
	noShowTeams, err := adminService.GetNoShowTeams(postForm.Route)
	if err != nil {
		utility.ResponseError(c, "获取失败，请稍后重试")
		return
	}

	results := make([]PointUsers, 0, len(teamMap)) // 按 teamMap 的长度预分配空间

	for point, teams := range teamMap {
		users := make([]User, 0, len(teams)*6) // 按 teams 的长度预分配空间
		for _, team := range teams {
			person, persons := model.GetPersonsInTeam(int(team.ID))

			// 筛选成员
			filteredUsers := make([]User, 0, 6)
			switch postForm.Type {
			case 0:
				filteredUsers = append(filteredUsers, buildUserData(person, team))
				for _, member := range persons {
					filteredUsers = append(filteredUsers, buildUserData(member, team))
				}
			case 1, 2, 3:
				filteredUsers = append(filteredUsers, filterMembersByType(person, persons, postForm.Type, team)...)
			}

			users = append(users, filteredUsers...)
		}
		// 存入有序结构
		results = append(results, PointUsers{Point: point, Location: constant.GetPointName(postForm.Route, point), Users: users})

	}
	users := make([]User, 0, len(noShowTeams)*4) // 按 teams 的长度预分配空间
	for _, team := range noShowTeams {
		person, persons := model.GetPersonsInTeam(int(team.ID))

		// 筛选成员
		filteredUsers := make([]User, 0, 6)
		switch postForm.Type {
		case 0:
			filteredUsers = append(filteredUsers, buildUserData(person, team))
			for _, member := range persons {
				filteredUsers = append(filteredUsers, buildUserData(member, team))
			}
		case 1, 2, 3:
			filteredUsers = append(filteredUsers, filterMembersByType(person, persons, postForm.Type, team)...)
		}

		users = append(users, filteredUsers...)
	}
	results = append(results, PointUsers{Point: -1, Location: "未到", Users: users})

	// 按 `point` 排序，保证顺序
	sort.Slice(results, func(i, j int) bool { return results[i].Point < results[j].Point })

	// 返回成功的响应，包含用户信息
	utility.ResponseSuccess(c, gin.H{
		"results": results,
	})
}

func DownloadTimeoutUsers(c *gin.Context) {
	var postForm GetTimeoutUsersData
	if err := c.ShouldBindQuery(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	// 获取超时队伍
	teamMap, err := adminService.GetTimeoutTeams(postForm.Minute, postForm.Route)
	if err != nil {
		utility.ResponseError(c, "获取失败，请稍后重试")
		return
	}

	// 获取用户信息（有序）
	var allUsers []User // 所有用户
	var genderMap = map[int8]string{1: "男", 2: "女"}
	var campusMap = map[uint8]string{1: "朝晖", 2: "屏峰", 3: "莫干山"}
	var typeMap = map[uint8]string{1: "学生", 2: "教职工", 3: "校友"}
	var statusMap = map[uint8]string{1: "队员", 2: "队长"}
	var walkStatusMap = map[uint8]string{1: "未开始", 2: "进行中", 3: "进行中", 4: "放弃", 5: "已完成"}

	point := constant.PointMap[postForm.Route]
	var i uint8
	var teams []model.Team
	for i = 0; i < point; i++ {
		if _, exists := teamMap[int8(i)]; exists {
			teams = teamMap[int8(i)]
		} else {
			continue
		}
		for _, team := range teams {
			person, persons := model.GetPersonsInTeam(int(team.ID))

			// 筛选成员
			var filteredUsers []User
			switch postForm.Type {
			case 0:
				filteredUsers = append(filteredUsers, buildUserData(person, team))
				for _, member := range persons {
					filteredUsers = append(filteredUsers, buildUserData(member, team))
				}
			case 1, 2, 3:
				filteredUsers = append(filteredUsers, filterMembersByType(person, persons, postForm.Type, team)...)
			}

			allUsers = append(allUsers, filteredUsers...)
		}
	}

	// 构建表头
	headers := []string{"上个点位", "上个点位签到时间", "队伍编号", "队伍名称", "姓名", "队伍担当", "当前状态", "性别", "学号", "电话", "校区", "学院", "参与者类型"}

	// 构建所有用户的行

	var rows [][]any
	for _, user := range allUsers {
		point := constant.GetPointName(postForm.Route, user.Point)

		row := []any{
			point,                                   // 上个点位
			user.Time.Format("2006-01-02 15:04:05"), // 到达上个点位时间
			user.TeamID,                             // 队伍编号
			user.TeamName,                           // 队伍名称
			user.Name,                               // 姓名
			statusMap[user.Status],                  // 队伍担当
			walkStatusMap[user.WalkStatus],          // 当前状态
			genderMap[user.Gender],                  // 性别
			user.StuId,                              // 学号
			user.Tel,                                // 电话
			campusMap[user.Campus],                  // 校区
			user.College,                            // 学院
			typeMap[user.Type],                      // 类型
		}
		rows = append(rows, row)
	}

	// 创建 Excel 文件
	data := utility.File{
		Sheets: []utility.Sheet{
			{
				Name:    "超时用户信息", // Sheet 名称
				Headers: headers,
				Rows:    rows,
			},
		},
	}

	// 保存为 Excel 文件
	fileName := "Route" + fmt.Sprint(postForm.Route) + "TimeoutUser.xlsx"
	filePath := "./file/" // 文件保存路径
	host := global.Config.GetString("frontend.url")
	url, err := utility.CreateExcelFile(data, fileName, filePath, host)
	if err != nil {
		utility.ResponseError(c, "生成文件失败")
		return
	}

	// 返回文件下载 URL
	utility.ResponseSuccess(c, gin.H{
		"url": url,
	})
}

func buildUserData(person model.Person, team model.Team) User {
	return User{
		Name:       person.Name,
		Gender:     person.Gender,
		StuId:      person.StuId,
		Campus:     person.Campus,
		College:    person.College,
		Tel:        person.Tel,
		Type:       person.Type,
		Time:       team.Time,
		Point:      team.Point,
		TeamID:     team.ID,
		TeamName:   team.Name,
		Status:     person.Status,
		WalkStatus: person.WalkStatus,
		Location:   constant.GetPointName(team.Route, team.Point),
	}
}

func filterMembersByType(person model.Person, persons []model.Person, userType uint8, team model.Team) []User {
	var result []User

	if person.Type == userType {
		result = append(result, buildUserData(person, team))
	}

	// 筛选符合条件的其他成员
	for _, member := range persons {
		if member.Type == userType {
			result = append(result, buildUserData(member, team))
		}
	}
	return result
}
