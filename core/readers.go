package core

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/binacsgo/log"
)

type updata struct {
	data     string
	instance string
	// other
}

type reader struct {
	manager *ManagerService
	logger  log.Logger
	ipcFile string
}

type ReadersService struct {
	Manager *ManagerService `inject-name:"ExporterService"`
	Logger  log.Logger      `inject-name:"ReadersLogger"`
	Readers map[string]*reader
}

func (rs *ReadersService) AfterInject() error {
	rs.Readers = make(map[string]*reader)
	return nil
}

// AddReader add a reader
func (rs *ReadersService) AddReader(file string) {
	r := &reader{
		manager: rs.Manager,
		ipcFile: file,
		logger:  rs.Logger,
	}
	rs.Logger.Debug("AddReader: ", "reader", r.ipcFile)
	rs.Readers[file] = r
	go r.loop()
}

func (r *reader) loop() {
	for {
		// mutex ?
		oridata := r.readMetrics()
		if len(oridata) < 10 {
			continue
		}
		newdata := r.buildNewData(oridata)
		r.collectMetrics(newdata)
	}
}

// readMetrics read data from ipcFile and return string(data)
func (r *reader) readMetrics() string {
	data, err := ioutil.ReadFile(r.ipcFile)
	if err != nil {
		r.logger.Debug("readMetrics ReadFile err!", "ipcFile", r.ipcFile, "err", err)
		time.Sleep(10 * time.Second)
	}
	return string(data)
}

// buildNewData return updata based on oridata
func (r *reader) buildNewData(oridata string) updata {
	// get instance, for further maybe get from metachain
	start := strings.Index(oridata, "instance") + 10
	len := strings.Index(oridata[start:], "\"")
	instance := oridata[start : start+len]
	rt := updata{
		data:     oridata,
		instance: instance,
	}
	return rt
}

// collectMetrics push data into dataChan
func (r *reader) collectMetrics(data updata) {
	select {
	case r.manager.DataChan <- data:
		r.logger.Info("collectMetrics dataChan <- data", "file", r.ipcFile, "data size", len(data.data))
		r.logger.Debug("collectMetrics dataChan <- data", "data", data.data)
		break
	default:
		r.logger.Error("collectMetrics dataChan is full", "channel size", len(r.manager.DataChan))
	}
}
