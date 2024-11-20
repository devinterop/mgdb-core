package activity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/devinterop/mgdb-core/app/structs"
	"github.com/devinterop/mgdb-core/packages/logging"
	"github.com/sirupsen/logrus"
)

type ActivityServiceBackend struct{}

var activitylogconfig structs.ActivityLogConfiguration

func InitActivityLog(activitylogconfig structs.ActivityLogConfiguration) {
	activitylogconfig = activitylogconfig
}

func sendToActivityLog(data ActivityInfo) {
	server := activitylogconfig.Server
	servicePath := activitylogconfig.ServicePath
	url := server + servicePath

	byteData, err := json.Marshal(data)
	if err != nil {
		//logrus.Error("err: ", err)
		logging.LoggerV2(logging.Error, err)
	}
	//var jsonStr = []byte(string(byteData))
	// log.Println("send log json :", string(byteData))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byteData))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", os.Getenv("token"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logging.LoggerV2(logging.Error, err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result structs.JsonLogResponseEror
	json.Unmarshal([]byte(body), &result)
	logrus.WithFields(logrus.Fields{
		"application": activitylogconfig.AppName,
		"module":      "ActivityServiceBackend",
		"method":      "sendToActivityLog",
	}).Debug(fmt.Sprintf("result: %s, %s", resp.Status, body))
}

func ActivityLog(data ActivityLogInfo) {

	var logObj ActivityInfo
	logObj.PersonalId = data.PersonalId
	logObj.ApplicationId = activitylogconfig.AppId
	logObj.ApplicationType = data.ApplicationType
	logObj.Activities = data.Activities
	logObj.SubActivities = data.SubActivities
	logObj.Detail = data.Detail
	logObj.ReferenceId = data.ReferenceId
	logging.LoggerV2(logging.Debug, logObj)
	sendToActivityLog(logObj)

}
