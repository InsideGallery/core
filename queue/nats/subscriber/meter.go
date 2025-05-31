package subscriber

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	SubscriptionsPendingCount = "nats.subscriptions.pending.msgs"
	SubscriptionsPendingBytes = "nats.subscriptions.pending.bytes"
	SubscriptionsDroppedMsgs  = "nats.subscriptions.dropped.count"
	SubscriptionCountMsgs     = "nats.subscriptions.send.count"

	Bytes string = "By"

	Subject = attribute.Key("subject")
)

type SubscriptionDetails struct {
	Pending      int64
	PendingBytes int64
	Dropped      int64
	Delivered    int64
}

func getSubscriptionMetrics(sub *nats.Subscription) (*SubscriptionDetails, error) {
	pMsg, pBytes, err := sub.Pending()
	if err != nil {
		return nil, err
	}

	dropped, err := sub.Dropped()
	if err != nil {
		return nil, err
	}

	count, err := sub.Delivered()
	if err != nil {
		return nil, err
	}

	return &SubscriptionDetails{
		Pending:      int64(pMsg),
		PendingBytes: int64(pBytes),
		Dropped:      int64(dropped),
		Delivered:    count,
	}, nil
}

func (s *Subscription) setupMetrics() error {
	m := s.Meter()
	if m == nil {
		return nil
	}

	_, err := m.Int64ObservableGauge(SubscriptionsPendingCount,
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			m, err := getSubscriptionMetrics(s.Subscription)
			if err != nil {
				return err
			}

			observer.Observe(m.Pending, metric.WithAttributes(Subject.String(s.Subscription.Subject)))

			return nil
		}))
	if err != nil {
		return err
	}

	_, err = m.Int64ObservableGauge(SubscriptionsPendingBytes, metric.WithUnit(Bytes),
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			m, err := getSubscriptionMetrics(s.Subscription)
			if err != nil {
				return err
			}

			observer.Observe(m.PendingBytes, metric.WithAttributes(Subject.String(s.Subscription.Subject)))

			return nil
		}))
	if err != nil {
		return err
	}

	_, err = m.Int64ObservableGauge(SubscriptionsDroppedMsgs,
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			m, err := getSubscriptionMetrics(s.Subscription)
			if err != nil {
				return err
			}

			observer.Observe(m.Dropped, metric.WithAttributes(Subject.String(s.Subscription.Subject)))

			return nil
		}))
	if err != nil {
		return err
	}

	_, err = m.Int64ObservableGauge(SubscriptionCountMsgs,
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			m, err := getSubscriptionMetrics(s.Subscription)
			if err != nil {
				return err
			}

			observer.Observe(m.Delivered, metric.WithAttributes(Subject.String(s.Subscription.Subject)))

			return nil
		}))
	if err != nil {
		return err
	}

	return nil
}
