# file_exporter

file_exporter's function:

> Collect metrics-data from files and send them to prometheus-pushgateway

## 0. Quick Start

Run exporter just like this:

```shell
$ make

$ ./bin/exporter --configfile ./config.toml
```

## 1. Usage

File_exporter retrieve all files containing the keywords in the directory, read their content and send them to pushgateway.

Eg:

```toml
[ExporterConfig]
RootPath = "/path/to/workspace/"       # workspace
Gateway = "http://127.0.0.1:9091"      # url of the prometheus-pushgateway 
[ExporterConfig.Dir_Keyword]           # directory and the keywords
"/tmp" = "file_metrics"
```

File_exporter will read all the files named "file_metrics*" under the "tmp" directory.

## 2. What's more?

Based on [PQueue](https://github.com/binacsgo/pqueue), file_exporter record the send time of each file, and it will delete the data on pushgateway in some time.
