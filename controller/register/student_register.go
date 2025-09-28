package register

import (
	"errors"
	"github.com/zjutjh/WeJH-SDK/oauth"
	"github.com/zjutjh/WeJH-SDK/oauth/oauthException"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// StudentRegisterData 定义接收学生报名用的数据的类型
type StudentRegisterData struct {
	StuID    string `json:"stu_id" binding:"required"`
	Password string `json:"password" binding:"required"`
	ID       string `json:"id" binding:"required"`
	Campus   uint8  `json:"campus" binding:"required"`
	College  string `json:"college" binding:"required"`
	Contact  struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel" binding:"required"`
	} `json:"contact" binding:"required"`
}

func StudentRegister(context *gin.Context) {
	// 获取 openID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过是否合法了

	// 获取报名数据
	var postData StudentRegisterData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "上传数据错误")
		return
	}

	var user model.Person
	result := global.DB.Where("stu_id =? Or identity = ? Or tel = ?", postData.StuID, postData.ID, postData.Contact.Tel).Take(&user)
	if result.RowsAffected != 0 {
		utility.ResponseError(context, "已有身份信息，请检查是否填写错误")
		return
	}

	cookie, info, err := oauth.GetUserInfo(postData.StuID, postData.Password)
	var oauthErr *oauthException.Error
	if errors.As(err, &oauthErr) {
		if errors.Is(oauthErr, oauthException.WrongAccount) || errors.Is(oauthErr, oauthException.WrongPassword) {
			utility.ResponseError(context, "账号或密码错误")
			return
		} else if errors.Is(oauthErr, oauthException.ClosedError) {
			utility.ResponseError(context, "统一夜间关闭，请白天尝试")
			return
		} else if errors.Is(oauthErr, oauthException.NotActivatedError) {
			utility.ResponseError(context, "账号未激活，请自行到统一网站重新激活")
			return
		} else if errors.Is(oauthErr, oauthException.EditPasswordError) {
			utility.ResponseError(context, "统一密码需要修改, 请手动登录统一修改")
			return
		}
	}
	if err != nil || cookie == nil {
		utility.ResponseError(context, "统一验证失败，请稍后再试")
		return
	}
	if info.UserTypeDesc == "教师职工" {
		utility.ResponseError(context, "您是教师职工，请返回教师注册")
		return
	}
	var gender int8
	if info.Gender == "male" {
		gender = 1
	} else {
		gender = 2
	}

	person := model.Person{
		OpenId:     jwtData.OpenID,
		Name:       info.Name,
		Gender:     gender,
		StuId:      postData.StuID,
		Status:     0,
		College:    postData.College,
		Identity:   postData.ID,
		Campus:     postData.Campus,
		Qq:         postData.Contact.QQ,
		Wechat:     postData.Contact.Wechat,
		Tel:        postData.Contact.Tel,
		CreatedOp:  2,
		JoinOp:     5,
		TeamId:     -1,
		WalkStatus: 1,
		Type:       1,
	}
	result = global.DB.Create(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "报名失败，请检查后重试")
	} else {
		utility.ResponseSuccess(context, nil)
	}
}
