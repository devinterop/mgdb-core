package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/devinterop/mgdb-core/packages/logging"
)

// DeleteService is to handle Delete function
type DeleteService struct{}

var logrusFieldMongodbDeleteService = structs.LogrusField{
	Module: "MongodbDeleteService",
}

// FindOneAndDelete is for Delete document
func (deleteservice DeleteService) FindOneAndDelete(filter interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbDeleteService
	logrusField.Method = "FindOneAndDelete"

	//check Collection is exist
	if !checkCollectionExist(coll) {
		return nil, nil, false
	}

	//create the context
	exp := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), exp)
	defer cancel()

	//select the collection
	collection := Database.Collection(coll)

	//log.Printf("filter : %v", filter)
	logging.Logger(cnst.Debug, fmt.Sprint("filter: ", filter), logrusField)

	result := collection.FindOneAndDelete(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err(), true
	}

	id := filter.(map[string]interface{})["id"]

	doc := bson.M{}
	decodeErr := result.Decode(&doc)
	//log.Printf("result : %v", decodeErr)
	logging.Logger(cnst.Debug, fmt.Sprint("result:  ", decodeErr), logrusField)

	return id, decodeErr, true
}

// FindOneAndUpdate is for update document
func (deleteservice DeleteService) FindOneAndUpdate(filter interface{}, arrayFilter []interface{}, update interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbDeleteService
	logrusField.Method = "FindOneAndUpdate"

	//check Collection is exist
	if !checkCollectionExist(coll) {
		return nil, nil, false
	}

	//create the context
	exp := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), exp)
	defer cancel()

	//select the collection
	collection := Database.Collection(coll)

	//create an instance of an options and set the desired options
	upsert := true
	after := options.After
	arrayFilters := options.ArrayFilters{
		Filters: arrayFilter,
	}

	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
		ArrayFilters:   &arrayFilters,
	}

	//log.Printf("arrayFilters : %v", arrayFilter)
	//log.Printf("filter : %v", filter)
	//log.Printf("update : %v", update)
	logging.Logger(cnst.Debug, fmt.Sprint("arrayFilters: ", arrayFilter), logrusField)
	logging.Logger(cnst.Debug, fmt.Sprint("filter: ", filter), logrusField)
	logging.Logger(cnst.Debug, fmt.Sprint("update: ", update), logrusField)

	result := collection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		return nil, result.Err(), true
	}

	id := filter.(map[string]interface{})["id"]

	doc := bson.M{}
	decodeErr := result.Decode(&doc)
	//log.Printf("result : %v", decodeErr)
	logging.Logger(cnst.Debug, fmt.Sprint("result: ", decodeErr), logrusField)

	return id, decodeErr, true
}

// DeleteMany is for Delete document
func (deleteservice DeleteService) DeleteMany(filter interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbDeleteService
	logrusField.Method = "DeleteMany"

	//check Collection is exist
	if !checkCollectionExist(coll) {
		return nil, nil, false
	}

	//create the context
	exp := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), exp)
	defer cancel()

	//select the collection
	collection := Database.Collection(coll)

	//log.Printf("filter : %v", filter)
	logging.Logger(cnst.Debug, fmt.Sprint("filter: ", filter), logrusField)

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err, true
	}

	id := filter.(map[string]interface{})["id"]

	//log.Printf("result : %v", result)
	//log.Printf("err : %v", err)
	logging.Logger(cnst.Debug, fmt.Sprint("result: ", result), logrusField)

	return id, err, true
}

// UpdateDocuments is to insert many document
func (deleteservice DeleteService) DeleteManyWithFilter(filter interface{}, data interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbDeleteService
	logrusField.Method = "DeleteManyWithFilter"

	//check Collection is exist
	if !checkCollectionExist(coll) {
		return nil, nil, false
	}

	//create the context
	exp := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), exp)
	defer cancel()

	//select the collection
	collection := Database.Collection(coll)

	result, err := collection.UpdateMany(ctx, filter, data)
	if err != nil {
		return nil, err, true
	}

	id := filter.(map[string]interface{})["id"]
	//log.Printf("result names: %v", result)
	logging.Logger(cnst.Debug, fmt.Sprint("result: ", result), logrusField)

	return id, err, true
}
