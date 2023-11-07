package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/devinterop/mgdb-core/packages/logging"
)

// InitService is to handle create function relation db query
type InitService struct{}

var Database *mongo.Database

var logrusFieldMongodbInitService = structs.LogrusField{
	Module: "MongodbInitService",
}

// DBConnection ..
func DBConnection(c *mongo.Database) {
	Database = c
}

// checkCollectionExist is to check collection exist or not
func checkCollectionExist(collection string) bool {
	logrusField := logrusFieldMongodbInitService
	logrusField.Method = "checkCollectionExist"

	filter := bson.D{{}}
	collectionList, err := Database.ListCollectionNames(context.TODO(), filter)
	if err != nil {
		// Handle error
		// panic(err)
		//log.Printf("Failed to get coll names: %v", err)
		logging.Logger(cnst.Fatal, fmt.Sprint("Failed to get collection names: ", err), logrusField)
		return false
	}
	for _, name := range collectionList {
		if name == collection {
			return true
		}
	}
	return false
}

// get len Documents
// checkCollectionExist is to check collection exist or not
func CountDocuments(collection string, data primitive.M) (int64, error) {
	if !checkCollectionExist(collection) {
		return 0, nil
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	count, err := Database.Collection(collection).CountDocuments(ctx, data)
	defer cancel()

	return count, err
}
