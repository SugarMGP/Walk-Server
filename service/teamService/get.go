package teamService

import (
	"walk-server/global"
	"walk-server/model"
)

func GetTeamByID(id uint) (*model.Team, error) {
	team := model.Team{}
	result := global.DB.Where(
		&model.Team{
			ID: id,
		},
	).First(&team)

	if result.Error != nil {
		return nil, result.Error
	}
	return &team, nil
}

func GetTeamByCode(code string) (*model.Team, error) {
	team := model.Team{}
	result := global.DB.Where(
		&model.Team{
			Code: code,
		},
	).First(&team)

	return &team, result.Error
}

func GetTeamByCaptain(captainID string) (*model.Team, error) {
	team := model.Team{}
	result := global.DB.Where(
		&model.Team{
			Captain: captainID,
		},
	).First(&team)

	return &team, result.Error
}
