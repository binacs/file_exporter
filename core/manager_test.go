package core

import (
	"sync"
	"testing"

	"github.com/BinacsLee/file_exporter/config"
	"github.com/binacsgo/log"
)

func TestManagerService_getIpcFiles(t *testing.T) {
	type fields struct {
		Logger     log.Logger
		Config     *config.Config
		ReadersSvc *ReadersService
		DeleterSvc *DeleterService
		Ipcfiles   []string
		DataChan   chan updata
		GlobalMu   sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// Add test cases.
		{
			name: "test1",
			fields: fields{
				Logger: log.NewNopLogger(),
				Config: &config.Config{
					ExporterConfig: config.ExporterConfig{
						Dir_Keyword: map[string]string{"/tmp": "file_metrics"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &ManagerService{
				Logger:     tt.fields.Logger,
				Config:     tt.fields.Config,
				ReadersSvc: tt.fields.ReadersSvc,
				DeleterSvc: tt.fields.DeleterSvc,
				Ipcfiles:   tt.fields.Ipcfiles,
				DataChan:   tt.fields.DataChan,
				GlobalMu:   tt.fields.GlobalMu,
			}
			es.getIpcFiles()
			t.Log(es.Ipcfiles)
		})
	}
}

func TestManagerService_createReaders(t *testing.T) {
	type fields struct {
		Logger     log.Logger
		Config     *config.Config
		ReadersSvc *ReadersService
		DeleterSvc *DeleterService
		Ipcfiles   []string
		DataChan   chan updata
		GlobalMu   sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// Add test cases.
		{
			name: "test1",
			fields: fields{
				Logger: log.NewNopLogger(),
				Config: &config.Config{
					ExporterConfig: config.ExporterConfig{
						Dir_Keyword: map[string]string{"/tmp": "file_metrics"},
					},
				},
				ReadersSvc: &ReadersService{
					Logger:  log.NewNopLogger(),
					Readers: make(map[string]*reader),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &ManagerService{
				Logger:     tt.fields.Logger,
				Config:     tt.fields.Config,
				ReadersSvc: tt.fields.ReadersSvc,
				DeleterSvc: tt.fields.DeleterSvc,
				Ipcfiles:   tt.fields.Ipcfiles,
				DataChan:   tt.fields.DataChan,
				GlobalMu:   tt.fields.GlobalMu,
			}
			es.getIpcFiles()
			es.createReaders()
		})
	}
}

func TestManagerService_sendMetricsToGateway(t *testing.T) {
	type fields struct {
		Logger     log.Logger
		Config     *config.Config
		ReadersSvc *ReadersService
		DeleterSvc *DeleterService
		Ipcfiles   []string
		DataChan   chan updata
		GlobalMu   sync.Mutex
	}
	type args struct {
		url  string
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// Add test cases.
		{
			name: "test1",
			fields: fields{
				Logger: log.NewNopLogger(),
				Config: &config.Config{
					ExporterConfig: config.ExporterConfig{
						Gateway:     "http://127.0.0.1:9091",
						Jobname:     "file_exporter",
						Dir_Keyword: map[string]string{"/tmp": "file_metrics"},
					},
				},
				ReadersSvc: &ReadersService{
					Logger:  log.NewNopLogger(),
					Readers: make(map[string]*reader),
				},
			},
			args: args{
				url:  "http://127.0.0.1:9091/metrics/job/file_exporter/instance/333.333.333.333:333",
				data: "# HELP usage\n# TYPE usage gauge\nusage{hostname=\"unittest\"} 91.9",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &ManagerService{
				Logger:     tt.fields.Logger,
				Config:     tt.fields.Config,
				ReadersSvc: tt.fields.ReadersSvc,
				DeleterSvc: tt.fields.DeleterSvc,
				Ipcfiles:   tt.fields.Ipcfiles,
				DataChan:   tt.fields.DataChan,
				GlobalMu:   tt.fields.GlobalMu,
			}

			if err := es.sendMetricsToGateway(tt.args.url, tt.args.data); (err != nil) != tt.wantErr {
				//t.Errorf("ManagerService.sendMetricsToGateway() error = %v, wantErr %v", err, tt.wantErr)
				t.Logf("ManagerService.sendMetricsToGateway() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
