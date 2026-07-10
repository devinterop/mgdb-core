package cnst

// ค่าเหล่านี้ต้องตรงกับ key ของ actionMap ใน activityapi's getNameInfo (ไม่ใช่ค่า Name ที่เป็น output)
// "edit"/"update" และ "delete"/"cancel" เป็นคำพ้องที่ map ไปยัง NameInfo เดียวกันฝั่ง activityapi
const (
	Create       = "create"
	List         = "list"
	Read         = "info"
	Edit         = "edit"
	EditUpdate   = "update"
	Delete       = "delete"
	Cancel       = "cancel"
	DeleteCancel = "delete"
)
