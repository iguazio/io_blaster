# io_blaster

## about
io_blaster is a tool designed to run extremely fast IO in order to run stress tests.
It was also designed to have great control on the IO scheduling and content in order to be able to create various scenarios with ease.

io_blaster is still a work in progress and currently contain only HTTP, remote shell IO workloads.

## install
* go get github.com/iguazio/io_blaster
* cd $GOPATH/src/github.com/iguazio/io_blaster
* go build

## guide
### how to run
./io_blaster -c <config_file_path> -o <output_file_path>
* Notice that the <output_file_path> actual path will be <output_file_path>.log
* You can add -v option to run in verbose mode (this will generate a log for every request sent and also print worker stats in the end in addition to the workload stats)
* Notice that in order to run shell workloads with more than 10 workers you will need to edit MaxStartups param in sshd_config to support for more than 10 connections. To allow 100 connections run `sudo vi /etc/ssh/sshd_config` and add at the end `MaxStartups 100:30:200`

### config format

#### Config file should be a json with the following format:  
```
{[workload_config_1, workload_config_2, ...]}
```

#### workload_config is a json in one of the following format:  
###### workload_config using HTTP 
```
{  
    "name" : "workload_name", // must be unique name
    "allowed_status" : ["200", "404"], // must contain all allowed statuses as strings (panic on non-allowed status)  
    "start_time": 0, // the time in seconds from the run start in which the workload will start its run. if depends_on_worklaod is set then the workload will only start after start_time and the workloads it depenends on finished.  
    "duration" : 10, // number of seconds the workload will run. if end_on_var_value is set its possible that the workload will end before the duration time but not after it  
    "depends_on_workload" : ["workload_2_name", "workload_3_name", ...], // set workloads that this workload depends on - will only start running after them  
    "end_on_var_value" : {"<var_1_name>" : <config_field_1>, <var_2_name>" : <config_field_2>, ...}, // define conditions for the worker to finish its work before the duration end based on var values. check config_field_json format below  
    "workers": 32, // number of workers the workload will use  
    "vars" : <vars_config>, // define variables to be used by the workload/workers. check vars_config format below  
    "type" : "HTTP", // define workload type
    "http_config" : <http_config> // set the http_config part - only used if type="HTTP". see http_config format below  
}  
```
###### workload_config using SHELL 
```
{  
    "name" : "workload_name", // must be unique name
    "allowed_status" : ["0", "2"], // must contain all allowed statuses as strings (panic on non-allowed status)  
    "start_time": 0, // the time in seconds from the run start in which the workload will start its run. if depends_on_worklaod is set then the workload will only start after start_time and the workloads it depenends on finished.  
    "duration" : 10, // number of seconds the workload will run. if end_on_var_value is set its possible that the workload will end before the duration time but not after it  
    "depends_on_workload" : ["workload_2_name", "workload_3_name", ...], // set workloads that this workload depends on - will only start running after them  
    "end_on_var_value" : {"<var_1_name>" : <config_field_1>, <var_2_name>" : <config_field_2>, ...}, // define conditions for the worker to finish its work before the duration end based on var values. check config_field_json format below  
    "workers": 32, // number of workers the workload will use  
    "vars" : <vars_config>, // define variables to be used by the workload/workers. check vars_config format below  
    "type" : "SHELL", // define workload type
    "shell_config" : <shell_config> // set the shell_config part - only used if type="SHELL". see shell_config format below 
}  
```

#### config_field should be a json in one of the following formats:  
###### config_field using CONST values  
```
{ 
    "type" : "CONST",  
    "value" : <value>, // can be string or number (not tested with complex structs)  
    "op" : "==", // used only for config_field in end_on_var_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}  
```
  
###### config_field using current var value (easier and faster then using format if all you want is a var value)
```
{  
    "type" : "VAR",  
    "var_name" : "var_1_name", // will use the var current value
    "op" : "==", // used only for config_field in end_on_var_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}  
```

###### config_field using a format based on current var values 
```
{  
    "type" : "FORMAT",  
    "format" : "dir_%s/file_%d", // the format to be used (golang format - can also use %v)
    "args" : ["var_1_name", "var_2_name"], // var names used as args for the format
    "op" : "==", // used only for config_field in end_on_var_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}  
```

#### vars_config should be a json with the following format:  
```
{ 
    /* 
       notice blaster also have builtin vars that can be used
       io_blaster_uid - unique id for each request
       io_blaster_worker_id - worker id starting from 0
    */
    "const" : // used to define const values 
    {
        "var_1_name" : 
        { 
            "value" : <value> // can be string or number (not tested with complex structs)
        },
        "var_2_name" : 
        { 
            "value" : <value> // can be string or number (not tested with complex structs)
        },
    }
    "file" :
    {
        "var_3_name" : 
        { 
            "path" : "/tmp/payload.txt" // path to file containing the data (data is parsed as string data)
        }
    }
    "random" : // can generate random int/string values (charset for random strings is a-z, A-Z, 0-9)
    {
        "once" : // workload will run random once and all workers will use same value
        {
            "var_4_name" : // example of random string
            {
                "type" : "STRING",
                "length" : 100 // will generate random string with length=100               
            }
        },
        "worker_once" : // each worker will run random once so each worker will have different random values
        {
            "var_5_name" : // example of random int
            {
                "type": "INT",
                "min_value" : 0, // random value min value (inclusive)
                "max_value" : 100 // random value max value (inclusive)
            }
        },
        "each" : // workers will calculate the var value with new random for each request
        {
            "var_6_name" :
            {
                "type" : "STRING",
                "length" : 100 // will generate random string with length=100
            }
        }
    }
    "enum" : 
    {
        "workload_sim_each" : // simulate 1 enum per workload var (only simulate so no thread sync is needed). each worker starts from min_value+worker_index (instead of just min_value) and worker increases the var value by workers number instead of by 1 for each request
        {
            "var_7_name" :
            {
                "min_value" : 0 // starting value of the enum
            }
        },
        "worker_each" : // each worker will run its own internal enum. worker increases the its var value by 1 for each request
        {
            "var_8_name" :
            {
                "min_value" : 10 // starting value of the enum
            }
        },
        "on_time" : // will increase var value only on interval (can be used to sync var increase between workloads)
        {
            "var_9_name" :
            {
                "min_value" : 0, // starting value of the enum
                "interval": 1 // interval in seconds in which the enum will run the ++ (in the example each 1 sec the var will increase its value by 1)
            }
        }
    }
    "response_value" : // vars with value taken from the response
    {
        "var_10_name" :
        {
            "update_on_status" : ["200"], // will only parse the value on responses with this statuses.
            "field_path" : ["Records", "0", "Data"], // json field path needed to get the value from the response. empty array will use response data as blob. in this example the var value will be taken from response json field Records[0].Data (see https://godoc.org/github.com/jmoiron/jsonq for json field path info)
            "init_value" : <value>, // can be string or number (not tested with complex structs)  
            "expected_values" : [<config_field_1>, <config_field_2>] // list of expected values - panic on none expected value
        }
    }
}
```

#### http_config should be a json with the following format:  
```
{
    "method" : <config_field>, // set the method string using config_field
    "url" : <config_field>, // set the url string using config_field
    "headers" : // set the headers with mapping of header names to values generated by config_fields 
    {
        "header_1_name" : <config_field>,
        "header_2_name" : <config_field>
    },
    "body" : <config_field>, // set the body string using config_field
}
```

#### shell_config should be a json with the following format:  
```
{
    "user" : <config_field>, // set the user string using config_field (user used to ssh)
    "password" : <config_field>, // set the password string using config_field (password used to ssh - if not set will only work if current machine io_blaster is running from is in authorized hosts of target host)
    "host" : <config_field>, // set the host string using config_field (host to ssh to)
    "cmd" : <config_field> // set the cmd string using config_field (shell command to run)
}
```

For more accurate understanding of the config file format check some of the [test files](https://github.com/iguazio/io_blaster/tree/master/test_files).

## planned features
* add support for array vars - support for const number/string array vars
* add support for dist var - while random.worker_once/random.each can be used to distribute values between workers it might create a bit unbalanced distribution. dist var will be able to iterate in a cyclic way over an array of values to distribute them between the workers
* improve enum vars to support inc/dec by some const value instead of just inc by 1
* add support for global const vars - will allow to define const/file/random.global_once vars in 1 global vars area so wont need to redefine for each workload.
* add support for arrayed_format - with the support for array vars it would be usefull to be able to create a format based on array without having to reference each element of the array. for example if you have an array with 100 elements and you want your http body to contain a json with a record for each array element you will be able to to do it in 1 line arrayed_format instead of regular format referencing the 100 elements.
* add support for arrayed response_value vars - same idea as in arrayed_format. this will help if you have a response containing an array and you want to load,verify the whole array
* add support for capnp response parsing in a similar way to the current json response parsing
