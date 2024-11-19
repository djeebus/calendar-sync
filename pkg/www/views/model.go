package views

import (
	"calendar-sync/pkg/container"
	"calendar-sync/pkg/tasks/workflows"
)

type Views struct {
	ctr       container.Container
	workflows *workflows.Workflows
}

func New(ctr container.Container, workflows *workflows.Workflows) Views {
	return Views{ctr: ctr, workflows: workflows}
}
