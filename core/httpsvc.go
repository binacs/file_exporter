package core

import (
	"net/http"

	"github.com/BinacsLee/file_exporter/config"
	"github.com/binacsgo/log"
	"github.com/gin-gonic/gin"
)

type HttpService struct {
	Logger     log.Logger      `inject-name:"HttpLogger"`
	ManagerSvc *ManagerService `inject-name:"ManagerService"`
	ReadersSvc *ReadersService `inject-name:"ReadersService"`
	Config     *config.Config  `inject-name:"Config"`
}

func (hs *HttpService) AfterInject() error {
	return nil
}

func (hs *HttpService) SetRouter(r *gin.Engine) {
	hs.SetManagerRounter(r.Group("manager"))
}

func (hs *HttpService) SetManagerRounter(r *gin.RouterGroup) {
	r.POST("/reload", hs.ReloadManager)
}

func (hs *HttpService) ReloadManager(ctx *gin.Context) {
	hs.Logger.Info("config reload start", "old config", hs.Config)
	err := hs.Config.Reload()
	if err != nil {
		hs.Logger.Error(err.Error())
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	hs.ManagerSvc.OnStart()
	hs.Logger.Info("config reload success", "new config", hs.Config)
	ctx.String(http.StatusOK, "reload finish")
	return
}
