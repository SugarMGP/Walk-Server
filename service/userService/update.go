package userService

import (
	"walk-server/model"
)

func Update(a *model.Person) {
	model.UpdatePerson(a.OpenId, a)
}

func Set(openId string, a *model.Person) error {
	return model.SetPerson(openId, a)
}
