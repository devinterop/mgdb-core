package service

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/devinterop/mgdb-core/packages/logging"
)

// ReadService is to handle create function relation db query
type ReadService struct{}

var logrusFieldMongodbReadService = structs.LogrusField{
	Module: "MongodbReadService",
}

// AggregateDocument is to read documents
func (readservice ReadService) AggregateDocument(data []bson.M, collection string) (interface{}, error, bool) {
	if !checkCollectionExist(collection) {
		return nil, nil, false
	}
	result := []bson.M{}
	opts := options.Aggregate()
	cur, err := Database.Collection(collection).Aggregate(context.TODO(), data, opts)
	if err = cur.All(context.TODO(), &result); err != nil {
		return nil, err, true
	}
	// if err != nil {
	// 	return nil, err, true
	// }
	//Map result to slice
	// for cur.Next(context.TODO()) {
	// 	t := bson.M{}
	// 	err := cur.Decode(&t)
	// 	if err != nil {
	// 		return nil, err, true
	// 	}
	// 	result = append(result, t)
	// }
	cur.Close(context.TODO())
	if len(result) == 0 {
		return result, mongo.ErrNoDocuments, true
	}

	return result, err, true
}

func (readservice ReadService) FindDocument(filter bson.M, projection bson.M, collection string, sort interface{}, limit int64, skip int64) (interface{}, error, bool) {
	if !checkCollectionExist(collection) {
		return nil, nil, false
	}

	result := []bson.M{}
	opts := options.Find()
	opts.SetProjection(projection)
	opts.SetSort(sort)
	opts.SetSkip(skip)
	opts.SetLimit(limit)
	cur, err := Database.Collection(collection).Find(context.TODO(), filter, opts)
	if err = cur.All(context.TODO(), &result); err != nil {
		return nil, err, true
	}
	// if err != nil {
	// 	return nil, err, true
	// }
	//Map result to slice
	// for cur.Next(context.TODO()) {
	// 	t := bson.M{}
	// 	err := cur.Decode(&t)
	// 	if err != nil {
	// 		return nil, err, true
	// 	}
	// 	result = append(result, t)
	// }
	cur.Close(context.TODO())
	if len(result) == 0 {
		return result, mongo.ErrNoDocuments, true
	}

	return result, err, true
}

func (readservice ReadService) FindDocumentCount(filter bson.M, projection bson.M, collection string, sort interface{}, limit int64, skip int64) (interface{}, error, bool) {
	logrusField := logrusFieldMongodbReadService
	logrusField.Method = "FindDocumentCount"

	if !checkCollectionExist(collection) {
		return nil, nil, false
	}
	// result := []bson.M{}

	cotp := options.Count()
	if limit != 0 {
		cotp.SetLimit(limit)
	}
	cotp.SetSkip(skip)

	no, err := Database.Collection(collection).CountDocuments(context.TODO(), filter, cotp)
	if err != nil {
		logging.Logger(cnst.Error, fmt.Sprint("error: ", err.Error()), logrusField)
		return nil, err, false
	}
	// "read = = == = count = ", no)

	return no, err, true
}

func (readservice ReadService) FindOneDocument(filter bson.M, projection bson.M, collection string, sort interface{}, skip int64) (interface{}, error, bool) {
	opts := options.FindOne()
	opts.SetProjection(projection)
	opts.SetSort(sort)
	opts.SetSkip(skip)
	result := bson.M{}
	err := Database.Collection(collection).FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		return nil, mongo.ErrNoDocuments, false
	}

	return result, err, true
}
