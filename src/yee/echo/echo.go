package echo

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"yee/config"
)

type LogModel struct {
	Template string
	Current  string
	File     *os.File
}

var tempFile string
var ColorMap map[string]string
var infoLog *LogModel
var errorLog *LogModel

func init() {
	ColorMap = make(map[string]string)
	ColorMap["file"] = "\033[0;33m%s\033[0m"
	ColorMap["sql"] = "\033[0;34m%s\033[0m"
	ColorMap["time"] = "\033[0;37m%s\033[0m"
	ColorMap["error"] = "\033[31;31m%s\033[0m"
	ColorMap["info"] = "\033[33;37m%s\033[0m"
	infoLog = &LogModel{
		Template: config.String("log_info_file", ""),
		Current:  "",
		File:     nil,
	}
	errorLog = &LogModel{
		Template: config.String("log_error_file", ""),
		Current:  "",
		File:     nil,
	}
}

func (lm *LogModel) write(str string) {
	if lm.Template == "" {
		return
	}
	dateTime := time.Now().In(config.CstZone())
	nowDate := dateTime.Format("2006_01_02")
	logFile := strings.ReplaceAll(lm.Template, "{date}", nowDate)
	//log.Printf(logFile)
	if lm.Current != logFile {
		if lm.File != nil {
			lm.File.Close()
			lm.File = nil
		}
		lm.Current = logFile
		stdout, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			lm.File = stdout
		}
	}
	if lm.File != nil {
		nowTime := dateTime.Format("15:04:05")
		lm.File.WriteString(nowTime + " " + str)
	}
}

func outer(color, file, text string, sTime string, runType int) {
	//客户端和服务端都输出
	if runType == 1 || runType == 2 {
		format := "%s"
		color = strings.ToLower(color)
		if _, ok := ColorMap[color]; ok {
			format = ColorMap[color]
		}
		if file != "" {
			log.Printf("\033[0;33m%s\033[0m", "> "+file)
		}
		if sTime != "" {
			format += "\033[0;37m%s\033[0m"
			log.Printf(format, text, "   ----time: "+sTime)
		} else {
			log.Printf(format, text)
		}
	}
	//检查写入日志
	if runType == 1 || runType == 3 {
		if color == "error" {
			if file != "" {
				errorLog.write("> " + file + "\n")
			}
			if sTime != "" {
				errorLog.write(text + "\ntime:" + sTime + "\n")
			} else {
				errorLog.write(text + "\n")
			}
		} else {
			if file != "" {
				infoLog.write("> " + file + "\n")
			}
			if sTime != "" {
				infoLog.write(text + "\ntime:" + sTime + "\n")
			} else {
				infoLog.write(text + "\n")
			}
		}
	}
}

func Print(bytes []byte, runType int) {
	if bytes[0] == '{' && bytes[len(bytes)-1] == '}' && json.Valid(bytes) {
		var tempMap map[string]interface{}
		err := json.Unmarshal(bytes, &tempMap)
		if err != nil {
			outer("info", "", string(bytes), "", runType)
			return
		}
		act := "log"
		data := make([]interface{}, 0)
		sTime := ""
		file := ""
		if _, ok := tempMap["act"]; ok {
			switch tempMap["act"].(type) {
			case string:
				act = tempMap["act"].(string)
				break
			}
		} else {
			outer("info", "", string(bytes), "", runType)
			return
		}
		if _, ok := tempMap["data"]; ok {
			switch tempMap["data"].(type) {
			case string, int, int64, float64, float32, bool:
				data = append(data, tempMap["data"])
				break
			case []interface{}:
				data = tempMap["data"].([]interface{})
				break
			}
		}
		if _, ok := tempMap["time"]; ok {
			switch tempMap["time"].(type) {
			case string:
				sTime = tempMap["time"].(string)
				break
			case int, int64:
				sTime = strconv.Itoa(tempMap["time"].(int))
				break
			case float64, float32:
				sTime = strconv.FormatFloat(tempMap["time"].(float64), 'g', 10, 64)
				break
			}
		}
		if _, ok := tempMap["file"]; ok {
			tFile := ""
			switch tempMap["file"].(type) {
			case string:
				tFile = tempMap["file"].(string)
			}
			if tFile != "" && tempFile != tFile {
				file = tFile
				tempFile = file
			}
		}
		if len(data) > 0 {
			temps := make([]string, 0)
			for _, it := range data {
				switch it.(type) {
				case string:
					temps = append(temps, it.(string))
					break
				case int:
					temps = append(temps, strconv.Itoa(it.(int)))
					break
				case int64:
					temps = append(temps, strconv.Itoa(int(it.(int64))))
					break
				case float64, float32:
					s := strconv.FormatFloat(it.(float64), 'g', 10, 64)
					temps = append(temps, s)
					break
				case bool:
					if it.(bool) {
						temps = append(temps, "true")
					} else {
						temps = append(temps, "false")
					}
					break
				default:
					break
				}
			}
			outer(act, file, strings.Join(temps, "   "), sTime, runType)
		}
	}
}
