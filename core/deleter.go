package core

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/BinacsLee/file_exporter/config"

	"github.com/binacsgo/log"
	"github.com/binacsgo/pqueue"
)

// DeleteMsg key-interface
type DeleteMsg struct {
	Instance string
	SentTime int64
	Expire   int64
}

func (msg1 *DeleteMsg) KeyEqual(msg interface{}) bool {
	msg2 := msg.(*DeleteMsg)
	return msg1.Instance == msg2.Instance
}

// DeleterService
type DeleterService struct {
	Logger     log.Logger     `inject-name:"DeleterLogger"`
	Config     *config.Config `inject-name:"Config"`
	HasMsgChan chan bool
	PQueue     *pqueue.PQueue
	httpc      *http.Client
}

func (ds *DeleterService) AfterInject() error {
	ds.HasMsgChan = make(chan bool, constMaxIpcfiles)
	ds.PQueue = pqueue.NewPQueue()
	ds.httpc = &http.Client{Timeout: constHTTPTimeout}
	ds.OnStart()
	return nil
}

func (ds *DeleterService) OnStart() {
	go ds.loop()
}

func (ds *DeleterService) loop() {
	for {
		select {
		case <-ds.HasMsgChan:
			if ds.PQueue.Size() == 0 {
				ds.Logger.Error("Unexpected! PQueue is empty while HasMsgChan is not empty", "len(chan)", len(ds.HasMsgChan))
				continue
			}
			for {
				now := time.Now().Unix()
				msg := ds.PQueue.GetMin().(*DeleteMsg)
				if msg.SentTime == -1 {
					break
				}
				url := ds.Config.ExporterConfig.Gateway + "/metrics/job/" + ds.Config.ExporterConfig.Jobname + "/instance/" + msg.Instance
				if now >= msg.SentTime+msg.Expire {
					err := ds.deleteMetrics(url)
					if err != nil {
						ds.Logger.Error("deleteMetrics error.", "err", err)
						// delete fail, do as follows:
						// TODO check channel full or not
						ds.HasMsgChan <- true
						break
					}
					ds.DelMsg(msg)
					break
				}
				time.Sleep(time.Second)
			}
		}
	}
}

func (ds *DeleterService) deleteMetrics(url string) error {
	ds.Logger.Info("deleteMetrics", "url", url)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return errors.New("deleteMetrics NewRequest Error.")
	}
	resp, err := ds.httpc.Do(req)
	if err != nil {
		return errors.New("deleteMetrics Do request Error.")
	}
	defer resp.Body.Close()
	// Pushgateway 0.10+ responds with StatusOK, earlier versions with StatusAccepted.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(resp.Body) // Ignore any further error as this is for an error message only.
		ds.Logger.Error("unexpected status code", "code", resp.StatusCode, "target", url, "body", string(body))
		return errors.New("deleteMetrics response status Error.")
	}
	ds.Logger.Info("deleteMetrics Succeed", "url", url)
	return nil
}

func (ds *DeleterService) AddMsg(instance string, expire int64) {
	ds.Logger.Debug("before add", "pqueue-size", ds.PQueue.Size())
	msg := &DeleteMsg{
		Instance: instance,
		SentTime: time.Now().Unix(),
		Expire:   expire,
	}
	update := ds.PQueue.Set(msg.Instance, msg)
	if !update {
		ds.HasMsgChan <- true
	}
	ds.Logger.Debug("after add", "pqueue-size", ds.PQueue.Size(), "update", update, "instance", msg.Instance)
}

func (ds *DeleterService) DelMsg(msg *DeleteMsg) {
	ds.Logger.Debug("before delete", "pqueue-size", ds.PQueue.Size())
	ds.PQueue.DelMin()
	ds.Logger.Debug("after delete", "pqueue-size", ds.PQueue.Size(), "instance", msg.Instance)
}
