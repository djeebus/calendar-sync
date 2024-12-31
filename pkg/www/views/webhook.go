package views

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/tasks/workflows"
)

func (v Views) Webhook(c echo.Context) error {
	req := c.Request()
	requestLogger := logs.GetLogger(req.Context())
	reqHeaders := req.Header
	logHeaders(requestLogger, reqHeaders)

	args := workflows.ProcessWebhookEventArgs{
		ChannelID:     reqHeaders.Get("X-Goog-Channel-ID"),
		MessageNumber: reqHeaders.Get("X-Goog-Message-Number"),
		ResourceID:    reqHeaders.Get("X-Goog-Resource-ID"),
		ResourceState: reqHeaders.Get("X-Goog-Resource-State"),
		ResourceUri:   reqHeaders.Get("X-Goog-Resource-URI"),
		ChannelToken:  reqHeaders.Get("X-Goog-Channel-Token"),
	}

	go v.background(c, func(ctx context.Context) {
		if err := v.workflows.ProcessWebhookEvent(ctx, args); err != nil {
			logger := logs.GetLogger(ctx)
			logger.Error().Str("channel-id", args.ChannelID).Err(err).Msg("failed to process webhook event")
		}
	})

	c.Response().WriteHeader(200)
	return nil
}

func logHeaders(logger *zerolog.Logger, headers http.Header) {
	w := logger.Info()

	for key, values := range headers {
		w = w.Strs("header="+key, values)
	}

	w.Msg("webhook received")
}
