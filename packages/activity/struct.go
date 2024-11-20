package activity

// //////////////////////////////////////////////////////////////////////////////////////////////////
type ActivityInfo struct {
	Id              string `json:"id,omitempty"`
	PersonalId      string `json:"personalId,omitempty"`      // id personal
	DateTime        string `json:"dateTime,omitempty"`        // ช่วงเวลาของกิจกรรม
	ApplicationId   string `json:"applicationId,omitempty"`   // application id
	ApplicationType string `json:"applicationType,omitempty"` // ประเภทของกิจกรรม (web admin,customer,staff,e-commerce,blueposh)
	Activities      string `json:"activities,omitempty"`      // ประเภทเมนูที่เกิดกิจกรรม
	SubActivities   string `json:"subActivities,omitempty"`   // ประเภทการกระทำที่เกิดกิจกรรม
	Detail          string `json:"detail,omitempty"`          // รายละเอียดของกิจกรรม
	ReferenceId     string `json:"referenceId,omitempty"`     // id ของข้อมูลต้นทาง
	Last_updated    string `json:"last_updated,omitempty"`
}

type ActivityLogInfo struct {
	PersonalId      string `json:"personalId,omitempty"`      // id personal
	DateTime        string `json:"dateTime,omitempty"`        // ช่วงเวลาของกิจกรรม
	ApplicationId   string `json:"applicationId,omitempty"`   // application id
	ApplicationType string `json:"applicationType,omitempty"` // ประเภทของกิจกรรม (web admin,customer,staff,e-commerce,blueposh)
	Activities      string `json:"activities,omitempty"`      // ประเภทเมนูที่เกิดกิจกรรม
	SubActivities   string `json:"subActivities,omitempty"`   // ประเภทการกระทำที่เกิดกิจกรรม
	Detail          string `json:"detail,omitempty"`          // รายละเอียดของกิจกรรม
	ReferenceId     string `json:"referenceId,omitempty"`     // id ของข้อมูลต้นทาง
}
