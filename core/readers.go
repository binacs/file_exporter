package core

import (
	"context"
	"io/ioutil"
	"runtime"
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
	ctx     context.Context
	cancel  context.CancelFunc
}

func (rs *ReadersService) AfterInject() error {
	rs.Readers = make(map[string]*reader)
	rs.ctx, rs.cancel = context.WithCancel(context.Background())
	return nil
}

func (rs *ReadersService) Cancel() {
	rs.Logger.Info("ReadersService Cancel, before cancel", "go routines", runtime.NumGoroutine())
	rs.cancel()
	time.Sleep(2 * time.Second)
	rs.Logger.Info("ReadersService Cancel, after cancel", "go routines", runtime.NumGoroutine())
	rs.ctx, rs.cancel = context.WithCancel(context.Background())
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
	go r.loop(rs.ctx)
}

func (r *reader) loop(ctx context.Context) {
	t := time.NewTimer(constReaderInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// mutex ?
			oridata := r.readMetrics()
			if len(oridata) < 10 {
				continue
			}
			newdata := r.buildNewData(oridata)
			r.collectMetrics(newdata)
			t.Reset(constReaderInterval)
		}
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
