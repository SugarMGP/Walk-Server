package teamService

import (
	"walk-server/global"
	"walk-server/model"
)

func Update(a *model.Team) {
	global.DB.Save(a)
}

func Delete(a *model.Team) error {
	return global.DB.Delete(a).Error
}

func Create(a *model.Team) error {
	return global.DB.Create(a).Error
}

func UpdateCaptain(teamID int, openID string) error {
	return global.DB.Model(&model.Team{}).Where("id = ?", teamID).Update("captain", openID).Error
}
