package admin

import (
	"crypto/rand"
	"github.com/gin-gonic/gin"
	"strings"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"
)

type Data struct {
	Name    string `json:"name"`
	Account string `json:"account"` // 学号
}

type CreateRouteAdminData struct {
	ZH      [][]Data `json:"zh" `
	PFHalf  [][]Data `json:"pf_half"`
	PFAll   [][]Data `json:"pf_all"`
	MGSHalf [][]Data `json:"mgs_half"`
	MGSAll  [][]Data `json:"mgs_all"`
	Secret  string   `json:"secret" binding:"required"`
}

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // 包含字母和数字的字符集 	// 6 bits 用于表示一个字母的索引 	// 63 bits 中可以容纳的字母索引的数量 	// 替换成你想要的固定密码
)

func CreateRouteAdmin(c *gin.Context) {
	var postForm CreateRouteAdminData
	if err := c.ShouldBindJSON(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	admins := make([]model.Admin, 0)

	processData := func(data [][]Data, point int8) {
		for i := 0; i < len(data); i++ {
			for j := 0; j < len(data[i]); j++ {
				admins = append(admins, model.Admin{
					Name:     data[i][j].Name,
					Account:  data[i][j].Account,
					Password: generateRandomPassword(),
					Point:    point,
					Route:    uint8(j),
				})
			}
		}
	}

	processData(postForm.ZH, 1)
	processData(postForm.PFHalf, 2)
	processData(postForm.PFAll, 3)
	processData(postForm.MGSHalf, 4)
	processData(postForm.MGSAll, 5)

	result := global.DB.Create(&admins)
	if result.Error != nil {
		utility.ResponseError(c, "数据库错误: "+result.Error.Error())
		return
	}

	utility.ResponseSuccess(c, gin.H{
		"admins": admins,
	})
}

// generateRandomString 生成一个指定长度的随机字符串，使用字母和数字。
func generateRandomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)

	// 创建一个字节切片来存储随机字节
	randomBytes := make([]byte, n)

	// 使用 crypto/rand.Read 填充字节切片
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("无法生成随机字节: " + err.Error()) // 在生产环境中，不要 panic，而是返回错误
	}

	// 将随机字节转换为字符串
	for i := 0; i < n; i++ {
		// 使用字节值作为 letterBytes 的索引
		sb.WriteByte(letterBytes[randomBytes[i]%byte(len(letterBytes))])
	}

	return sb.String()
}

// generateRandomAccountAndPassword 生成一个随机账号和一个固定密码。
func generateRandomPassword() string {
	account := generateRandomString(6) // 6 个随机字母数字字符
	return account
}
