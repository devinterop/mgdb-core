package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/devinterop/mgdb-core/app/models/model_name_db/db/service"

	"fmt"
	"reflect"
	"time"

	"github.com/devinterop/mgdb-core/app/structs"
	"github.com/devinterop/mgdb-core/packages/logging"
	"github.com/devinterop/mgdb-core/utils"

	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateController is for insert logic
type CreateController struct{}

var logrusFieldMongodbCreateController = structs.LogrusField{
	Module: "MongodbCreateController",
}

func (create *CreateController) InsertDocumentObj(jsonPost structs.JsonService, mapGenerateID ...[]string) (bool, interface{}) {
	logrusField := logrusFieldMongodbCreateController
	logrusField.Method = "InsertDocumentObj"

	byteArray, err := json.Marshal(jsonPost)
	if err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	logging.Logger(cnst.Debug, fmt.Sprint("jsonPost: ", string(byteArray)), logrusField)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{}")))
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byteArray))
	if len(mapGenerateID) > 0 {
		return create.InsertDocument(c, mapGenerateID[0])
	}
	return create.InsertDocument(c)
}

// InsertDocument is for Document insert
func (create *CreateController) InsertDocument(c *gin.Context, mapGenerateID ...[]string) (bool, interface{}) {
	logrusField := logrusFieldMongodbCreateController
	logrusField.Method = "InsertDocument"

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
	condition, err := jsonbody.Condition.(map[string]interface{})
	if err {
	}

	//Check if Condition is empty
	if len(condition) == 0 {
		if len(mapGenerateID) > 0 {
			resultStatus, resultData = insertNewDocument(jsonbody, c, mapGenerateID[0])
		} else {
			resultStatus, resultData = insertNewDocument(jsonbody, c)
		}
	} else {
		resultStatus, resultData = insertWithCondition(jsonbody, c)
	}

	return resultStatus, resultData
}

// InsertNewDocument is for insert new document
func insertNewDocument(jsonbody structs.JsonBody, c *gin.Context, mapGenerateID ...[]string) (bool, interface{}) {
	logrusField := logrusFieldMongodbCreateController
	logrusField.Method = "insertNewDocument"

	userservice := service.CreateService{}
	var result bool
	//switch-case for check type of jsonbody.Data to separate type of document (InsertOne or InsertMany)
	switch reflect.TypeOf(jsonbody.Data).Kind() {

	//InsertMany
	case reflect.Slice:
		//check if Atomicity feild not setup
		if !jsonbody.Atomicity {
			// 500
			c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": "Atomicity not setup!"})
			break
		}
		var jsondata []interface{}
		list := reflect.ValueOf(jsonbody.Data)
		for i := 0; i < list.Len(); i++ {
			jsondata = append(jsondata, list.Index(i).Interface())
		}

		for _, doc := range jsondata {
			id := utils.GenerateID("Dc")
			doc.(map[string]interface{})["id"] = id
			doc.(map[string]interface{})["last_updated"] = time.Now()
			for _, result := range doc.(map[string]interface{}) {
				// check jsondata contain document in array
				if reflect.TypeOf(result).Kind() == reflect.Slice {
					for _, r := range result.([]interface{}) {
						if reflect.TypeOf(r).Kind() == reflect.Map {
							r.(map[string]interface{})["id"] = utils.GenerateID("Ar")
						}
					}
				}
			}
		}
		id, err, col := userservice.InsertManyDocuments(jsondata, jsonbody.Collection)
		if err != nil || !col {
			if !col {
				// 500
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": "Collection not found!"})
			} else {
				// 500
				logging.Logger(cnst.Error, fmt.Sprint("userservice.InsertManyDocuments(): error: ", err.Error()), logrusField)
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": err.Error()})
			}
		} else {
			result = true
			// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP201, "message": "The data have created successfully", "results": id})
			return result, id
		}
	//InsertOne
	case reflect.Map:
		jsondata := jsonbody.Data.(map[string]interface{})
		//Set document id with prefix
		jsondata["id"] = utils.GenerateID("Dc")
		jsondata["last_updated"] = time.Now()
		if len(mapGenerateID) > 0 { // มีการ ระบุ field ที่ต้องการ gen id  , หากไม่ระบุมา จะ genให้แค่ id ชั้นนอก field เดียว

			jsondata = utils.CheckJsonData(jsondata, mapGenerateID[0])
		}

		id, err, col := userservice.InsertOneDocument(jsondata, jsonbody.Collection)
		if err != nil || !col {
			if !col {
				// 500
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": "Collection not found!"})
			} else {
				// 500
				logging.Logger(cnst.Error, fmt.Sprint("userservice.InsertOneDocument(): error: ", err.Error()), logrusField)
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": err.Error()})
			}
		} else {
			result = true
			// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP201, "message": "The data have created successfully", "results": id})
			return result, id
		}
	}
	return result, nil
}

// InsertWithCondition is for Document insert with condition
func insertWithCondition(jsonbody structs.JsonBody, c *gin.Context) (bool, interface{}) {
	logrusField := logrusFieldMongodbCreateController
	logrusField.Method = "insertWithCondition"

	var result bool
	var newId string
	condition := jsonbody.Condition.(map[string]interface{})
	for _, result := range condition {
		//check jsondata contain document in array
		if reflect.TypeOf(result).Kind() == reflect.Map {
			for k, r := range result.(map[string]interface{}) {
				if k != utils.MapOperators(k) {
					result.(map[string]interface{})[utils.MapOperators(k)] = r
					delete(result.(map[string]interface{}), k)
				}
			}
		}
	}

	jsondata := jsonbody.Data.(map[string]interface{})
	for key, result := range jsondata {
		//check jsondata contain array
		if reflect.TypeOf(result).Kind() == reflect.Slice {
			//check jsondata contain document in array
			for _, r := range jsondata[key].([]interface{}) {
				if reflect.TypeOf(r).Kind() == reflect.Map {
					if _, ok := r.(map[string]interface{})["id"]; !ok {
						newId = utils.GenerateID("Ar")
						r.(map[string]interface{})["id"] = newId
					}
				}
			}
			if jsonbody.Duplicate != nil {
				jsondata[key] = bson.M{
					"$each": result,
				}
			}
		} else if reflect.TypeOf(result).Kind() == reflect.Map {

			result.(map[string]interface{})["id"] = utils.GenerateID("Ar")
		}
	}

	update := bson.M{}

	// check if item can duplicate in array
	if jsonbody.Duplicate == nil {
		// not set
		jsondata["last_updated"] = time.Now()
		update = bson.M{
			"$set": jsondata,
		}
	} else if !(*jsonbody.Duplicate) {
		// set to false
		update = bson.M{
			"$addToSet": jsondata,
			"$set": bson.M{
				"last_updated": time.Now(),
			},
		}
	} else {
		// set to true
		update = bson.M{
			"$push": jsondata,
			"$set": bson.M{
				"last_updated": time.Now(),
			},
		}
	}

	userservice := service.CreateService{}
	id, err, col := userservice.UpdateDocuments(condition, update, jsonbody.Collection)

	if err != nil || !col {
		if !col {
			// 500
			c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": "Collection not found!"})
		} else {
			// 500
			logging.Logger(cnst.Error, fmt.Sprint("userservice.UpdateDocuments(): error: ", err.Error()), logrusField)
			c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following data haven't created", "errors": err.Error()})
		}
	} else {
		result = true
		if newId != "" {
			return result, newId
		}
		// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP201, "message": "The data have created successfully", "results": id})
		return result, id
	}
	return result, nil
}
