# io_blaster

### about
io_blaster is a tool designed to run extremely fast IO in order to run stress tests.
It was also designed to have great control on the IO scheduling and content in order to be able to create various scenarios with ease.

io_blaster is still a work in progress and currently contain only HTTP, remote shell IO workloads.

### install
* go get github.com/iguazio/io_blaster
* cd $GOPATH/src/github.com/iguazio/io_blaster
* go build

### guide
##### how to run
./io_blaster -c <config_file_path> -o <output_file_path>
* Notice that the <output_file_path> actual path will be <output_file_path>.log
* You can add -v option to run in verbose mode (this will generate a log for every request sent and also print worker stats in the end in addition to the workload stats)

##### config format
For accurate understanding of the config file format check the [struct ConfigIoBlaster](https://github.com/iguazio/io_blaster/blob/master/Config/Config.go#L127) and some of the [test files](https://github.com/iguazio/io_blaster/tree/master/test_files).
