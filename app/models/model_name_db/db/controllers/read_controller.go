package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"

	"github.com/devinterop/mgdb-core/app/models/model_name_db/db/service"

	//"github.com/devinterop/mgdb-core/app/structs"
	"github.com/devinterop/mgdb-core/app/structs"

	"fmt"

	//"github.com/devinterop/mgdb-core/app/models/model_name_db/structs"
	"github.com/devinterop/mgdb-core/packages/logging"

	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReadController is for insert logic
type ReadController struct{}

var logrusFieldMongodbReadController = structs.LogrusField{
	Module: "MongodbReadController",
}

func (auth *ReadController) FindAggregate(pipeline []primitive.M, collectionName string) (bool, interface{}) {
	logrusField := logrusFieldMongodbReadController
	logrusField.Method = "FindAggregate"
	var resultStatus bool
	var resultData interface{}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{}")))

	userservice := service.ReadService{}
	result, err, collection := userservice.AggregateDocument(pipeline, collectionName)
	if err != nil || !collection {
		if !collection {
			c.JSON(500, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": "Collection not found!"})
		} else {
			logging.Logger(cnst.Error, fmt.Sprint("FindAggregate(): error: ", err.Error()), logrusField)
			c.JSON(500, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": err.Error()})
		}
	} else {
		// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following items have gotten successfully", "results": result})
		resultStatus = true
		resultData = result

	}

	return resultStatus, resultData
}

func (auth *ReadController) FindDocumentObj(jsonPost structs.JsonService, mapCon ...map[string]interface{}) (bool, interface{}) {
	logrusField := logrusFieldMongodbReadController
	logrusField.Method = "FindDocumentObj"

	byteArray, err := json.Marshal(jsonPost)
	if err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	//fmt.Println("byteArray=", string(byteArray))
	logging.Logger(cnst.Debug, fmt.Sprint("jsonPost: ", string(byteArray)), logrusField)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{}")))
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byteArray))
	if len(mapCon) > 0 {
		return auth.FindDocument(c, jsonPost, false, mapCon[0])
	}
	return auth.FindDocument(c, jsonPost, false)
}

// / count = true ; only count document non get data
func (auth *ReadController) FindDocumentObjCount(jsonPost structs.JsonService, count bool) (bool, interface{}) {
	logrusField := logrusFieldMongodbReadController
	logrusField.Method = "FindDocumentObjCount"

	byteArray, err := json.Marshal(jsonPost)
	if err != nil {
		// panic(err)
		logging.Logger(cnst.Fatal, err, logrusField)
	}
	//fmt.Println("byteArray=", string(byteArray))
	logging.Logger(cnst.Debug, fmt.Sprint("jsonPost: ", string(byteArray)), logrusField)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{}")))
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byteArray))
	//os.Setenv("collection", jsonPost.Collection)
	return auth.FindDocument(c, jsonPost, count)
}

// FindDocument is for Document insert
func (auth *ReadController) FindDocument(c *gin.Context, jsonService structs.JsonService, count bool, mapCon ...map[string]interface{}) (bool, interface{}) {
	logrusField := logrusFieldMongodbReadController
	logrusField.Method = "FindDocument"
	logging.Logger(cnst.Debug, fmt.Sprint("filter:FindDocument ", jsonService.Condition), logrusField)

	var resultStatus bool
	var resultData interface{}
	aggregate := false
	// var con interface{}
	// var primitiveType primitive.M
	var jsonbody structs.JsonBody

	bytes, er := json.Marshal(jsonService)
	{
		if er == nil {
			err := json.Unmarshal(bytes, &jsonbody)
			if err != nil {
				logging.Logger(cnst.Fatal, err, logrusField)
			}
		}
	}

	if jsonbody.AggregatePipeline != nil { //  Aggregate case
		aggregate = true
	} else {
		//Check if data is empty
		if jsonbody.Projection == nil { //  finddoc case
			logging.Logger(cnst.Error, gin.H{"error": "'Projection': required field is not set"}, logrusField)
			return resultStatus, resultData
		}
	}

	limit := jsonbody.Limit
	skip := limit * (jsonbody.Offset - 1)

	projection := bson.M{}
	date := bson.M{}
	fmt.Println("jsonbody.Projection = == ", jsonbody.Projection)
	if jsonbody.Projection != nil {
		projection, date, aggregate = projectionSet(jsonbody.Projection, jsonbody.Timezone)
	}

	condition := make(map[string]interface{})
	if jsonbody.Condition != nil {
		condition = jsonbody.Condition.(map[string]interface{})
	}

	if len(mapCon) > 0 {
		v := []map[string]interface{}{
			condition, mapCon[0],
		}
		condition = mergeMaps(v...)
	}
	fmt.Println("aggregate = == ", aggregate)
	//find with aggregate
	if aggregate {
		pipeline := []bson.M{}
		if reflect.TypeOf(jsonbody.AggregatePipeline).Kind() == reflect.Slice {
			pipelinesInterface := jsonbody.AggregatePipeline.([]interface{})
			for _, elem := range pipelinesInterface {
				if doc, ok := elem.(map[string]interface{}); ok {
					pipeline = append(pipeline, primitive.M(doc))
				} else {
					// Handle the case where the element is not a map[string]interface{}
					fmt.Println("Skipping element:", elem)
				}

			}
			// pipeline = pipelinesInterface
		} else {
			condition = bson.M{"$match": condition}
			projection = bson.M{"$project": projection}
			sort := bson.M{"$sort": jsonbody.Sort}
			limits := bson.M{"$limit": limit}
			skips := bson.M{"$skip": skip}
			addFields := bson.M{"$addFields": date}

			if len(date) != 0 && len(condition) != 0 {
				pipeline = []bson.M{condition, projection, addFields, sort, skips, limits}
			} else if len(date) == 0 && len(condition) != 0 {
				pipeline = []bson.M{condition, projection, sort, skips, limits}
			} else if len(date) != 0 && len(condition) == 0 {
				pipeline = []bson.M{projection, addFields, sort, skips, limits}
			} else {
				pipeline = []bson.M{projection, sort, skips, limits}
			}
		}
		//fmt.Println("Aggregate(): ", pipeline)
		logging.Logger(cnst.Debug, fmt.Sprint("Aggregate: ", pipeline), logrusField)

		userservice := service.ReadService{}
		result, err, collection := userservice.AggregateDocument(pipeline, jsonbody.Collection)
		if err != nil || !collection {
			if !collection {
				// 500
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": "Collection not found!"})
			} else {
				// 500
				logging.Logger(cnst.Error, fmt.Sprint("userservice.AggregateDocument(): error: ", err.Error()), logrusField)
				c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": err.Error()})
			}
		} else {
			// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following items have gotten successfully", "results": result})
			resultStatus = true
			resultData = result

		}

		//find document
	} else {
		filter := condition
		if len(date) != 0 {
			projection["last_updated"] = date
		}
		//fmt.Println("Find() filter: ", filter)
		//fmt.Println("Find() projection: ", projection)
		logging.Logger(cnst.Debug, fmt.Sprint("filter: ", filter), logrusField)
		logging.Logger(cnst.Debug, fmt.Sprint("projection: ", projection), logrusField)

		userservice := service.ReadService{}
		if count == true {
			result, _, collection := userservice.FindDocumentCount(filter, projection, jsonbody.Collection, jsonbody.Sort, int64(limit), int64(skip))
			if collection == true {
				return true, result
			} else {
				return false, nil
			}
		}
		if *jsonbody.FindOne {
			result, err, collection := userservice.FindOneDocument(filter, projection, jsonbody.Collection, jsonbody.Sort, int64(skip))
			if err != nil || !collection {
				if !collection {
					// 500
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": "Collection not found!"})
				} else {
					// 500
					logging.Logger(cnst.Error, fmt.Sprint("userservice.FindDocument(): error: ", err.Error()), logrusField)
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": err.Error()})
				}
			} else {
				// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following items have gotten successfully", "results": result})
				resultStatus = true
				resultData = result

			}
		} else {
			result, err, collection := userservice.FindDocument(filter, projection, jsonbody.Collection, jsonbody.Sort, int64(limit), int64(skip))
			if err != nil || !collection {
				if !collection {
					// 500
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": "Collection not found!"})
				} else {
					// 500
					logging.Logger(cnst.Error, fmt.Sprint("userservice.FindDocument(): error: ", err.Error()), logrusField)
					c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": err.Error()})
				}
			} else {
				// c.JSON(200, gin.H{"statusCode": setting.AppSetting.HTTP200, "message": "The following items have gotten successfully", "results": result})
				resultStatus = true
				resultData = result

			}
		}
	}
	return resultStatus, resultData
}

// projectionSet is for setup projection data
func projectionSet(p interface{}, timezone string) (bson.M, bson.M, bool) {
	projection := bson.M{}
	date := bson.M{}
	dateList := bson.M{}
	aggregate := false
	project, err := p.(map[string]interface{})
	if err {
	}
	//set default 0 for document's id (MongoDB)
	if len(project) != 0 {
		projection["_id"] = 0
	}
	//set projection data
	for key, result := range project {
		if result == "date" {
			date = bson.M{
				"$dateToString": bson.M{
					"date":     "$" + key,
					"timezone": timezone,
					"format":   "%Y-%m-%dT%H:%M:%S.%L%z",
				},
			}
			dateList[key] = date
		} else {
			projection[key] = result
		}

		if result == 0.0 {
			if key == "_id" {
				continue
			}
			aggregate = true
		}
	}
	return projection, dateList, aggregate
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
