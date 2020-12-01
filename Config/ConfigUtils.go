package Config

import (
	"encoding/json"
	"fmt"
	"github.com/iguazio/io_blaster/Utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
)

func ParseField(fieldConfig *ConfigField, calculatedVars CalculatedVars, workloadName string) interface{} {
	if fieldConfig == nil {
		return ""
	}
	switch fieldConfig.Type {
	case "FORMAT":
		args := make([]interface{}, 0)
		for _, argName := range fieldConfig.FormatArgs {
			args = append(args, calculatedVars[argName.(string)])
		}
		return fmt.Sprintf(fieldConfig.Format, args...)
	case "ARRAY_FORMAT":
		arrayIndexesInArgs := make([]int, 0)
		args := make([]interface{}, 0)
		for argIndex, argName := range fieldConfig.FormatArgs {
			if _, ok := argName.(string); ok {
				args = append(args, calculatedVars[argName.(string)])
				for _, arrayArgsValue := range fieldConfig.ArrayArgs {
					if argIndex == arrayArgsValue {
						arrayIndexesInArgs = append(arrayIndexesInArgs, argIndex)
					}
				}
			} else {
				isArray := false
				for _, arrayArgsValue := range fieldConfig.ArrayArgs {
					if argIndex == arrayArgsValue {
						arrayIndexesInArgs = append(arrayIndexesInArgs, argIndex)
						isArray = true
					}
				}
				if !isArray {
					log.Panicln(fmt.Sprintf("Workload %s found array_format field with non-string arg that is not an array. field=%+v", workloadName, fieldConfig))
				}

				argArrayLen := int64(argName.(float64))
				argArray := make([]interface{}, argArrayLen)
				for argArrayIndex := range argArray {
					argArray[argArrayIndex] = argArrayIndex
				}
				args = append(args, argArray)
			}
		}
		return Utils.ArrayFormat(fieldConfig.Format, args, arrayIndexesInArgs, fieldConfig.ArrayJoinString)
	case "CONST":
		return fieldConfig.Value
	case "VAR":
		return calculatedVars[fieldConfig.VarName]
	default:
		log.Panicln(fmt.Sprintf("Workload %s found field with unsupported type. field=%+v", workloadName, fieldConfig))
		break
	}
	return nil
}

func VarRunRandom(varConfig *ConfigVarsRandomOrEnum) interface{} {
	switch varConfig.Type {
	case "STRING":
		if varConfig.Length == 0 {
			log.Panicln(fmt.Sprintf("Found random string var with legnth=0. var=%+v", varConfig))
		}
		return Utils.GenerateRandomString(varConfig.Length)
	case "BASE64":
		if varConfig.Length == 0 {
			log.Panicln(fmt.Sprintf("Found random base64 var with legnth=0. var=%+v", varConfig))
		}
		return Utils.GenerateRandomBase64(varConfig.Length)
	case "INT":
		if varConfig.MaxValue <= varConfig.MinValue {
			log.Panicln(fmt.Sprintf("Found random int var with max_value <= min_value. var=%+v", varConfig))
		}
		var seededRand *rand.Rand = Utils.GetSeededRandom()
		return seededRand.Int63n(varConfig.MaxValue-varConfig.MinValue+1) + varConfig.MinValue
	default:
		log.Panicln(fmt.Sprintf("Found random var with unsupported type. var=%+v", varConfig))
	}

	return nil
}

func (calculatedVars CalculatedVars) CalculatedRandomVarsConfig(workloadName string, configVarsRandomOrEnumMap ConfigVarsRandomOrEnumMap, assertExist bool) {
	for varName, varConfig := range configVarsRandomOrEnumMap {
		if assertExist {
			if _, ok := calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", workloadName, varName))
			}
		}
		calculatedVars[varName] = VarRunRandom(varConfig)
		calculatedVars.RunTriggers(varConfig.Triggers, varName, workloadName)
	}
}

func (calculatedVars CalculatedVars) RunTriggers(triggers []*ConfigVarsTrigger, varName string, workloadName string) {
	varValue := calculatedVars[varName]
	for _, triggerConfig := range triggers {
		if compare_res, err := Utils.CompareInterface(triggerConfig.OnValue.Op, varValue, ParseField(triggerConfig.OnValue, calculatedVars, workloadName)); err != nil {
			log.Panicln(fmt.Sprintf("Workload %s found trigger config with unsupported op or missmatched value types. config=%s", workloadName, triggerConfig))
		} else if compare_res {
			calculatedVars[triggerConfig.VarToSet] = ParseField(triggerConfig.ValueToSet, calculatedVars, workloadName)
		}
	}
}

func (config *ConfigIoBlaster) LoadConfig(config_file_path string) {
	json_file, err := os.Open(config_file_path)
	if err != nil {
		log.Panicln("Failed to open config file")
	}
	defer json_file.Close()

	byteValue, _ := ioutil.ReadAll(json_file)
	err = json.Unmarshal(byteValue, config)
	if err != nil {
		log.Panicln("Failed to parse config file json", err)
	}
}
