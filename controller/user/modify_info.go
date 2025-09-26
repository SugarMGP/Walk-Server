package user

import (
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type UserModifyData struct {
	Campus  uint8  `json:"campus" binding:"required"`
	College string `json:"college" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel" binding:"required"`
	} `json:"contact" binding:"required"`
}

func ModifyInfo(context *gin.Context) {
	// 获取 open ID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过数据了
	openID := jwtData.OpenID

	// 获取 post data
	var postData UserModifyData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "参数错误，请重试")
		return
	}

	// 获取个人信息
	person, _ := model.GetPerson(openID)
	person.Campus = postData.Campus
	person.Qq = postData.Contact.QQ
	person.Wechat = postData.Contact.Wechat
	person.Tel = postData.Contact.Tel
	person.College = postData.College
	person.Identity = postData.ID

	// 更新数据
	model.UpdatePerson(openID, person)
	utility.ResponseSuccess(context, nil)
}
