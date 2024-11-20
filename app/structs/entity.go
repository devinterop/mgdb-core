package structs

//Jsonbody struct is to handle Json body request
type JsonBody struct {
	Collection        string      `json:"collection"`
	Reference         string      `json:"reference"`
	Condition         interface{} `json:"condition"`
	Data              interface{} `json:"data"`
	Projection        interface{} `json:"projection"`
	ArrayFilter       interface{} `json:"arrayFilter"`
	Sort              interface{} `json:"sort"`
	Limit             int         `json:"limit"`
	Offset            int         `json:"offset"`
	Timezone          string      `json:"timezone"`
	Atomicity         bool        `json:"atomicity"`
	Duplicate         *bool       `json:"duplicate"`
	Multi             *bool       `json:"multi"`
	Replacement       *bool       `json:"replacement"`
	UpdateFilter      interface{} `json:"updateFilter"`
	FindOne           *bool       `json:"findone"`
	IsInsert          *bool       `json:"isInsert"`
	AggregatePipeline interface{} `json:"aggregatePipeline"`
	DateKey           []string    `json:"dateKey"`
}

type JsonResponse struct {
	StatusCode interface{} `json:"statusCode"` // string -> int
	Message    string      `json:"message"`
	Results    interface{} `json:"results"`
	SaveStatus bool        `json:"saveStatus"`
	Errors     []ErrorInfo `json:"errors,omitempty"`
}

// use swag json create body (create)
type JsonCreateBody struct {
	Collection string      `json:"collection" example:"collectionName"`
	Condition  interface{} `json:"condition"`
	Data       interface{} `json:"data"`
	Atomicity  bool        `json:"atomicity"`
	Duplicate  *bool       `json:"duplicate"`
}

type JsonReadbody struct {
	Collection string      `json:"collection" example:"collectionName"`
	Condition  interface{} `json:"condition"`
	Projection interface{} `json:"projection"`
	Sort       interface{} `json:"sort"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	Timezone   string      `json:"timezone" example:"+0700"`
}

//Jsonbody to service
type JsonService struct {
	Collection        string      `json:"collection"`
	Reference         string      `json:"reference"`
	Condition         interface{} `json:"condition"`
	Data              interface{} `json:"data"`
	Projection        interface{} `json:"projection"`
	ArrayFilter       interface{} `json:"arrayFilter"`
	Sort              interface{} `json:"sort"`
	Limit             int         `json:"limit"`
	Offset            int         `json:"offset"`
	Timezone          string      `json:"timezone"`
	Atomicity         bool        `json:"atomicity"`
	Duplicate         bool        `json:"duplicate"`
	Multi             bool        `json:"multi"`
	Replacement       bool        `json:"replacement"`
	UpdateFilter      interface{} `json:"updateFilter"`
	FindOne           bool        `json:"findone"`
	IsInsert          bool        `json:"isInsert"`
	AggregatePipeline interface{} `json:"aggregatePipeline"`
	DateKey           []string    `json:"dateKey"`
}

//Jsonbody struct is to handle Json body request
type JsonLogBody struct {
	Level    string      `json:"level"`
	Message  interface{} `json:"message"`
	App_id   string      `json:"app_id"`
	App_name string      `json:"app_name"`
}

type JsonLogResponseEror struct {
	StatusCode interface{} `json:"statusCode"` // string -> int
	Message    string      `json:"message"`
	Results    interface{} `json:"results"`
}

type ValidateResponse struct {
	UserId string `json:"userId"`
	AppId  string `json:"appId"`
}

type LogConfiguration struct {
	Server      string `json:"server,omitempty"`
	ServicePath string `json:"servicePath,omitempty"`
	AppId       string `json:"appId,omitempty"`
	AppName     string `json:"appName,omitempty"`
	Level       string `json:"logLevel,omitempty"`
	OnServerLog bool   `json:"onServerLog,omitempty"`
}

type ActivityLogConfiguration struct {
	Server      string `json:"server,omitempty"`
	ServicePath string `json:"servicePath,omitempty"`
	AppId       string `json:"appId,omitempty"`
	AppName     string `json:"appName,omitempty"`
}

type LogrusField struct {
	Application string `json:"application,omitempty"`
	Module      string `json:"module,omitempty"`
	Method      string `json:"method,omitempty"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
}
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
