package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/devinterop/mgdb-core/app/models/model_name_db/db/service"
	cnst "github.com/devinterop/mgdb-core/cnst"

	//"github.com/devinterop/mgdb-core/app/models/model_name_db/structs"
	"github.com/devinterop/mgdb-core/app/structs"
	"github.com/devinterop/mgdb-core/packages/logging"
	"github.com/devinterop/mgdb-core/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// DeleteController is for insert logic
type DeleteController struct{}

var logrusFieldMongodbDeleteController = structs.LogrusField{
	Module: "MongodbDeleteController",
}

func (u *DeleteController) DeleteDocumentObj(jsonPost structs.JsonService) (bool, interface{}) {
	logrusField := logrusFieldMongodbDeleteController
	logrusField.Method = "DeleteDocumentObj"

	byteArray, err := json.Marshal(jsonPost)
	if err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	logging.Logger(cnst.Debug, fmt.Sprint("jsonPost: ", string(byteArray)), logrusField)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{}")))
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byteArray))

	return u.DeleteDocument(c)
}

// DeleteDocument is for
func (u *DeleteController) DeleteDocument(c *gin.Context) (bool, interface{}) {
	logrusField := logrusFieldMongodbDeleteController
	logrusField.Method = "DeleteDocument"

	var jsonbody structs.JsonBody
	var resultStatus bool
	var resultData interface{}
	//Check if jsonbody is not following struck format
	if err := c.ShouldBindJSON(&jsonbody); err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
		c.JSON(http.StatusBadRequest, err) // 401 -> 400
		//return resultStatus, resultData
	}

	con, e := jsonbody.Condition.(map[string]interface{})
	if e {
	}
	condition := utils.ConvertOperators(con).(map[string]interface{})

	//Check if Multi is empty
	if jsonbody.Multi == nil {
		// not set
		c.JSON(http.StatusBadRequest, gin.H{"error": "'Multi': required field is not set"}) // 401 -> 400
		// return resultStatus, resultData
	} else if !(*jsonbody.Multi) {
		// set to false
		data, e := jsonbody.Data.(map[string]interface{})
		if e {
		}
		update := utils.ConvertOperators(data).(map[string]interface{})
		//Check if Data is empty
		if len(data) == 0 {
			userservice := service.DeleteService{}
			id, err, coll := userservice.FindOneAndDelete(condition, jsonbody.Collection)
			if err != nil || !coll {
				if !coll {
					// 500
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": "Collection not found!"})
				} else {
					// 500
					logging.Logger(cnst.Error, fmt.Sprint("userservice.FindOneAndDelete(): error: ", err.Error()), logrusField)
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": err.Error()})
				}
			} else {
				// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following users have deleted successfully", "results": id})
				resultStatus = true
				resultData = id
				return resultStatus, resultData
			}
		} else {

			//arrayFilters
			arrayFilters := []interface{}{}
			updateFilter, e := jsonbody.UpdateFilter.(map[string]interface{})
			if e {
			}
			for k, v := range updateFilter {
				arrayFilters = append(arrayFilters, bson.M{k: utils.ConvertOperators(v)})
			}
			logging.Logger(cnst.Debug, fmt.Sprint("arrayFilters: ", arrayFilters), logrusField)

			lastupdate := bson.M{
				"last_updated": time.Now(),
			}
			update["$set"] = lastupdate
			userservice := service.DeleteService{}
			id, err, coll := userservice.FindOneAndUpdate(condition, arrayFilters, update, jsonbody.Collection)
			if err != nil || !coll {
				if !coll {
					// 500
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": "Collection not found!"})
				} else {
					// 500
					logging.Logger(cnst.Error, fmt.Sprint("userservice.FindOneAndUpdate(): error: ", err.Error()), logrusField)
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": err.Error()})
				}
			} else {
				// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following users have deleted successfully", "results": id})
				resultStatus = true
				resultData = id
				return resultStatus, resultData
			}
		}

	} else {
		data, e := jsonbody.Data.(map[string]interface{})
		if e {
		}
		update := utils.ConvertOperators(data).(map[string]interface{})
		//Check if Data is empty
		if len(data) == 0 {
			userservice := service.DeleteService{}
			id, err, coll := userservice.DeleteMany(condition, jsonbody.Collection)
			if err != nil || !coll {
				if !coll {
					// 500
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": "Collection not found!"})
				} else {
					// 500
					logging.Logger(cnst.Error, fmt.Sprint("userservice.DeleteMany(): error: ", err.Error()), logrusField)
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": err.Error()})
				}
			} else {
				// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following users have deleted successfully", "results": id})
				resultStatus = true
				resultData = id
				return resultStatus, resultData
			}
		} else {
			lastupdate := bson.M{
				"last_updated": time.Now(),
			}
			update["$set"] = lastupdate
			userservice := service.DeleteService{}
			id, err, coll := userservice.DeleteManyWithFilter(condition, update, jsonbody.Collection)
			if err != nil || !coll {
				if !coll {
					// 500
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": "Collection not found!"})
				} else {
					// 500
					logging.Logger(cnst.Error, fmt.Sprint("userservice.DeleteManyWithFilter(): error: ", err.Error()), logrusField)
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following user haven’t deleted", "errors": err.Error()})
				}
			} else {
				// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following users have deleted successfully", "results": id})
				resultStatus = true
				resultData = id
				return resultStatus, resultData
			}
		}
	}

	return resultStatus, resultData
}
