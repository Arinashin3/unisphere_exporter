package api

import (
	"errors"
	"strings"
)

const (
	UnityAPIPrefix   = "/api"
	UnityTypesPrefix = UnityAPIPrefix + "/types"
	UnityInstances   = "/instances"
)

type UnityAction string

const (
	UnityBasicSystemInfo     UnityAction = "basicSystemInfo"
	UnitySystemCapacity      UnityAction = "systemCapacity"
	UnitySystem              UnityAction = "system"
	UnityLun                 UnityAction = "lun"
	UnityPool                UnityAction = "pool"
	UnityStorageResource     UnityAction = "storageResource"
	UnityEvent               UnityAction = "event"
	UnityAlert               UnityAction = "alert"
	UnityMetric              UnityAction = "metric"
	UnityMetricRealTimeQuery UnityAction = "metricRealTimeQuery"
	UnityMetricQueryResult   UnityAction = "metricQueryResult"
	UnityMetricValue         UnityAction = "metricValue"
	UnityFilesystem          UnityAction = "filesystem"
)

func (_action UnityAction) String() string {
	return string(_action)
}

type UnityActionOptions struct {
	Action  UnityAction
	Fields  []string
	Filters []string
	mode    string
	key     string
	Compact bool
}

func NewUnityActionOptions(action string) *UnityActionOptions {
	return &UnityActionOptions{
		Action:  UnityAction(action),
		mode:    "instances",
		Compact: true,
	}
}
func (_opt *UnityActionOptions) WithId(id string) {
	_opt.mode = "id"
	_opt.key = id
}
func (_opt *UnityActionOptions) WithName(name string) {
	_opt.mode = "name"
	_opt.key = name
}
func (_opt *UnityActionOptions) ParseRaw() (string, error) {
	var raw string
	switch _opt.mode {
	case "instances":
		raw = UnityTypesPrefix + "/" + string(_opt.Action) + UnityInstances
	case "id":
		raw = UnityAPIPrefix + UnityInstances + "/" + string(_opt.Action) + "/" + _opt.key
	case "name":
		raw = UnityAPIPrefix + UnityInstances + "/" + string(_opt.Action) + "/name:" + _opt.key
	default:
		return "", errors.New("unsupported action mode")
	}

	raw += "?"
	if _opt.Compact {
		raw += "compact=true"
	} else {
		raw += "compact=false"
	}
	// No Have Fields & Filters...
	if len(_opt.Fields)+len(_opt.Filters) == 0 {
		return raw, nil
	}
	raw += "&"

	// Add Fields...
	if len(_opt.Fields) > 0 {
		raw += "fields=" + strings.Join(_opt.Fields, ",")
		if len(_opt.Filters) > 0 {
			raw += "&"
		}
	}

	// Add Filters...
	if len(_opt.Filters) > 0 {
		for i, s := range _opt.Filters {
			s = strings.Replace(s, " ", "%20", -1)
			raw += "filter=" + s
			if i < len(_opt.Filters)-1 {
				raw += "&"
			}
		}
	}
	return raw, nil
}
