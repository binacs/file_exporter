package core

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/BinacsLee/file_exporter/config"

	"github.com/binacsgo/log"
	"github.com/binacsgo/treemap"
)

// DeleteMsg key-interface
type DeleteMsg struct {
	Instance string
	SentTime int64
	Expire   int64
}

func (msg1 DeleteMsg) LessThan(msg interface{}) bool {
	msg2 := msg.(DeleteMsg)
	if msg1.SentTime == msg2.SentTime {
		return msg1.Instance < msg2.Instance
	}
	return msg1.SentTime < msg2.SentTime
}

func (msg1 DeleteMsg) Equal(msg interface{}) bool {
	msg2 := msg.(DeleteMsg)
	return msg1.Instance == msg2.Instance
}

// DeleterService
type DeleterService struct {
	Logger     log.Logger     `inject-name:"DeleterLogger"`
	Config     *config.Config `inject-name:"Config"`
	OrderMap   *treemap.TreeMap
	HasMsgChan chan bool
}

func (ds *DeleterService) AfterInject() error {
	ds.OrderMap = treemap.NewMap()
	ds.HasMsgChan = make(chan bool, constMaxIpcfiles)
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
			if ds.OrderMap.Empty() {
				ds.Logger.Error("Unexpected! order map is empty while HasMsgChan is not empty", "len(chan)", len(ds.HasMsgChan))
				continue
			}
			for {
				now := time.Now().Unix()
				msg := ds.OrderMap.Min().Key.(DeleteMsg)
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
	client := &http.Client{}
	resp, err := client.Do(req)
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
	msg := DeleteMsg{
		Instance: instance,
		SentTime: time.Now().Unix(),
		Expire:   expire,
	}
	update := ds.OrderMap.Store(msg, msg.SentTime) // value to be decided
	if !update {
		ds.HasMsgChan <- true
	}
	ds.Logger.Debug("after add", "map-size", ds.OrderMap.OrderMap.Size())
}

func (ds *DeleterService) DelMsg(msg DeleteMsg) {
	ds.OrderMap.Delete(msg)
	ds.Logger.Debug("after delete", "map-size", ds.OrderMap.OrderMap.Size())
}
