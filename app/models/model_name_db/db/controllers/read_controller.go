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
func (auth *ReadController) FindDocumentObjCount(jsonPost structs.JsonService, count bool, mapCon ...map[string]interface{}) (bool, interface{}) {
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
	if len(mapCon) > 0 {
		return auth.FindDocument(c, jsonPost, count, mapCon[0])
	}
	return auth.FindDocument(c, jsonPost, count)
}

// FindDocument is for Document insert
func (auth *ReadController) FindDocument(c *gin.Context, jsonService structs.JsonService, count bool, mapCon ...map[string]interface{}) (bool, interface{}) {
	logrusField := logrusFieldMongodbReadController
	logrusField.Method = "FindDocument"
	logging.Logger(cnst.Debug, fmt.Sprintf("filter:FindDocument %v", jsonService.Condition), logrusField)

	var resultStatus bool
	var resultData interface{}

	jsonbody, err := parseJsonService(jsonService)
	if err != nil {
		logging.Logger(cnst.Fatal, err, logrusField)
		return resultStatus, resultData
	}

	aggregate := jsonbody.AggregatePipeline != nil
	if !aggregate && jsonbody.Projection == nil {
		logging.Logger(cnst.Error, gin.H{"error": "'Projection': required field is not set"}, logrusField)
		return resultStatus, resultData
	}

	limit, skip := jsonbody.Limit, jsonbody.Limit*(jsonbody.Offset-1)
	projection, date, aggregate := getProjectionAndDate(jsonbody.Projection, jsonbody.Timezone, aggregate)
	condition := mergeConditions(jsonbody.Condition, mapCon...)

	if aggregate {
		return handleAggregate(c, jsonbody, condition, projection, date, limit, skip, logrusField)
	} else {
		return handleFind(c, jsonbody, count, condition, projection, date, limit, skip, logrusField)
	}
}

func projectionSet(p interface{}, timezone string) (bson.M, bson.M, bool) {
	projection := bson.M{}
	dateList := bson.M{}
	aggregate := false

	// Type assertion with proper error handling
	project, _ := p.(map[string]interface{})

	// Set default 0 for document's id (MongoDB)
	if len(project) != 0 {
		projection["_id"] = 0
	}

	// Set projection data
	for key, result := range project {
		if result == "date" {
			date := bson.M{
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

		// Check for aggregation requirement
		if result == 0.0 && key != "_id" {
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

func parseJsonService(jsonService structs.JsonService) (structs.JsonBody, error) {
	var jsonbody structs.JsonBody
	bytes, err := json.Marshal(jsonService)
	if err == nil {
		err = json.Unmarshal(bytes, &jsonbody)
	}
	return jsonbody, err
}

func getProjectionAndDate(projection interface{}, timezone string, aggregate bool) (bson.M, bson.M, bool) {
	proj := bson.M{}
	date := bson.M{}
	if projection != nil {
		proj, date, aggregate = projectionSet(projection, timezone)
	}
	return proj, date, aggregate
}

func mergeConditions(condition interface{}, mapCon ...map[string]interface{}) map[string]interface{} {
	cond := make(map[string]interface{})
	if condition != nil {
		cond = condition.(map[string]interface{})
	}
	if len(mapCon) > 0 {
		cond = mergeMaps(append([]map[string]interface{}{cond}, mapCon...)...)
	}
	return cond
}

func handleAggregate(c *gin.Context, jsonbody structs.JsonBody, condition, projection, date bson.M, limit, skip int, logrusField structs.LogrusField) (bool, interface{}) {
	pipeline := createPipeline(jsonbody, condition, projection, date, limit, skip)
	logging.Logger(cnst.Debug, fmt.Sprintf("Aggregate: %v", pipeline), logrusField)

	userservice := service.ReadService{}
	result, err, collection := userservice.AggregateDocument(pipeline, jsonbody.Collection)
	if err != nil || !collection {
		handleError(c, err, collection, logrusField)
		return false, nil
	}

	return true, result
}

func createPipeline(jsonbody structs.JsonBody, condition, projection, date bson.M, limit, skip int) []bson.M {
	pipeline := []bson.M{}
	if reflect.TypeOf(jsonbody.AggregatePipeline).Kind() == reflect.Slice {
		for _, elem := range jsonbody.AggregatePipeline.([]interface{}) {
			if doc, ok := elem.(map[string]interface{}); ok {
				pipeline = append(pipeline, primitive.M(doc))
			}
		}
	} else {
		addFields := bson.M{"$addFields": date}
		sort := bson.M{"$sort": jsonbody.Sort}
		limits := bson.M{"$limit": limit}
		skips := bson.M{"$skip": skip}
		pipeline = buildPipeline(condition, projection, addFields, sort, skips, limits)
	}
	return pipeline
}

func buildPipeline(condition, projection, addFields, sort, skips, limits bson.M) []bson.M {
	if len(addFields) != 0 && len(condition) != 0 {
		return []bson.M{condition, projection, addFields, sort, skips, limits}
	} else if len(addFields) == 0 && len(condition) != 0 {
		return []bson.M{condition, projection, sort, skips, limits}
	} else if len(addFields) != 0 && len(condition) == 0 {
		return []bson.M{projection, addFields, sort, skips, limits}
	} else {
		return []bson.M{projection, sort, skips, limits}
	}
}

func handleFind(c *gin.Context, jsonbody structs.JsonBody, count bool, condition, projection, date bson.M, limit, skip int, logrusField structs.LogrusField) (bool, interface{}) {
	userservice := service.ReadService{}
	if count {
		return handleCount(c, jsonbody, userservice, condition, projection, limit, skip)
	}
	return handleDocumentFind(c, jsonbody, userservice, condition, projection, limit, skip, logrusField)
}

func handleCount(c *gin.Context, jsonbody structs.JsonBody, userservice service.ReadService, condition, projection bson.M, limit, skip int) (bool, interface{}) {
	result, _, collection := userservice.FindDocumentCount(condition, projection, jsonbody.Collection, jsonbody.Sort, int64(limit), int64(skip))
	if collection {
		return true, result
	}
	return false, nil
}

func handleDocumentFind(c *gin.Context, jsonbody structs.JsonBody, userservice service.ReadService, condition, projection bson.M, limit, skip int, logrusField structs.LogrusField) (bool, interface{}) {
	var result interface{}
	var err error
	var collection bool

	if *jsonbody.FindOne {
		result, err, collection = userservice.FindOneDocument(condition, projection, jsonbody.Collection, jsonbody.Sort, int64(skip))
	} else {
		result, err, collection = userservice.FindDocument(condition, projection, jsonbody.Collection, jsonbody.Sort, int64(limit), int64(skip))
	}

	if err != nil || !collection {
		handleError(c, err, collection, logrusField)
		return false, nil
	}

	return true, result
}

func handleError(c *gin.Context, err error, collection bool, logrusField structs.LogrusField) {
	if !collection {
		c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": "Collection not found!"})
	} else {
		logging.Logger(cnst.Error, fmt.Sprintf("Error: %v", err.Error()), logrusField)
		c.JSON(http.StatusInternalServerError, gin.H{"statusCode": http.StatusInternalServerError, "message": "The following item haven't gotten", "errors": err.Error()})
	}
}
