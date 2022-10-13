package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/os/glog"
	"jackpot-ts/gots"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
func main() {

	//// pprof 端口
	//go func() {
	//	port := "0.0.0.0:9876"
	//	g.Log().Infof("try pprof for tetris http server run ,port %s \n", port)
	//	err := http.ListenAndServe(port, nil)
	//	if err != nil {
	//		g.Log().Infof("pprof for tetris http server run err[%+v] \n", err)
	//		return
	//	}
	//}()
	//dao.JackpotPlayerInfo.DB.SetDebug(true)
	//dao.JackpotGlobalInfo.DB.SetDebug(true)
	//dao.JackpotSpinRecord.DB.SetDebug(true)
	//dao.MockMoney.DB.SetDebug(true)

	f, err := os.Create("./log.log")
	if err != nil {
		panic(err)
	}
	defaultLogger := glog.DefaultLogger()
	defaultLogger.SetWriter(f)
	glog.SetDefaultLogger(defaultLogger)
	taskPool := gots.NewTaskPool(200)
	defer taskPool.Stop()
	g := gin.Default()
	g.Use(Cors())
	g.Handle(http.MethodGet, "/changeSpinLevel", taskPool.ChangeSpinLevel)
	g.Handle(http.MethodGet, "/queryMoney", taskPool.QueryMoney)
	g.Handle(http.MethodGet, "/spin", taskPool.Spin)
	g.Handle(http.MethodGet, "/config", taskPool.QueryConfig)
	g.Handle(http.MethodGet, "/broadcast", taskPool.QueryBroadCast)
	//g.Handle(http.MethodGet, "/exchangeMoney", taskPool.ExchangeMoney)
	g.Handle(http.MethodGet, "/addMoney", taskPool.AddMoney)
	runError := g.Run("0.0.0.0:9800")
	if runError != nil {
		panic(runError)
	}
}
