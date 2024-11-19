package workflows

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"calendar-sync/pkg"
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/tasks/activities"
)

func (w *Workflows) WatchAll(ctx context.Context) error {
	ctx, _ = setupLogger(ctx, "WatchAll")

	watchConfigs, err := w.a.GetAllWatches(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get watches")
	}

	goodWatches, badWatches := splitWatches(watchConfigs.WatchConfigs)

	for _, watch := range badWatches {
		go w.deleteCalendar(ctx, watch.ID)
	}

	watchConfigsByCalendarID := pkg.ToSet(goodWatches, func(i persistence.WatchConfig) string {
		return i.CalendarID
	})

	inviteConfigs, err := w.a.GetAllInvites(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get invites")
	}

	for _, inviteConfig := range inviteConfigs.InviteConfigs {
		go w.watchCalendar(ctx, watchConfigsByCalendarID, inviteConfig.CalendarID)
	}

	copyConfigs, err := w.a.GetAllCopies(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get copies")
	}

	for _, copyConfig := range copyConfigs.CopyConfigs {
		go w.watchCalendar(ctx, watchConfigsByCalendarID, copyConfig.SourceID)
		go w.watchCalendar(ctx, watchConfigsByCalendarID, copyConfig.DestinationID)
	}

	return nil
}

func splitWatches(configs []persistence.WatchConfig) ([]persistence.WatchConfig, []persistence.WatchConfig) {
	var good, bad []persistence.WatchConfig

	for _, config := range configs {
		if !config.Expiration.IsZero() && config.Expiration.After(time.Now()) {
			good = append(good, config)
			continue
		}

		bad = append(bad, config)
	}

	return good, bad
}

func (w *Workflows) watchCalendar(
	ctx context.Context, existingWatches map[string]struct{}, calendarID string,
) {
	_, ok := existingWatches[calendarID]
	if ok {
		return
	}

	args := activities.WatchCalendarArgs{CalendarID: calendarID}
	if _, err := w.a.WatchCalendar(ctx, args); err != nil {
		log := logs.GetLogger(ctx)
		log.Warn().Err(err).
			Str("calendar-id", calendarID).
			Msg("failed to watch calendar")
	}
	existingWatches[calendarID] = struct{}{}
}

func (w *Workflows) deleteCalendar(ctx context.Context, watchID int) {
	args := activities.DeleteWatchConfigArgs{
		WatchID: watchID,
	}
	if _, err := w.a.DeleteWatchConfig(ctx, args); err != nil {
		log := logs.GetLogger(ctx)
		log.Error().Err(err).
			Int("watch-id", watchID).
			Msg("failed to delete watch config")
	}
}
