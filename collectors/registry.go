package collectors

import "log/slog"

var (
	RegistryModules = make(map[string]Module)
)

type Module interface {
	Run(logger *slog.Logger, col *Collector)
	Init(key string)
	GetEnabled() bool
	SetConfig(body []byte)
}

func RegisterModule(name string, module Module) {
	RegistryModules[name] = module
}
