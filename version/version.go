package version

const (
	Maj = "1"
	Min = "0"
	Fix = "0"
)

var (
	Version         = "Unknown"
	ExporterVersion = "1.0.0"
	GitCommit       string
)

func init() {
	if GitCommit != "" {
		ExporterVersion += "-" + GitCommit
	}
	Version = ExporterVersion
}
