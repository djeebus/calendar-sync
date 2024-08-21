package activities

import (
	"go.temporal.io/sdk/worker"

	"calendar-sync/pkg/container"
)

type Activities struct {
	ctr container.Container
}

func Register(w worker.Worker, ctr container.Container) {
	a := Activities{ctr}
	w.RegisterActivity(&a)
}
