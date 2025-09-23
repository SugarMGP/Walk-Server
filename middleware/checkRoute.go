package middleware

import "walk-server/model"

// CheckRoute 检查管理员权限
func CheckRoute(admin *model.Admin, team *model.Team) bool {
	return team.Route == admin.Route ||
		(team.Route == 4 && admin.Route == 5) ||
		(team.Route == 5 && admin.Route == 4) ||
		(team.Route == 2 && admin.Route == 3) ||
		(team.Route == 3 && admin.Route == 2)
}
