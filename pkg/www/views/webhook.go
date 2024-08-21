package views

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/client"

	"calendar-sync/pkg/temporal/workflows"
)

func (v Views) Webhook(c echo.Context) error {
	req := c.Request()

	reqHeaders := req.Header

	ctx := c.Request().Context()
	opts := client.StartWorkflowOptions{}
	args := workflows.ProcessWebhookEventArgs{
		ChannelID:     reqHeaders.Get("X-Goog-Channel-ID"),
		MessageNumber: reqHeaders.Get("X-Goog-Message-Number"),
		ResourceID:    reqHeaders.Get("X-Goog-Resource-ID"),
		ResourceState: reqHeaders.Get("X-Goog-Resource-State"),
		ResourceUri:   reqHeaders.Get("X-Goog-Resource-URI"),
		ChannelToken:  reqHeaders.Get("X-Goog-Channel-Token"),
	}
	if _, err := v.ctr.TemporalClient.ExecuteWorkflow(ctx, opts, workflows.ProcessWebhookEvent, args); err != nil {
		return errors.Wrap(err, "failed to trigger workflow")
	}

	c.Response().WriteHeader(200)
	return nil
}
