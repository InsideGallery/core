package app

import (
	"testing"

	"github.com/FrogoAI/mq-balancer/subscriber"
)

func TestMQBalancerSubscriptionMetricNamesRemainDashboardContract(t *testing.T) {
	want := map[string]string{
		"pending messages": subscriber.SubscriptionsPendingCount,
		"pending bytes":    subscriber.SubscriptionsPendingBytes,
		"dropped messages": subscriber.SubscriptionsDroppedMsgs,
		"delivered count":  subscriber.SubscriptionCountMsgs,
	}

	assertMQMetricName(t, want, "pending messages", "queue.subscriptions.pending.msgs")
	assertMQMetricName(t, want, "pending bytes", "queue.subscriptions.pending.bytes")
	assertMQMetricName(t, want, "dropped messages", "queue.subscriptions.dropped.count")
	assertMQMetricName(t, want, "delivered count", "queue.subscriptions.send.count")
}

func assertMQMetricName(t *testing.T, names map[string]string, key string, want string) {
	t.Helper()

	if got := names[key]; got != want {
		t.Fatalf("%s metric = %q, want %q", key, got, want)
	}
}
