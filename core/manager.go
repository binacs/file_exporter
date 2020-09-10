package core

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/BinacsLee/file_exporter/config"

	"github.com/binacsgo/log"
)

type ManagerService struct {
	Logger     log.Logger      `inject-name:"ManagerLogger"`
	Config     *config.Config  `inject-name:"Config"`
	ReadersSvc *ReadersService `inject-name:"ReadersService"`
	DeleterSvc *DeleterService `inject-name:"DeleterService"`
	Ipcfiles   []string
	DataChan   chan updata // string or other struct with more information
	GlobalMu   sync.Mutex
}

func (es *ManagerService) AfterInject() error {
	es.Ipcfiles = []string{}
	es.DataChan = make(chan updata, constMaxChanSize)
	return nil
}

func (es *ManagerService) OnStart() {
	es.getIpcFiles()
	es.createReaders()
	go es.loop()
	es.Logger.Info("exporter OnStart() succeed.")
}

func (es *ManagerService) ReStart() {
	es.getIpcFiles()
	es.createReaders()
	es.Logger.Info("exporter ReStart() succeed.")
}

// getIpcFiles parse the configs.Dir_Keyword and append filepaths to Ipcfiles([]string)
func (es *ManagerService) getIpcFiles() {
	es.Ipcfiles = []string{}
	for dir, keyword := range es.Config.ExporterConfig.Dir_Keyword {
		file_list, err := ioutil.ReadDir(dir)
		if err != nil {
			es.Logger.Error("getIpcFiles error", "dir", dir, "err", err)
		}
		for _, v := range file_list {
			if strings.Contains(v.Name(), keyword) {
				if string(dir[len(dir)-1]) != "/" {
					dir += "/"
				}
				file := dir + v.Name()
				es.Ipcfiles = append(es.Ipcfiles, file)
			}
		}
	}
	es.Logger.Info("getIpcFiles", "files", es.Ipcfiles)
}

// createReaders create reader for each Ipcfile
func (es *ManagerService) createReaders() {
	es.ReadersSvc.Readers = make(map[string]*reader)
	for _, file := range es.Ipcfiles {
		_, ok := es.ReadersSvc.Readers[file]
		es.Logger.Info("manager createReaders()", "exporter", file)
		if !ok {
			es.ReadersSvc.AddReader(file)
		}
	}
}

// loop read data from dataChan, joint 'url' and send the data to pushgateway
func (es *ManagerService) loop() {
	for {
		select {
		case data := <-es.DataChan:
			url := es.Config.ExporterConfig.Gateway + "/metrics/job/" + es.Config.ExporterConfig.Jobname + "/instance/" + data.instance
			// data中的其他字段亦可处理
			err := es.sendMetricsToGateway(url, data.data)
			if err != nil {
				// resend?
				es.Logger.Error("manager loop() sendMetricsToGateway Error.", "error", err)
				continue
			}
			es.DeleterSvc.AddMsg(data.instance, constExpireTime) // expire time to-be decided
		}
	}
}

// sendMetricsToGateway send data to the gateway, ATTENTION 'url' is diff from gateway URL
func (es *ManagerService) sendMetricsToGateway(url, data string) error {
	//url := es.configs.Gateway + "/metrics/job/jobname"
	es.Logger.Info("sendMetricsToGateway", "url", url, "data size", len(data))
	sr := strings.NewReader(data)
	br := bufio.NewReader(sr)
	req, err := http.NewRequest(http.MethodPost, url, br)
	if err != nil {
		return errors.New("sendMetricsToGateway NewRequest Error.")
	}
	//req.Header.Set("Content-Type", string(expfmt.FmtProtoDelim))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("sendMetricsToGateway Do request Error.")
	}
	defer resp.Body.Close()
	// Pushgateway 0.10+ responds with StatusOK, earlier versions with StatusAccepted.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(resp.Body) // Ignore any further error as this is for an error message only.
		es.Logger.Error("unexpected status code", "code", resp.StatusCode, "target", url, "body", string(body))
		return errors.New("sendMetricsToGateway response status Error.")
	}
	es.Logger.Info("sendMetricsToGateway Succeed", "url", url, "data size", len(data))
	return nil
}
