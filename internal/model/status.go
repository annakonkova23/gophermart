package model

const (
	StatusNew        = "NEW"
	StatusProcessed  = "PROCESSED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
)

const (
	StatusCalcRegistered = "REGISTERED"
	StatusCalcProcessed  = "PROCESSED"
	StatusCalcInvalid    = "INVALID"
	StatusCalcProcessing = "PROCESSING"
)

var StatusesXmapCalc = map[string]string{
	StatusCalcRegistered: StatusProcessing,
	StatusCalcProcessed:  StatusProcessed,
	StatusCalcInvalid:    StatusInvalid,
	StatusCalcProcessing: StatusProcessing,
}

var StatusesCalcFinish = map[string]bool{
	"REGISTERED": false,
	"PROCESSED":  true,
	"INVALID":    true,
	"PROCESSING": false,
}
