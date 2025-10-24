package enum

type SeverityEnum int64

const (
	SeverityEMERGENCY SeverityEnum = iota
	SeverityALERT
	SeverityCRITICAL
	SeverityERROR
	SeverityWARNING
	SeverityNOTICE
	SeverityINFO
	SeverityDEBUG
	SeverityOK
)

var Severity = map[SeverityEnum]string{
	SeverityEMERGENCY: "EMERGENCY",
	SeverityALERT:     "ALERT",
	SeverityCRITICAL:  "CRITICAL",
	SeverityERROR:     "ERROR",
	SeverityWARNING:   "WARNING",
	SeverityNOTICE:    "NOTICE",
	SeverityINFO:      "INFO",
	SeverityDEBUG:     "DEBUG",
	SeverityOK:        "OK",
}

func (_enum SeverityEnum) String() string {
	return Severity[_enum]
}
