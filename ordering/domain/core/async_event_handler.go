package core

// AsyncEventHandler คือ interface สำหรับจัดการกับเหตุการณ์ต่างๆ ในระบบ
type AsyncEventHandler interface {
	HandleEvent(event Event) error
	GetAggregateType() string
	GetSubscriptionName() string
}
