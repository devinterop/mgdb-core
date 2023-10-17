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

// UpdateService is to handle update function
type UpdateService struct{}

var logrusFieldMongodbUpdateService = structs.LogrusField{
	Module: "MongodbUpdateService",
}

// FindOneAndUpdate is for update document
func (updateservice UpdateService) FindOneAndUpdate(filter interface{}, arrayFilter []interface{}, update interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbUpdateService
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

	//log.Printf("filter : %v", filter)
	//log.Printf("update : %v", update)
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

// FindOneAndReplace is for replace document
func (updateservice UpdateService) FindOneAndReplace(filter interface{}, update interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbUpdateService
	logrusField.Method = "FindOneAndReplace"

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
	opt := options.FindOneAndReplaceOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	//log.Printf("filter : %v", filter)
	//log.Printf("update : %v", update)
	logging.Logger(cnst.Debug, fmt.Sprint("filter: ", filter), logrusField)
	logging.Logger(cnst.Debug, fmt.Sprint("update: ", update), logrusField)

	result := collection.FindOneAndReplace(ctx, filter, update, &opt)
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

// UpdateDocuments is to insert many document
func (updateservice UpdateService) UpdateManyDocuments(condition interface{}, data interface{}, coll string) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbUpdateService
	logrusField.Method = "UpdateManyDocuments"

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

	result, err := collection.UpdateMany(ctx, condition, data)
	if err != nil {
		return nil, err, true
	}

	id := condition.(map[string]interface{})["id"]
	//log.Printf("result names: %v", result)
	logging.Logger(cnst.Debug, fmt.Sprint("result: ", result), logrusField)

	return id, err, true
}
