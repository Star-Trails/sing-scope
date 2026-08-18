package singboxapi

import (
	"context"
	"io"
	"time"

	"sing-scope/internal/domain"
)

// ConnectionBatchHandler is invoked whenever a batch of connection events is received from sing-box.
type ConnectionBatchHandler func(events []domain.FlowEvent, isReset bool)

// RunConnectionStream opens and maintains the SubscribeConnections stream, delivering batches to handler.
func RunConnectionStream(ctx context.Context, client *Client, interval time.Duration, handler ConnectionBatchHandler) error {
	stream, err := client.SubscribeConnections(ctx, interval)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		batch, err := stream.Recv()
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}

		now := time.Now()
		protoEvents := batch.GetEvents()
		domainEvents := make([]domain.FlowEvent, 0, len(protoEvents))

		for _, pe := range protoEvents {
			de := NormalizeConnectionEvent(pe, now)
			if de != nil {
				domainEvents = append(domainEvents, *de)
			}
		}

		if len(domainEvents) > 0 || batch.GetReset_() {
			handler(domainEvents, batch.GetReset_())
		}
	}
}
