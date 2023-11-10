package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/devinterop/mgdb-core/app/models/model_name_db/db/service"
	"github.com/devinterop/mgdb-core/app/structs"

	"fmt"
	"time"

	//"github.com/devinterop/mgdb-core/app/models/model_name_db/structs"
	"github.com/devinterop/mgdb-core/packages/logging"
	"github.com/devinterop/mgdb-core/utils"

	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateController is for insert logic
type UpdateController struct{}

var logrusFieldMongodbUpdateController = structs.LogrusField{
	Module: "MongodbUpdateController",
}

func (u *UpdateController) UpdateDocumentObj(jsonPost structs.JsonService, mapGenerateID ...[]string) (bool, interface{}) {
	logrusField := logrusFieldMongodbUpdateController
	logrusField.Method = "UpdateDocumentObj"

	byteArray, err := json.Marshal(jsonPost)
	if err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	//fmt.Println("update =  = =", string(byteArray))
	logging.Logger(cnst.Debug, fmt.Sprint("jsonPost: ", string(byteArray)), logrusField)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{}")))
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byteArray))
	if len(mapGenerateID) > 0 {
		return u.UpdateDocument(c, mapGenerateID[0])
	}
	return u.UpdateDocument(c)
}

// UpdateDocument is for
func (u *UpdateController) UpdateDocument(c *gin.Context, mapGenerateID ...[]string) (bool, interface{}) {
	logrusField := logrusFieldMongodbUpdateController
	logrusField.Method = "UpdateDocument"

	var resultStatus bool
	var resultData interface{}
	var jsonbody structs.JsonBody
	//Check if jsonbody is not following struck format
	if err := c.ShouldBindJSON(&jsonbody); err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
		c.JSON(http.StatusBadRequest, err) // 401 -> 400
		return resultStatus, resultData
	}
	//Check if data is empty
	if jsonbody.Data == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'Data': required field is not set"}) // 401 -> 400
		return resultStatus, resultData
	}

	condition, e := jsonbody.Condition.(map[string]interface{})
	if e {
	}
	//Check if Condition is empty
	if len(condition) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'Condition': required field is not set"}) // 401 -> 400
		return resultStatus, resultData
	}

	//Check if Multi is empty
	if jsonbody.Multi == nil {
		// not set
		c.JSON(http.StatusBadRequest, gin.H{"error": "'Multi': required field is not set"}) // 401 -> 400
		return resultStatus, resultData
	} else if !(*jsonbody.Multi) {
		// set to false
		if len(mapGenerateID) > 0 {
			resultStatus, resultData = updateOneDocument(jsonbody, c, mapGenerateID[0])
		} else {
			resultStatus, resultData = updateOneDocument(jsonbody, c)
		}
	} else {
		// set to true
		// fmt.Println("updateMultipleDocument = = ")
		resultStatus, resultData = updateMultipleDocument(jsonbody, c)
	}

	return resultStatus, resultData
}

func updateOneDocument(jsonbody structs.JsonBody, c *gin.Context, mapGenerateID ...[]string) (bool, interface{}) {
	logrusField := logrusFieldMongodbUpdateController
	logrusField.Method = "updateOneDocument"

	var resultStatus bool
	var resultData interface{}
	condition, e := jsonbody.Condition.(map[string]interface{})
	if e {
	}

	jsondata := jsonbody.Data.(map[string]interface{})
	if len(mapGenerateID) > 0 { // มีการ ระบุ field ที่ต้องการ gen id  , หากไม่ระบุมา จะ genให้แค่ id ชั้นนอก field เดียว
		jsondata = utils.CheckJsonData(jsondata, mapGenerateID[0])
	}
	// for key, result := range jsondata {    //old
	// 	//check jsondata contain array
	// 	if reflect.TypeOf(result).Kind() == reflect.Slice {
	// 		//check jsondata contain document in array
	// 		for _, r := range jsondata[key].([]interface{}) {
	// 			if reflect.TypeOf(r).Kind() == reflect.Map {
	// 				if _, ok := r.(map[string]interface{})["id"]; !ok {
	// 					newId := utils.GenerateID("Ar")
	// 					r.(map[string]interface{})["id"] = newId
	// 				}
	// 			}
	// 		}
	// 	} else if reflect.TypeOf(result).Kind() == reflect.Map {
	// 		// if _, ok := result.(map[string]interface{})["id"]; !ok {
	// 		// 	result.(map[string]interface{})["id"] = utils.GenerateID("Ar")
	// 		// }
	// 	}
	// }   //old
	update := bson.M{}

	userservice := service.UpdateService{}
	if jsonbody.Replacement == nil {
		// not set
		c.JSON(http.StatusBadRequest, gin.H{"error": "'Replacement': required field is not set"}) // 401 -> 400
		return resultStatus, resultData
	} else if !(*jsonbody.Replacement) {
		// set to false
		jsondata["last_updated"] = time.Now()

		//update
		inc := bson.M{}
		set := bson.M{}
		for k, v := range jsondata {
			if k == "inc" {
				inc[k] = v
			} else {
				set[k] = v
			}
		}
		fmt.Println("inc : ", inc)
		if len(inc) == 0 {
			update = bson.M{
				"$set": set,
			}
		} else {
			update = bson.M{
				"$inc": inc["inc"],
				"$set": set,
			}
		}
		//arrayFilters
		arrayFilters := []interface{}{}
		updateFilter, e := jsonbody.UpdateFilter.(map[string]interface{})
		if e {
		}
		// fmt.Println("updateFilter : ",updateFilter)
		for k, v := range updateFilter {
			arrayFilters = append(arrayFilters, bson.M{k: utils.ConvertOperators(v)})
		}
		//fmt.Println("arrayFilters : ", arrayFilters)
		logging.Logger(cnst.Debug, fmt.Sprint("arrayFilters: ", arrayFilters), logrusField)

		id, err, coll := userservice.FindOneAndUpdate(condition, arrayFilters, update, jsonbody.Collection)
		if err != nil || !coll {
			if !coll {
				// 500
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't updated", "errors": "Collection not found!"})
			} else {
				// 500
				logging.Logger(cnst.Error, fmt.Sprint("userservice.FindOneAndUpdate(): error: ", err.Error()), logrusField)
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't updated", "errors": err.Error()})
			}
		} else {
			// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following data have updated successfully", "results": id})
			resultStatus = true
			resultData = id
		}

	} else {
		// set to true
		jsondata["id"] = condition["id"]
		jsondata["last_updated"] = time.Now()
		//fmt.Println("FindOneAndReplace  ", jsondata)
		logging.Logger(cnst.Debug, fmt.Sprint("jsonbody: ", jsondata), logrusField)
		update = jsondata

		id, err, coll := userservice.FindOneAndReplace(condition, update, jsonbody.Collection)
		if err != nil || !coll {
			if !coll {
				// 500
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't updated", "errors": "Collection not found!"})
			} else {
				// 500
				logging.Logger(cnst.Error, fmt.Sprint("userservice.FindOneAndReplace(): error: ", err.Error()), logrusField)
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't updated", "errors": err.Error()})
			}
		} else {
			// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following data have updated successfully", "results": id})
			resultStatus = true
			resultData = id
		}
	}
	return resultStatus, resultData
}

func updateMultipleDocument(jsonbody structs.JsonBody, c *gin.Context) (bool, interface{}) {
	logrusField := logrusFieldMongodbUpdateController
	logrusField.Method = "updateMultipleDocument"

	var resultStatus bool
	var resultData interface{}
	condition := utils.ConvertOperators(jsonbody.Condition).(map[string]interface{})
	update := utils.ConvertOperators(jsonbody.Data).(map[string]interface{})
	lastupdate := bson.M{
		"last_updated": time.Now(),
	}
	update["$set"] = lastupdate

	userservice := service.UpdateService{}

	id, err, coll := userservice.UpdateManyDocuments(condition, update, jsonbody.Collection)
	if err != nil || !coll {
		if !coll {
			// 500
			c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't updated", "errors": "Collection not found!"})
		} else {
			// 500
			logging.Logger(cnst.Error, fmt.Sprint("userservice.UpdateManyDocuments(): error: ", err.Error()), logrusField)
			c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't updated", "errors": err.Error()})
		}
	} else {
		// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following data have updated successfully", "results": id})
		resultStatus = true
		resultData = id
	}
	return resultStatus, resultData
}
