package conf

import (
	"log"

	"github.com/devinterop/mgdb-core/app/structs"
	"github.com/devinterop/mgdb-core/utils"
	"github.com/spf13/viper"
)

// var Config *viper.Viper

var logrusFieldConfig = structs.LogrusField{
	Module: "Config",
}

// Init is an exported method that takes the environment starts the viper
// (external lib) and returns the configuration struct.
func InitConfigYaml(env string) {
	// logrusField := logrusFieldConfig
	// logrusField.Method = "InitConfigYaml"

	//var err error
	var viperYaml = viper.New()
	viperYaml.SetConfigName(env) // ชื่อไฟล์ Config

	viperYaml.SetConfigType("yaml")
	viperYaml.AddConfigPath("conf") // path จะให้ค้นหากี่ที่ก็ได้ แค่เรียกคำสั่งนี้ไปอีกก็พอ
	// viper.AddConfigPath("ost.utils/conf/") // path ที่ให้ค้นหาไฟล์ Config
	// การค้นหาตาม Path จะเรียงลำดับตาม Path ที่ถูกเพิ่มก่อน ถ้าเจอก็จะหยุดค้นหาใน Path ต่อไปเลย

	// เริ่มการค้นหาไฟล์ Config และอ่านไฟล์
	err := viperYaml.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing configuration file")
		// logging.Logger(cnst.Fatal, "error on parsing yaml configuration file", logrusField)
	} else {
		utils.SetViperYaml(viperYaml)
	}

}

func InitConfigJson(file string) {
	// logrusField := logrusFieldConfig
	// logrusField.Method = "InitConfigJson"

	//var err error
	var viperJson = viper.New()
	viperJson.SetConfigName(file) // ชื่อไฟล์ Config

	viperJson.SetConfigType("json")
	viperJson.AddConfigPath("conf") // path จะให้ค้นหากี่ที่ก็ได้ แค่เรียกคำสั่งนี้ไปอีกก็พอ
	// viper.AddConfigPath("ost.utils/conf/") // path ที่ให้ค้นหาไฟล์ Config
	// การค้นหาตาม Path จะเรียงลำดับตาม Path ที่ถูกเพิ่มก่อน ถ้าเจอก็จะหยุดค้นหาใน Path ต่อไปเลย

	// เริ่มการค้นหาไฟล์ Config และอ่านไฟล์
	err := viperJson.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing json configuration file")
		// logging.Logger(cnst.Fatal, "error on parsing json configuration file", logrusField)
	} else {
		utils.SetViperJson(viperJson)
		//fmt.Println("InitConfigJson(): config: ", targetStruct)
		//utils.ReadViperJsonToStruct(targetStruct)
	}

}
