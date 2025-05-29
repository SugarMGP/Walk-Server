package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// Time 中间件：计算并记录请求处理的时间
func Time() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 继续处理请求
		c.Next()

		// 请求处理完毕后，计算耗时
		duration := time.Since(start)

		// 打印请求耗时（可以根据需求做修改，例如存入日志文件等）
		log.Printf("Request [%s] %s took %v", c.Request.Method, c.Request.URL, duration)
	}
}
