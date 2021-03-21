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
* You can add -s <_stats_File_Path_> to dump workload, workers stats json to the given file path at the end of the run.
* Notice that in order to run shell workloads with more than 10 workers you will need to edit MaxStartups param in sshd_config to support for more than 10 connections. To allow 100 connections run `sudo vi /etc/ssh/sshd_config` and add at the end `MaxStartups 100:30:200`, then restart the sshd for the changes to take effect by running `sudo service sshd restart`

### config format

#### Config file should be a json with the following format:  
```
{
    "vars" : <vars_config>, // define global variables to be used by the workloads/workerss. check vars_config format below. only vars that are calculated once can be used here (const, file, random.once, random.array_once, enum.array_once)  
    "workloads" : [workload_config_1, workload_config_2, ...],       
}
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
    "op" : "==", // used only for config_field in end_on_var_value/trigger_config.on_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}  
```
  
###### config_field using current var value (easier and faster then using format if all you want is a var value)
```
{  
    "type" : "VAR",  
    "var_name" : "var_1_name", // will use the var current value
    "op" : "==", // used only for config_field in end_on_var_value/trigger_config.on_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}  
```

###### config_field using a format based on current var values 
```
{  
    "type" : "FORMAT",  
    "format" : "dir_%s/file_%d", // the format to be used (golang format - can also use %v)    
    "args" : ["var_1_name", "var_2_name"], // var names used as args for the format
    "op" : "==", // used only for config_field in end_on_var_value/trigger_config.on_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}
```
  
###### config_field using array_format based on current var/arrays values (notice an example of how the array_format works can be found [here](https://play.golang.org/p/QrTkCE-CPOS))
```
{ 	
    "type" : "ARRAY_FORMAT",  
    "format" : "dir_%s/file_%d", // the format to be used (golang format - can also use %v)
    "array_args" : [<array_arg_1_index>], // vars names must be names of vars containing arrays. 
    "array_join_string" : ",", // will be used to join array_format parts
    "args" : ["var_1_name", "var_2_name"], // var names used as args for the format. can also contain number for range len. for example if array_args=[0] and args=[5] then it will auto generate array [0,1,2,3,4] in index 0 of the args
    "op" : "==", // used only for config_field in end_on_var_value/trigger_config.on_value. set the op for the comparison. can be "==", ">", "<", ">=", "<="
}
```

#### trigger_config should be a json with the following format:
```
{
	"on_value" : <config_field>, // will be used to compare the value based on the op to the value of the var that is being triggered 
	"var_to_set" : "var_1_name", // var to set if the trigger value comparison passed
	"value_to_set" : <config_field> // value to set on the var
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
    /*
       notice that variable parsing order is as follow:
       1. const, file, random.once, random.worker_once, response_value (init value)
       2. random.each, enum.workload_sim_each, enum.worker_each, enum.on_time
       3. dist
       4. config_field 
       5. response_value (after the resposne on the request)
    */
    /*
       notice about triggers:
       1. triggers are optional (can omit the field from the var config)
       2. triggers are only checked when the value changes as part of the var config. this means that const var trigger will only trigger once on the first setting of the value, enum var trigger will only trigger when the enum function of the var increase its value. in other words the triggers are not checked when value is changed by other trigger.
       3. if a var defines several triggers then the var value before the triggers checking will be used in all the triggers on_value comparisons.
       4. triggers can be used to define new vars. see example in https://github.com/iguazio/io_blaster/tree/master/test_files/test_trigger.json
    */
    "const" : // used to define const values 
    {
        "var_1_name" : 
        { 
            "value" : <value> // can be string or number (not tested with complex structs)
            "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
        },
        "var_2_name" : 
        { 
            "value" : <value> // can be string or number (not tested with complex structs)
            "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
        },
    },
    "file" :
    {
        "var_3_name" : 
        { 
            "path" : "/tmp/payload.txt" // path to file containing the data (data is parsed as string data)
            "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
        }
    },
    "random" : // can generate random int/string values (charset for random strings is a-z, A-Z, 0-9)
    {
        "once" : // workload will run random once and all workers will use same value
        {
            "var_4_name" : // example of random string
            {
                "type" : "STRING",
                "length" : 100 // will generate random string with length=100
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
            }
        },
        "worker_once" : // each worker will run random once so each worker will have different random values
        {
            "var_5_name" : // example of random int
            {
                "type": "INT",
                "min_value" : 0, // random value min value (inclusive)
                "max_value" : 100 // random value max value (inclusive)
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
            }
        },
        "each" : // workers will calculate the var value with new random for each request
        {
            "var_6_name" :
            {
                "type" : "BASE64",
                "length" : 100 // will generate random base64 string with original blob length=100
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
            }
        }
        "array_once" : // workload/global will generate an array with <array_length> length and will then generate random data based on the params for each array index
        {
        	"var_7_name" : 
        	{
        		"type" : "STRING",
                "length" : 100 // will generate random string with length=100
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
                "array_length" : 4, // used only in array_once vars to set the size of the array to be generated
        	}
        }
    },
    "enum" : 
    {
        "workload_sim_each" : // simulate 1 enum per workload var (only simulate so no thread sync is needed). each worker starts from min_value+worker_index (instead of just min_value) and worker increases the var value by workers number instead of by 1 for each request
        {
            "var_8_name" :
            {
                "min_value" : 0 // starting value of the enum
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
            }
        },
        "worker_each" : // each worker will run its own internal enum. worker increases the its var value by 1 for each request
        {
            "var_9_name" :
            {
                "min_value" : 10 // starting value of the enum
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
            }
        },
        "on_time" : // will increase var value only on interval (can be used to sync var increase between workloads)
        {
            "var_10_name" :
            {
                "min_value" : 0, // starting value of the enum
                "interval": 1 // interval in seconds in which the enum will run the ++ (in the example each 1 sec the var will increase its value by 1)
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
            }
        },
        "array_once" : // workload/global will generate an array with <array_length> length and will then enumarate values into the array starting from <min_value>
        {
        	"var_11_name" : 
        	{
        		"min_value" : 0 // starting value of the enum
                "triggers" : [<trigger_config_1>, <trigger_config_2>, ...], // define triggers on the var
                "array_length" : 4, // used only in array_once vars to set the size of the array to be generated
        	}
        }
    },
    "response_value" : // vars with value taken from the response
    {
        "var_12_name" :
        {
            "update_on_status" : ["200"], // will only parse the value on responses with this statuses.
            "field_path" : ["Records", "0", "Data"], // json field path needed to get the value from the response. empty array will use response data as blob. in this example the var value will be taken from response json field Records[0].Data (see https://godoc.org/github.com/jmoiron/jsonq for json field path info)
            "init_value" : <value>, // can be string or number (not tested with complex structs)  
            "expected_values" : [<config_field_1>, <config_field_2>], // list of expected values - panic on none expected value
            "expected_values_array_vars" : [var_11_name], // list of var names containing the expected values
            "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
        }
    },
    "config_field"  : // var with value calculated like a config_field
    {
    	"var_13_name" : <config_field> // set the var value using config_field   	
       "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
    },
    "dist" : // distribute values from an array var to the workers in a cyclic way. the value will be based on worker_index % array_len 
    {
    	"var_14_name" :
    	{
           "array_var" : <array_var_name> // must be name of var containg an array
           "triggers" : [<trigger_config_1>, <trigger_config_2>, ...] // define triggers on the var
    	}
    }
}
```

#### http_config should be a json with the following format:  
```
{
    "request_timeout" : 120, // set request timeout in seconds (default value is 120)
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
* improve enum vars to support inc/dec by some const value instead of just inc by 1
* add support for arrayed response_value vars - same idea as in arrayed_format. this will help if you have a response containing an array and you want to load,verify the whole array
* add option to limit request/response payload output length when logging them on error or when running in debug mode (to not spam the logs)
* add default log path when none is given (<config_path>.log)
* add support for capnp response parsing in a similar way to the current json response parsing
