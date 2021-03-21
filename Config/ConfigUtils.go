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

func ParseField(fieldConfig *ConfigField, calculatedVars CalculatedVars, varsConfigLogPrefixString string) interface{} {
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
					log.Panicln(fmt.Sprintf("%s found array_format field with non-string arg that is not an array. field=%+v", varsConfigLogPrefixString, fieldConfig))
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
		log.Panicln(fmt.Sprintf("%s found field with unsupported type. field=%+v", varsConfigLogPrefixString, fieldConfig))
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

func (calculatedVars CalculatedVars) CalculatedRandomVarsConfig(varsConfigLogPrefixString string, configVarsRandomOrEnumMap ConfigVarsRandomOrEnumMap, assertExist bool) {
	for varName, varConfig := range configVarsRandomOrEnumMap {
		if assertExist {
			if _, ok := calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("%s contain 2 vars with same name %s", varsConfigLogPrefixString, varName))
			}
		}
		calculatedVars[varName] = VarRunRandom(varConfig)
		calculatedVars.RunTriggers(varConfig.Triggers, varName, varsConfigLogPrefixString)
	}
}

func (calculatedVars CalculatedVars) RunTriggers(triggers []*ConfigVarsTrigger, varName string, varsConfigLogPrefixString string) {
	varValue := calculatedVars[varName]
	for _, triggerConfig := range triggers {
		if compare_res, err := Utils.CompareInterface(triggerConfig.OnValue.Op, varValue, ParseField(triggerConfig.OnValue, calculatedVars, varsConfigLogPrefixString)); err != nil {
			log.Panicln(fmt.Sprintf("%s found trigger config with unsupported op or missmatched value types. config=%+v", varsConfigLogPrefixString, triggerConfig))
		} else if compare_res {
			calculatedVars[triggerConfig.VarToSet] = ParseField(triggerConfig.ValueToSet, calculatedVars, varsConfigLogPrefixString)
		}
	}
}

func (calculatedVars CalculatedVars) CalculateConstVars(varsConfigLogPrefixString string, vars *ConfigVars) {
	if vars == nil {
		return
	}

	for varName, varConfig := range vars.Const {
		if _, ok := calculatedVars[varName]; ok {
			log.Panicln(fmt.Sprintf("%s contain 2 vars with same name %s", varsConfigLogPrefixString, varName))
		}
		calculatedVars[varName] = varConfig.Value
		calculatedVars.RunTriggers(varConfig.Triggers, varName, varsConfigLogPrefixString)
	}

	for varName, varConfig := range vars.File {
		if _, ok := calculatedVars[varName]; ok {
			log.Panicln(fmt.Sprintf("%s contain 2 vars with same name %s", varsConfigLogPrefixString, varName))
		}
		file, err := os.Open(varConfig.Path)
		if err != nil {
			log.Panicln(fmt.Sprintf("Failed to open file %s from var %s", varConfig.Path, varName))
		}
		defer file.Close()

		byteValue, _ := ioutil.ReadAll(file)
		calculatedVars[varName] = string(byteValue)
		calculatedVars.RunTriggers(varConfig.Triggers, varName, varsConfigLogPrefixString)
	}

	if vars.Random != nil {
		calculatedVars.CalculatedRandomVarsConfig(varsConfigLogPrefixString, vars.Random.Once, true)

		for varName, varConfig := range vars.Random.ArrayOnce {
			if _, ok := calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("%s contain 2 vars with same name %s", varsConfigLogPrefixString, varName))
			}
			if varConfig.ArrayLength <= 0 {
				log.Panicln(fmt.Sprintf("%s contain random.array_once var with bad array_length. var name %s", varsConfigLogPrefixString, varName))
			}
			calculatedVars[varName] = make([]interface{}, varConfig.ArrayLength)
			for i := int64(0); i < varConfig.ArrayLength; i++ {
				calculatedVars[varName].([]interface{})[i] = VarRunRandom(varConfig)
			}
			calculatedVars.RunTriggers(varConfig.Triggers, varName, varsConfigLogPrefixString)
		}
	}

	if vars.Enum != nil {
		for varName, varConfig := range vars.Enum.ArrayOnce {
			if _, ok := calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", varsConfigLogPrefixString, varName))
			}
			if varConfig.ArrayLength <= 0 {
				log.Panicln(fmt.Sprintf("%s contain enum.array_once var with bad array_length. var name %s", varsConfigLogPrefixString, varName))
			}
			calculatedVars[varName] = make([]interface{}, varConfig.ArrayLength)
			for i := int64(0); i < varConfig.ArrayLength; i++ {
				calculatedVars[varName].([]interface{})[i] = varConfig.MinValue + i
			}
			calculatedVars.RunTriggers(varConfig.Triggers, varName, varsConfigLogPrefixString)
			for i := int64(0); i < varConfig.ArrayLength; i++ {
				if _, ok := calculatedVars[varName].([]interface{})[i].(float64); ok {
					calculatedVars[varName].([]interface{})[i] = int64(calculatedVars[varName].([]interface{})[i].(float64))
				}
			}
		}
	}
}

func (configWorkload *ConfigWorkload) GetVarsConfigLogPrefix() string {
	return fmt.Sprintf("Workload %s vars", configWorkload.Name)
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
