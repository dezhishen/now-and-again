package scheduler

import (
	"fmt"
	"sort"
)

// Handler defines a schedule type. Implementations map schedule_data
// to gocron job definitions. Register in init().
type Handler interface {
	Code() string
	BuildJob(data map[string]any) *JobDef
	IsOneShot() bool
}

var registry = map[string]Handler{}

func Register(h Handler) {
	if _, ok := registry[h.Code()]; ok {
		panic(fmt.Sprintf("handler %q already registered", h.Code()))
	}
	registry[h.Code()] = h
}

func HandlerByCode(code string) Handler { return registry[code] }

func AllHandlers() []Handler {
	list := make([]Handler, 0, len(registry))
	for _, h := range registry {
		list = append(list, h)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Code() < list[j].Code() })
	return list
}
