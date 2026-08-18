package singboxapi

import (
	"context"
	"io"
	"time"

	"sing-scope/internal/domain"
)

// StatusHandler is invoked on each system status update.
type StatusHandler func(status *domain.SystemStatus)

// RunStatusStream subscribes to the status stream and dispatches updates to handler.
func RunStatusStream(ctx context.Context, client *Client, interval time.Duration, handler StatusHandler) error {
	stream, err := client.SubscribeStatus(ctx, interval)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pbStatus, err := stream.Recv()
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}

		now := time.Now()
		domStatus := NormalizeStatus(pbStatus, now)
		if domStatus != nil {
			handler(domStatus)
		}
	}
}

// LogHandler is invoked on each log message batch.
type LogHandler func(logs []domain.LogMessage)

// RunLogStream subscribes to the log stream.
func RunLogStream(ctx context.Context, client *Client, handler LogHandler) error {
	stream, err := client.SubscribeLog(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pbLog, err := stream.Recv()
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}

		now := time.Now()
		logs := NormalizeLogs(pbLog, now)
		if len(logs) > 0 {
			handler(logs)
		}
	}
}

// GroupsHandler is invoked on each outbound group update.
type GroupsHandler func(groups []domain.OutboundGroup)

// RunGroupsStream subscribes to outbound group changes.
func RunGroupsStream(ctx context.Context, client *Client, handler GroupsHandler) error {
	stream, err := client.SubscribeGroups(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pbGroups, err := stream.Recv()
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}

		groups := NormalizeGroups(pbGroups)
		if len(groups) > 0 {
			handler(groups)
		}
	}
}
