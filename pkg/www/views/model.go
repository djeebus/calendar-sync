package views

import (
	"calendar-sync/pkg/container"
)

type Views struct {
	ctr container.Container
}

func New(ctr container.Container) Views {
	return Views{ctr: ctr}
}
