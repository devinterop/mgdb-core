package cnst

// ค่าเหล่านี้ต้องตรงกับ key ของ activityMap ใน activityapi's getNameInfo (ไม่ใช่ค่า Name ที่เป็น output)
const (
	// Customer ชื่อนี้ถูกใช้แล้วใน type_activity.go (ApplicationType) จึงตั้งชื่อแยกสำหรับหมวด Activities
	CustomerActivity = "customer"
	Medical          = "medical"
	Queue            = "queue"
	AppointmentSetup = "appointmentSetup"
	Healtdata        = "healtdata"
	Coupon           = "coupon"
	Personal         = "personal"
	Score            = "score"
	Product          = "product"
	Boradcast        = "boradcast"
	Article          = "article"
	POS              = "pos"
)
