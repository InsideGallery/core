package proxy

import (
	"context"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/queue/nats/subscriber"
	"github.com/InsideGallery/core/server/instance"
)

type Client struct {
	lastPing       time.Time
	client         NATSPopulator
	instanceGetter func() string

	proxyServiceSubject string
	mu                  sync.Mutex
}

func (c *Client) setLastPing() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastPing = time.Now()
}

func (c *Client) getLastPing() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.lastPing
}

func NewClient(client NATSPopulator, proxyServiceSubject string, instanceGetter func() string) *Client {
	if instanceGetter == nil {
		instanceGetter = instance.GetShortInstanceID
	}

	return &Client{
		client:              client,
		proxyServiceSubject: proxyServiceSubject,
		instanceGetter:      instanceGetter,
		lastPing:            time.Now(),
	}
}

func (c *Client) Subscribe(ctx context.Context, subjectToSubscribe string) (string, error) {
	data, err := easyjson.Marshal(Subscribe{
		InstanceID: c.instanceGetter(),
		Subject:    subjectToSubscribe,
	})
	if err != nil {
		return "", err
	}

	subject, err := c.client.RequesterWithContext(ctx, GetSubscribeSubject(c.proxyServiceSubject), data)

	return string(subject), err
}

func (c *Client) Unsubscribe(ctx context.Context, subjectToUnsubscribe string) error {
	data, err := easyjson.Marshal(Unsubscribe{
		InstanceID: c.instanceGetter(),
		Subject:    subjectToUnsubscribe,
	})
	if err != nil {
		return err
	}

	return c.client.PublishWithContext(ctx, GetUnsubscribeSubject(c.proxyServiceSubject), data)
}

func (c *Client) PongListener(natsHandler *subscriber.Subscriber) error {
	natsHandler.Subscribe(c.instanceGetter(), proxyQueue, func(_ context.Context, msg *nats.Msg) error {
		c.setLastPing()
		return msg.Respond([]byte(PongMsg))
	})

	return nil
}

func (c *Client) ServiceHealthcheck(checkEvery, maxPingDelay time.Duration, stopFunction func()) {
	ticker := time.NewTicker(checkEvery)
	for range ticker.C {
		lp := c.getLastPing()
		if lp.Add(maxPingDelay).Unix() <= time.Now().Unix() {
			stopFunction()
		}
	}
}

func (c *Client) RunAndWait(
	natsHandler *subscriber.Subscriber,
	checkEvery, maxPingDelay time.Duration,
	stopFunction func(),
) error {
	err := c.PongListener(natsHandler)
	if err != nil {
		return err
	}

	c.ServiceHealthcheck(checkEvery, maxPingDelay, stopFunction)

	return nil
}
