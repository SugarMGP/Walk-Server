package userService

import (
	"walk-server/global"
	"walk-server/model"
)

func GetUserByID(id string) (*model.Person, error) {
	var person model.Person
	result := global.DB.Where("identity = ?", id).First(&person)
	return &person, result.Error
}

func GetUserByOpenID(oid string) (*model.Person, error) {
	var person model.Person
	result := global.DB.Where("open_id = ?", oid).First(&person)
	return &person, result.Error
}

func GetUsersByTeamID(teamID uint) ([]model.Person, error) {
	var users []model.Person
	result := global.DB.Where("team_id = ?", teamID).Find(&users)
	return users, result.Error
}
