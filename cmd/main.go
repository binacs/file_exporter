package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/BinacsLee/file_exporter/config"
	"github.com/BinacsLee/file_exporter/core"
	"github.com/BinacsLee/file_exporter/version"

	"github.com/binacsgo/inject"
	"github.com/binacsgo/log"
)

var (
	versionFlag = flag.Bool("version", false, "Show version info")
	configFile  = flag.String("configfile", "./config.toml", "The config file path, better the absolute path")
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("%s\n", version.Version)
		return
	}
	// *Config
	cfg, err := config.LoadFromFile(*configFile)
	if err != nil {
		panic(err)
	}
	fmt.Println("config : ", cfg)

	// inject service
	logger := log.Sugar()
	exporter, httpSvc := initService(logger, cfg)
	exporter.OnStart()
	startHttpService(logger, cfg, httpSvc)
}

// initService init services
func initService(logger log.Logger, cfg *config.Config) (*core.ManagerService, *core.HttpService) {
	manager := core.ManagerService{}
	httpSvc := core.HttpService{}
	inject.Regist(Inject_LOGGER, logger)
	inject.Regist(Inject_Manager_LOGGER, logger.With("module", "manager"))
	inject.Regist(Inject_Readers_LOGGER, logger.With("module", "readers"))
	inject.Regist(Inject_Deleter_LOGGER, logger.With("module", "deleter"))
	inject.Regist(Inject_HttpSvc_LOGGER, logger.With("module", "httpsvc"))
	inject.Regist(Inject_Config, cfg)
	inject.Regist(Inject_Manager, &manager)
	inject.Regist(Inject_Readers, &core.ReadersService{})
	inject.Regist(Inject_Deleter, &core.DeleterService{})
	inject.Regist(Inject_Http, &httpSvc)
	err := inject.DoInject()
	if err != nil {
		panic(err.Error())
	}
	logger.Info("Inject finish")
	return &manager, &httpSvc
}

func startHttpService(logger log.Logger, cfg *config.Config, httpSvc *core.HttpService) {
	r := gin.New()
	r.Use(gin.Recovery())
	httpSvc.SetRouter(r)
	s := &http.Server{
		Addr:           ":" + cfg.ExporterConfig.HttpPort,
		Handler:        r,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
	logger.Info("http start finish", "port", cfg.ExporterConfig.HttpPort)
}
