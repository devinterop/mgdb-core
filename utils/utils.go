package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	oid "github.com/coolbed/mgo-oid"
	//"github.com/joho/godotenv"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/devinterop/mgdb-core/packages/logging"
)

// const charset = "abcdefghijklmnopqrstuvwxyz" +
//
//	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const charset = "0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

var ViperYaml *viper.Viper
var ViperJson *viper.Viper

var logrusFieldUtils = structs.LogrusField{
	Module: "Utils",
}

func SetViperYaml(v *viper.Viper) {
	//log.Printf("SetViperYaml: %s", "OK")
	ViperYaml = v
}

func SetViperJson(v *viper.Viper) {
	ViperJson = v
}

func GetViperJson() *viper.Viper {
	return ViperJson
}

// https://stackoverflow.com/questions/25349694/go-encoding-json-from-relfect-value
func ConvertStructToJson(s interface{}) []byte {
	logrusField := logrusFieldUtils
	logrusField.Method = "ConvertStructToJson"

	r := reflect.ValueOf(&s)
	f := reflect.Indirect(r).Elem()
	j, err := json.Marshal(f.Interface())
	if err != nil {
		//fmt.Println("error")
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	//logging.Logger(cnst.Debug, fmt.Sprintf("%+v => %s", s, string(j)), logrusField)
	return j
}

func GetStructField(v *interface{}, field string) reflect.Value {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).Elem().FieldByName(field)
	return f
}

func GetDataFromReflectValue(value reflect.Value, dataType string) interface{} {
	if dataType == "array" {
		var result = []interface{}{}
		len := value.Len()
		for i := 0; i < len; i++ {
			result = append(result, value.Index(i).Interface())
		}
		return result
	} else if dataType == "struct" {
		return value.Interface()
	} else if dataType == "int" {
		return value.Int()
	} else if dataType == "string" {
		return value.String()
	} else if dataType == "bool" {
		return value.Bool()
	} else if dataType == "float" {
		return value.Float()
	}

	return nil
}

func GetEnv(key string) string {
	return getEnvVariable(key)
}
func GetEnvBool(key string) bool {
	return getEnvVariableBool(key)
}
func GetYaml(key string) string {
	return getViperVariable(key)
}
func GetYamlDuration(key string) time.Duration {
	return getViperVariableDurationSec(key)
}
func GetYamlInt(key string) int {
	return getViperVariableInt(key)
}
func GetYamlBool(key string) bool {
	return getViperVariableBool(key)
}

func getEnvVariable(key string) string {
	return os.Getenv(key)
}

// func getVariableBool(key string) bool {
// 	logrusField := logrusFieldUtils
// 	logrusField.Method = "getVariableBool"

// 	rbool, err := strconv.ParseBool(key)
// 	if err != nil {
// 		//log.Fatal(err)
// 		logging.Logger(cnst.Fatal, err, logrusField)
//		return false
// 	}
// 	return rbool
// }

func getEnvVariableBool(key string) bool {
	logrusField := logrusFieldUtils
	logrusField.Method = "getEnvVariableBool"

	result := os.Getenv(key)
	boolValue, err := strconv.ParseBool(result)
	if err != nil {
		//log.Fatal(err)
		logging.Logger(cnst.Fatal, err, logrusField)
		return false
	}
	return boolValue

}

func getViperVariable(key string) string {
	value := ViperYaml.GetString(key)

	return value
}

func getViperVariableDurationSec(key string) time.Duration {
	value := ViperYaml.GetDuration(key)

	return value * time.Second
}

func getViperVariableInt(key string) int {
	value := ViperYaml.GetInt(key)

	return value
}

func getViperVariableBool(key string) bool {
	value := ViperYaml.GetBool(key)

	return value
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomCode(length int) string {
	return StringWithCharset(length, charset)
}

// GenerateID function is for Generate Document id
func GenerateID(prefix string) string {
	objectID := oid.NewOID()
	// fmt.Println("object id:", objectID.String())
	// fmt.Println("object timestamp", objectID.Timestamp())
	return prefix + objectID.String()
}

func GenerateUUID() string {
	objectID := uuid.New()
	// fmt.Println("object id:", objectID.String())
	// fmt.Println("object timestamp", objectID.Timestamp())
	return objectID.String()
}

type StrSlice struct {
	Str []string
}

// convertToSlice is to convert string to array
func ConvertToSlice(s string) []string {
	logrusField := logrusFieldUtils
	logrusField.Method = "ConvertToSlice"

	var str []string
	err := json.Unmarshal([]byte(s), &str)
	if err != nil {
		//log.Fatal(err)
		//logging.Logger(cnst.Fatal, err)
		//panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
		return nil
	}
	return str
}

// ConvertOperators is to convert Query Operators
func ConvertOperators(data interface{}) interface{} {
	if reflect.ValueOf(data).Kind() == reflect.Slice {
		d := reflect.ValueOf(data)
		tmpData := make([]interface{}, d.Len())
		returnSlice := make([]interface{}, d.Len())
		for i := 0; i < d.Len(); i++ {
			tmpData[i] = d.Index(i).Interface()
		}
		for i, v := range tmpData {
			returnSlice[i] = ConvertOperators(v)
		}
		return returnSlice
	} else if reflect.ValueOf(data).Kind() == reflect.Map {
		d := reflect.ValueOf(data)
		tmpData := make(map[string]interface{})
		for _, k := range d.MapKeys() {
			if d.MapIndex(k).Interface() != nil {
				typeOfValue := reflect.TypeOf(d.MapIndex(k).Interface()).Kind()
				if typeOfValue == reflect.Map || typeOfValue == reflect.Slice {
					tmpData[MapOperators(k.String())] = ConvertOperators(d.MapIndex(k).Interface())
				} else {
					tmpData[MapOperators(k.String())] = d.MapIndex(k).Interface()
				}
			} else {
				fmt.Println("object id:", MapOperators(k.String()))
				tmpData[MapOperators(k.String())] = nil
			}
		}
		return tmpData
	}
	return data
}

// MapOperators for set Query Operators
func MapOperators(str string) string {
	switch str {
	//---MongoDB Query Operators---//
	//comparison
	case "eq":
		return "$eq"
	case "gt":
		return "$gt"
	case "gte":
		return "$gte"
	case "in":
		return "$in"
	case "lt":
		return "$lt"
	case "lte":
		return "$lte"
	case "ne":
		return "$ne"
	case "nin":
		return "$nin"
	//logical
	case "and":
		return "$and"
	case "not":
		return "$not"
	case "nor":
		return "$nor"
	case "or":
		return "$or"
	//element
	case "exists":
		return "$exists"
	case "type":
		return "$type"
	//evaluation
	case "expr":
		return "$expr"
	case "jsonSchema":
		return "$jsonSchema"
	case "mod":
		return "$mod"
	case "regex":
		return "$regex"
	case "regexNo":
		return "regex"
	case "text":
		return "$text"
	case "where":
		return "$where"
	//geospatial
	case "geoIntersects":
		return "$geoIntersects"
	case "geoWithin":
		return "$geoWithin"
	case "near":
		return "$near"
	case "nearSphere":
		return "$nearSphere"
	//array
	case "all":
		return "$all"
	case "elemMatch":
		return "$elemMatch"
	case "size":
		return "$size"
	//bitwise
	case "bitsAllClear":
		return "$bitsAllClear"
	case "bitsAllSet":
		return "$bitsAllSet"
	case "bitsAnyClear":
		return "$bitsAnyClear"
	case "bitsAnySet":
		return "$bitsAnySet"

	//---MongoDB -Index of Expression Operators---//
	case "abs":
		return "$abs"
	case "accumulator":
		return "$accumulator"
	case "acos":
		return "$acos"
	case "acosh":
		return "$acosh"
	case "add":
		return "$add"
	case "addToSet":
		return "$addToSet"
	case "allElementsTrue":
		return "$allElementsTrue"
	// case "and":
	// 	return "$and"
	case "anyElementTrue":
		return "$anyElementTrue"
	case "arrayElemAt":
		return "$arrayElemAt"
	case "arrayToObject":
		return "$arrayToObject"
	case "asin":
		return "$asin"
	case "asinh":
		return "$asinh"
	case "atan":
		return "$atan"
	case "atan2":
		return "$atan2"
	case "atanh":
		return "$atanh"
	case "avg":
		return "$avg"
	case "binarySize":
		return "$binarySize"
	case "bsonSize":
		return "$bsonSize"
	case "ceil":
		return "$ceil"
	case "cmp":
		return "$cmp"
	case "concat":
		return "$concat"
	case "concatArrays":
		return "$concatArrays"
	case "cond":
		return "$cond"
	case "convert":
		return "$convert"
	case "cos":
		return "$cos"
	case "dateFromParts":
		return "$dateFromParts"
	case "dateFromString":
		return "$dateFromString"
	case "dateToParts":
		return "$dateToParts"
	case "dateToString":
		return "$dateToString"
	case "dayOfMonth":
		return "$dayOfMonth"
	case "dayOfWeek":
		return "$dayOfWeek"
	case "dayOfYear":
		return "$dayOfYear"
	case "degreesToRadians":
		return "$degreesToRadians"
	case "divide":
		return "$divide"
	// case "eq":
	// 	return "$eq"
	case "exp":
		return "$exp"
	case "filter":
		return "$filter"
	// case "first":
	// 	return "$first"  (array)
	// case "first":
	// 	return "$first"  (accumulator)
	case "floor":
		return "$floor"
	case "function":
		return "$function"
	// case "gt":
	// 	return "$gt"
	// case "gte":
	// 	return "$gte"
	case "hour":
		return "$hour"
	case "ifNull":
		return "$ifNull"
	// case "in":
	// 	return "$in"
	case "indexOfArray":
		return "$indexOfArray"
	case "indexOfBytes":
		return "$indexOfBytes"
	case "indexOfCP":
		return "$indexOfCP"
	case "isArray":
		return "$isArray"
	case "isNumber":
		return "$isNumber"
	case "isoDayOfWeek":
		return "$isoDayOfWeek"
	case "isoWeek":
		return "$isoWeek"
	case "isoWeekYear":
		return "$isoWeekYear"
	// case "last":
	// 	return "$last"  (array)
	// case "last":
	// 	return "$last"  (accumulator)
	case "let":
		return "$let"
	case "literal":
		return "$literal"
	case "ln":
		return "$ln"
	case "log":
		return "$log"
	case "log10":
		return "$log10"
	// case "lt":
	// 	return "$lt"
	// case "lte":
	// 	return "$lte"
	case "ltrim":
		return "$ltrim"
	case "map":
		return "$map"
	case "max":
		return "$max"
	case "mergeObjects":
		return "$mergeObjects"
	case "meta":
		return "$meta"
	case "millisecond":
		return "$millisecond"
	case "min":
		return "$min"
	case "minute":
		return "$minute"
	// case "mod":
	// 	return "$mod"
	case "month":
		return "$month"
	case "multiply":
		return "$multiply"
	// case "ne":
	// 	return "$ne"
	// case "not":
	// 	return "$not"
	case "objectToArray":
		return "$objectToArray"
	// case "or":
	// 	return "$or"
	case "pow":
		return "$pow"
	case "push":
		return "$push"
	case "radiansToDegrees":
		return "$radiansToDegrees"
	case "range":
		return "$range"
	case "reduce":
		return "$reduce"
	case "regexFind":
		return "$regexFind"
	case "regexFindAll":
		return "$regexFindAll"
	case "regexMatch":
		return "$regexMatch"
	case "replaceOne":
		return "$replaceOne"
	case "replaceAll":
		return "$replaceAll"
	case "reverseArray":
		return "$reverseArray"
	case "round":
		return "$round"
	case "rtrim":
		return "$rtrim"
	case "second":
		return "$second"
	case "setDifference":
		return "$setDifference"
	case "setEquals":
		return "$setEquals"
	case "setIntersection":
		return "$setIntersection"
	case "setIsSubset":
		return "$setIsSubset"
	case "setUnion":
		return "$setUnion"
	case "sin":
		return "$sin"
	// case "size":
	// 	return "$size"
	case "slice":
		return "$slice"
	case "split":
		return "$split"
	case "sqrt":
		return "$sqrt"
	case "stdDevPop":
		return "$stdDevPop"
	case "stdDevSamp":
		return "$stdDevSamp"
	case "strLenBytes":
		return "$strLenBytes"
	case "strLenCP":
		return "$strLenCP"
	case "strcasecmp":
		return "$strcasecmp"
	case "substr":
		return "$substr"
	case "substrBytes":
		return "$substrBytes"
	case "substrCP":
		return "$substrCP"
	case "subtract":
		return "$subtract"
	case "sum":
		return "$sum"
	case "switch":
		return "$switch"
	case "tan":
		return "$tan"
	case "toBool":
		return "$toBool"
	case "toDate":
		return "$toDate"
	case "toDecimal":
		return "$toDecimal"
	case "toDouble":
		return "$toDouble"
	case "toInt":
		return "$toInt"
	case "toLong":
		return "$toLong"
	case "toLower":
		return "$toLower"
	case "toObjectId":
		return "$toObjectId"
	case "toString":
		return "$toString"
	case "toUpper":
		return "$toUpper"
	case "trim":
		return "$trim"
	case "trunc":
		return "$trunc"
	// case "type":
	// 	return "$type"
	case "week":
		return "$week"
	case "year":
		return "$year"
	case "zip":
		return "$zip"

	//---MongoDB Update Operators---//
	//Fields
	case "currentDate":
		return "$currentDate"
	case "inc":
		return "$inc"
	// case "min":
	// 	return "$min"
	// case "max":
	// 	return "$max"
	case "mul":
		return "$mul"
	case "rename":
		return "$rename"
	case "set":
		return "$set"
	case "setOnInsert":
		return "$setOnInsert"
	case "unset":
		return "$unset"
	//Array
	// case "addToSet":
	// 	return "$addToSet"
	case "pop":
		return "$pop"
	case "pull":
		return "$pull"
	// case "push":
	// 	return "$push"
	case "pullAll":
		return "$pullAll"
	//Modifiers
	case "each":
		return "$each"
	case "position":
		return "$position"
	// case "slice":
	// 	return "$slice"
	case "sort":
		return "$sort"
	//Bitwise
	case "bit":
		return "$bit"

	}
	return str
}

func CheckJsonData(jsondata map[string]interface{}, mapGenerateID ...[]string) map[string]interface{} {
	for key, result := range jsondata {
		//check jsondata contain utils in array
		if reflect.TypeOf(result).Kind().String() == "slice" {
			for _, r := range jsondata[key].([]interface{}) {
				//----------------------------
				// subData := r.(map[string]interface{})
				// for key, result := range subData {
				// 	if reflect.TypeOf(result).Kind().String() == "slice" {
				// 		for _, r := range subData[key].([]interface{}) {
				// 			if reflect.TypeOf(r).Kind().String() == "map" {
				// 				if _, ok := r.(map[string]interface{})["id"]; !ok {
				// 					// fmt.Println("already have field id----------------------------------",r.(map[string]interface{})["id"])
				// 					r.(map[string]interface{})["id"] = GenerateID("Ar")
				// 				}
				// 			}
				// 		}
				// 	}
				// }
				//----------------------------
				if reflect.TypeOf(r).Kind().String() == "map" {
					subData := r.(map[string]interface{})

					for _, result := range subData {
						if reflect.TypeOf(result).Kind().String() == "slice" {
							if len(mapGenerateID) > 0 {
								CheckJsonData(subData, mapGenerateID[0])
							} else {
								CheckJsonData(subData)
							}
						}
					}

					if key != "image" && ContainInSlice(mapGenerateID[0], key) == true {
						if _, ok := r.(map[string]interface{})["id"]; !ok {
							r.(map[string]interface{})["id"] = GenerateID("Ar")
						}
					}

				}
			}
		}
	}

	return jsondata
}

func GetUrl(ipAddress string, apiGroup string, apiEndpoint string) string {
	var url string
	if ipAddress != "" {
		url = ipAddress
		if apiGroup != "" {
			url = url + apiGroup
		}
		if apiEndpoint != "" {
			url = url + apiEndpoint
		}
	}

	return url
}

func ContainInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func GetMonthNameTH(month int) string {

	switch month {
	//---MongoDB Query Operators---//
	//comparison
	case 1:
		return "มกราคม"
	case 2:
		return "กุมภาพันธ์"
	case 3:
		return "มีนาคม"
	case 4:
		return "เมษายน"
	case 5:
		return "พฤษภาคม"
	case 6:
		return "มิถุนายน"
	case 7:
		return "กรกฎาคม"
	case 8:
		return "สิงหาคม"
	case 9:
		return "กันยายน"
	case 10:
		return "ตุลาคม"
	case 11:
		return "พฤศจิกายน"
	case 12:
		return "ธันวาคม"
	}
	return ""
}

///

// 23 สิงหาคม 2564 18:27:50
func FormatDateTimeTH(raw string) string {
	/// 543

	splitDateTime := strings.Split(raw, " ")
	date := splitDateTime[0]
	time := splitDateTime[1]

	dateSplit := strings.Split(date, "-")
	year := dateSplit[0]
	month := dateSplit[1]
	day := dateSplit[2]

	intYear, _ := strconv.Atoi(year)
	intMonth, _ := strconv.Atoi(month)
	newYear := intYear + 543
	year = strconv.Itoa(newYear)

	// result :=  day + GetMonthNameTH(intMonth) + year + time
	return day + " " + GetMonthNameTH(intMonth) + " " + year + " " + time
}

func ParseFloat64ToInt64(input float64) int64 {
	var a float64 = input
	var b int = int(a)
	c := int64(b)
	return c
}

// input : string month
func MapMonth(monthStr string) string {
	var month string
	switch monthStr {
	case "January":
		month = "1"
	case "February":
		month = "2"

	case "March":
		month = "3"
	case "April":
		month = "4"

	case "May":
		month = "5"
	case "June":
		month = "6"
	case "July":
		month = "7"

	case "August":
		month = "8"
	case "September":
		month = "9"
	case "October":
		month = "10"

	case "November":
		month = "11"
	case "December":
		month = "12"
	}

	return month
}

// 01/07/2022 input format1
// 2022-08-20T08:44:18.127Z input format2
func SpilitDate(input string) (year string, month string, day string) {
	year = ""
	month = ""
	day = ""
	if input != "" {
		if strings.Contains(input, "T") {
			dateSplit := strings.Split(input, "T")
			if len(dateSplit) > 0 {
				dateSplit := strings.Split(dateSplit[0], "-")
				year = dateSplit[0]
				month = dateSplit[1]
				day = dateSplit[2]
			}
		} else if strings.Contains(input, "/") {
			dateSplit := strings.Split(input, "/")
			if len(dateSplit) > 0 {
				year = dateSplit[2]
				month = dateSplit[1]
				day = dateSplit[0]
			}
		}
	}
	return year, month, day
}

// fetch input  time 30m0s , 1h30m0s
func FetchTime(input string) (hr string, min string) {
	if len(input) > 0 {
		if strings.Contains(input, "h") { //case : 1h30m0s
			result := strings.Split(input, "h")
			hr = result[0]
			resultMin := strings.Split(result[1], "m")
			min = resultMin[0]
		} else { //case : 30m0s
			result := strings.Split(input, "m")
			hr = "0"
			min = result[0]
			if strings.Contains(min, "s") {
				min = strings.Trim(min, "s")
			}
		}
	}
	return hr, min
}

func ConvertStringToInt(input string) int {
	logrusField := logrusFieldUtils
	logrusField.Method = "ConvertStringToInt"
	out, err := strconv.Atoi(input)
	if err != nil {
		//fmt.Println(err.Error())
		logging.Logger(cnst.Fatal, err, logrusField)
		return -1
	}
	return out
}

func ToString(input interface{}) string {
	newStr := fmt.Sprint(input)
	if newStr == "0" {
		newStr = "00"
	}
	return newStr
}

// 0 = 0.00
// 15 = 0.25
// 30 = 0.50
// 45 = 0.75
// 60 = 1
func MapMinuteHour(input string) (string, string) {
	hrStr := "0"
	minStr := "0"
	if strings.Contains(input, ".") {
		strSplit := strings.Split(input, ".")
		interger := strSplit[0]
		decimal := strSplit[1]
		if decimal == "00" {
			hrStr = interger
			minStr = "00"
		} else if decimal == "25" {
			hrStr = interger
			minStr = "15"
		} else if decimal == "50" || decimal == "5" {
			hrStr = interger
			minStr = "30"
		} else if decimal == "75" {
			hrStr = interger
			minStr = "45"
		}
	} else { // จน เต็ม
		hrStr = input
		minStr = "00"
	}
	return hrStr, minStr
}

func ReverseFormatDate(input string) string {
	result := ""
	if input != "" {
		dateSplit := strings.Split(input, "/")
		if len(dateSplit) > 0 {
			year := dateSplit[0]
			month := dateSplit[1]
			day := dateSplit[2]
			result = day + "/" + month + "/" + year
		}
	}
	return result
}

func NewBoolean(state bool) *bool {
	if state == true {
		state = true
	} else if state == false {
		state = false
	}
	return &state
}

func FindNextDate(then time.Time, monday bool, tuesday bool, wednesday bool, thursday bool, friday bool, saturday bool, sunday bool) time.Time {
	// fmt.Println("FindNextDate2 :", then)

	for {
		if int(then.Weekday()) == 0 {
			if sunday == true {
				return then
			}

		} else if int(then.Weekday()) == 1 {
			if monday == true {
				return then
			}

		} else if int(then.Weekday()) == 2 {
			if tuesday == true {
				return then
			}

		} else if int(then.Weekday()) == 3 {
			if wednesday == true {
				return then
			}

		} else if int(then.Weekday()) == 4 {
			if thursday == true {
				return then
			}

		} else if int(then.Weekday()) == 5 {
			if friday == true {
				return then
			}

		} else if int(then.Weekday()) == 6 {
			if saturday == true {
				return then
			}

		}
		then = then.AddDate(0, 0, 1)
	}

	//return then
}

func CalculateStartDate(then time.Time, monday bool, tuesday bool, wednesday bool, thursday bool, friday bool, saturday bool, sunday bool) (time.Time, bool) {
	isFixWeekDay := false
	if monday == false && tuesday == false && wednesday == false && thursday == false && friday == false && saturday == false && sunday == false {
		return then, isFixWeekDay
	}
	for {
		if int(then.Weekday()) == 0 {
			if sunday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		} else if int(then.Weekday()) == 1 {
			if monday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		} else if int(then.Weekday()) == 2 {
			if tuesday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		} else if int(then.Weekday()) == 3 {
			if wednesday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		} else if int(then.Weekday()) == 4 {
			if thursday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		} else if int(then.Weekday()) == 5 {
			if friday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		} else if int(then.Weekday()) == 6 {
			if saturday == true {
				isFixWeekDay = true
				return then, isFixWeekDay
			}

		}
		then = then.AddDate(0, 0, 1)
	}

	// if sunday { //0
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 0 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 0 {
	// 				break
	// 			}
	// 		}
	// 	}
	// } else	if monday { //1
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 1 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 1 {
	// 				break
	// 			}
	// 		}
	// 	}
	// } else if tuesday { //2
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 2 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 2 {
	// 				break
	// 			}
	// 		}
	// 	}
	// } else if wednesday { //3
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 3 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 3 {
	// 				break
	// 			}
	// 		}
	// 	}
	// } else if thursday { //4
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 4 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 4 {
	// 				break
	// 			}
	// 		}
	// 	}
	// } else if friday { //5
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 5 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 5 {
	// 				break
	// 			}
	// 		}
	// 	}
	// } else if saturday { //6
	// 	isFixWeekDay = true
	// 	if int(then.Weekday()) != 6 {
	// 		for {
	// 			then = then.AddDate(0, 0, 1)
	// 			if int(then.Weekday()) == 6 {
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	//return then, isFixWeekDay
}

// /input 2022-07-01T08:44:18.127Z
// / 01/07/2022
func FormatDateStartGen(input string) string {
	if input != "" {
		if strings.Contains(input, "T") {
			splitStr := strings.Split(input, "T")
			date := splitStr[0]
			splitStr = strings.Split(date, "-")
			day := splitStr[2]
			month := splitStr[1]
			year := splitStr[0]

			newStr := day + "/" + month + "/" + year
			return newStr
		}

	}
	return ""
}

// /input 2022-07-26 16:39:48.5215856 +0700 +07
// / 01/07/2022
func FormatDateStartNotT(input string) string {
	if input != "" {
		if strings.Contains(input, " ") {
			splitStr := strings.Split(input, " ")
			date := splitStr[0]
			splitStr = strings.Split(date, "-")
			day := splitStr[2]
			month := splitStr[1]
			year := splitStr[0]

			newStr := day + "/" + month + "/" + year
			return newStr
		}

	}
	return ""
}

// input 26/08/2022 13:30
func FormatSortDate(intput string) time.Time {
	logrusField := logrusFieldUtils
	logrusField.Method = "FormatSortDate"

	sp := strings.Split(intput, " ")
	datestr := sp[0]
	timestr := sp[1]
	timestrSplit := strings.Split(timestr, ":")
	hrstr := timestrSplit[0]
	minstr := timestrSplit[1]
	yearstr, monthstr, daystr := SpilitDate(datestr)

	yearInt, err := strconv.Atoi(yearstr) //
	if err != nil {
		//fmt.Println(err.Error())
		logging.Logger(cnst.Fatal, err, logrusField)
		return time.Time{}
	}
	monthInt, err := strconv.Atoi(monthstr) //
	if err != nil {
		//fmt.Println(err.Error())
		logging.Logger(cnst.Fatal, err, logrusField)
		return time.Time{}
	}
	dayInt, err := strconv.Atoi(daystr) //
	if err != nil {
		//fmt.Println(err.Error())
		logging.Logger(cnst.Fatal, err, logrusField)
		return time.Time{}
	}
	hrInt, err := strconv.Atoi(hrstr) //
	if err != nil {
		//fmt.Println(err.Error())
		logging.Logger(cnst.Fatal, err, logrusField)
		return time.Time{}
	}
	minInt, err := strconv.Atoi(minstr) //
	if err != nil {
		//fmt.Println(err.Error())
		logging.Logger(cnst.Fatal, err, logrusField)
		return time.Time{}
	}
	month := time.Month(monthInt)

	then := time.Date(yearInt, month, dayInt, hrInt, minInt, 0, 0, time.Now().Location())

	return then
}

// input formate 2022-03-17 08:00
func GetDateFromString(intput string) (year int, month int, day int, hour int, min int) {
	logrusField := logrusFieldUtils
	logrusField.Method = "GetDateFromString"

	year = 0
	month = 0
	day = 0
	hour = 0
	min = 0
	if intput != "" {
		if strings.Contains(intput, " ") {
			splitDateTime := strings.Split(intput, " ")
			if len(splitDateTime) > 0 {
				datestr := splitDateTime[0]
				timestr := splitDateTime[1]

				yearStr, monthStr, dayStr := SpilitDate(datestr)

				hourstr, minstr := SpilitTime(timestr)

				year, err := strconv.Atoi(yearStr)
				if err != nil {
					//fmt.Println(err.Error())
					logging.Logger(cnst.Fatal, err, logrusField)
					return -1, -1, -1, -1, -1
				}

				month, err := strconv.Atoi(monthStr)
				if err != nil {
					//fmt.Println(err.Error())
					logging.Logger(cnst.Fatal, err, logrusField)
					return -1, -1, -1, -1, -1
				}

				day, err := strconv.Atoi(dayStr)
				if err != nil {
					//fmt.Println(err.Error())
					logging.Logger(cnst.Fatal, err, logrusField)
					return -1, -1, -1, -1, -1
				}

				hour, err := strconv.Atoi(hourstr)
				if err != nil {
					//fmt.Println(err.Error())
					logging.Logger(cnst.Fatal, err, logrusField)
					return -1, -1, -1, -1, -1
				}

				min, err := strconv.Atoi(minstr)
				if err != nil {
					//fmt.Println(err.Error())
					logging.Logger(cnst.Fatal, err, logrusField)
					return -1, -1, -1, -1, -1
				}
				return year, month, day, hour, min
			}
		}

	}

	return year, month, day, hour, min
}

// formate 08:00
func SpilitTime(input string) (hour string, min string) {
	if input != "" {
		new := strings.Split(input, ":")
		hour := new[0]
		min := new[1]
		return hour, min
	}
	return "", ""
}

// 01-02-2006 15:04:05 input
func SpilitOnlyYMD(input string) (day string, month string, year string) {
	year = ""
	month = ""
	day = ""
	if input != "" {
		if strings.Contains(input, " ") {
			dateSplit := strings.Split(input, " ")
			if len(dateSplit) > 0 {
				dateSplit := strings.Split(dateSplit[0], "-")
				month = dateSplit[0]
				//fmt.Println(dateSplit[0])
				day = dateSplit[1]
				//fmt.Println(dateSplit[1])
				year = dateSplit[2]
				//fmt.Println(dateSplit[2])
			}
		}
	}
	return day, month, year
}

// rr-2305-001 input
func SpilitCodeForLastIndex(input string) string {
	var lastIndex string
	if input != "" {
		if len(input) > 0 {
			input := strings.Split(input, "-")
			lastIndex = input[len(input)-1]
		}
	}
	return lastIndex
}
