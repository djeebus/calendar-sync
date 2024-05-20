package activities

import (
	"calendar-sync/pkg/container"
	"go.temporal.io/sdk/worker"
)

type Activities struct {
	ctr container.Container
}

func Register(w worker.Worker, ctr container.Container) {
	a := Activities{ctr}
	w.RegisterActivity(&a)
}
