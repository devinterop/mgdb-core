package db

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/devinterop/mgdb-core/app/models/model_name_db/db/service"

	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/devinterop/mgdb-core/app/structs"
	cnst "github.com/devinterop/mgdb-core/cnst"
	"github.com/devinterop/mgdb-core/packages/logging"
)

var logrusFieldMongodb = structs.LogrusField{
	Module: "Mongodb",
}
var Database *mongo.Database

func GetDBConnected() *mongo.Database {
	return Database
}

// Connect is for get mongo driver connection
func Connect(connectionString string, dbName string, userDb string, passDb string) {
	logrusField := logrusFieldMongodb
	logrusField.Method = "Connect"

	// Database Config
	credential := options.Credential{
		Username: userDb,
		Password: passDb,
	}
	// clientOptions := options.Client().ApplyURI(connectionString)
	clientOptions := options.Client().ApplyURI(connectionString).SetAuth(credential)
	client, err := mongo.NewClient(clientOptions)

	//Set up a context required by mongo.Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)

	//Cancel context to avoid memory leak
	defer cancel()

	// Ping our db connection
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		//log.Fatal("Couldn't connect to the database", err)
		logging.Logger(cnst.Fatal, fmt.Sprint("Couldn't connect to the database", err), logrusField)
	} else {
		//log.Println("Connected!")
		logging.Logger(cnst.Info, fmt.Sprint("DB Connected!"), logrusField)
	}

	//Connect to the database
	db := client.Database(dbName)
	Database = db
	service.DBConnection(db)

	return
}

func CollectionList(db *mongo.Database) []string {
	logrusField := logrusFieldMongodb
	logrusField.Method = "CollectionList"

	//Check collection is not empty
	filter := bson.D{{}}
	names, err := db.ListCollectionNames(context.TODO(), filter)
	if err != nil {
		// Handle error
		// Panic(err)
		//log.Printf("Failed to get coll names: %v", err)
		logging.Logger(cnst.Fatal, fmt.Sprint("Failed to get collection names: ", err), logrusField)
		return nil
	}
	return names
}
func CollectionValidate(db *mongo.Database, collections interface{}) {
	var collectionArr []string
	v := reflect.ValueOf(collections)
	for i := 0; i < v.NumField(); i++ {
		collectionArr = append(collectionArr, v.Field(i).String())
	}
	errCollect := validateCollection(db, collectionArr)
	if errCollect != nil {
		panic(errCollect)
	}
	return
}

func validateCollection(db *mongo.Database, listCollection []string) error {
	logrusField := logrusFieldMongodb
	logrusField.Method = "ValidateCollection"
	//Check collection is not empty
	//  checkbool := false
	lenColl := len(listCollection)
	lenCollTemp := 0
	var notMap []string
	var newError error = nil
	filter := bson.D{{}}
	collectionList, err := db.ListCollectionNames(context.TODO(), filter)
	if err != nil {
		// Handle error
		logging.Logger(cnst.Info, fmt.Sprint("Failed to get  ListCollectionNames ", err), logrusField)
		return nil
	}
	elementMap := make(map[string]string)
	for _, s := range collectionList {
		elementMap[s] = s
		// or just keys, without values: elementMap[s] = ""
	}

	for _, name := range listCollection {
		_, found := elementMap[name]
		if found {
			lenCollTemp++
		} else {
			logging.Logger(cnst.Info, fmt.Sprint("Not Found the collection names : %v", name), logrusField)
			notMap = append(notMap, name)
		}
	}
	if lenCollTemp != lenColl {
		newError = errors.New(strings.Join(notMap, " "))
	}
	return newError
}
