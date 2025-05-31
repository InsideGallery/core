package driver

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/metric"

	"github.com/InsideGallery/core/queue/generic/subscriber/interfaces"
	"github.com/InsideGallery/core/queue/nats/client"
)

type NATSMsg struct {
	*nats.Msg
}

func (m *NATSMsg) GetSubject() string {
	return m.Subject
}

func (m *NATSMsg) IsReply() bool {
	return m.Reply != ""
}

func (m *NATSMsg) ReplyTo() string {
	return m.Reply
}

func (m *NATSMsg) Copy(subject string) interfaces.Msg {
	return &NATSMsg{
		Msg: &nats.Msg{
			Data:    m.GetData(),
			Header:  m.GetHeader(),
			Reply:   m.ReplyTo(),
			Subject: subject,
			Sub:     m.Sub,
		},
	}
}

func (m *NATSMsg) SetHeader(key, value string) {
	if m.Header == nil {
		m.Header = nats.Header{}
	}

	m.Header.Set(key, value)
}

func (m *NATSMsg) Respond(data []byte) error {
	return m.Msg.Respond(data)
}

func (m *NATSMsg) GetHeader() map[string][]string {
	return m.Msg.Header
}

func (m *NATSMsg) GetData() []byte {
	return m.Msg.Data
}

func (m *NATSMsg) RespondMsg(msg interfaces.Msg) error {
	return m.Msg.RespondMsg(&nats.Msg{
		Data:    msg.GetData(),
		Header:  msg.GetHeader(),
		Reply:   msg.ReplyTo(),
		Subject: msg.GetSubject(),
	})
}

type NATSSubscription struct {
	*nats.Subscription
}

func (s *NATSSubscription) NextMsg(timeout time.Duration) (interfaces.Msg, error) {
	msg, err := s.Subscription.NextMsg(timeout)
	if err != nil {
		return nil, err
	}

	return &NATSMsg{Msg: msg}, nil
}

func (s *NATSSubscription) Drain() error {
	return s.Subscription.Drain()
}

func (s *NATSSubscription) GetSubject() string {
	return s.Subscription.Subject
}

func (s *NATSSubscription) Pending() (int64, int64, error) {
	v1, v2, err := s.Subscription.Pending()
	return int64(v1), int64(v2), err
}

func (s *NATSSubscription) Dropped() (int64, error) {
	v, err := s.Subscription.Dropped()
	return int64(v), err
}

func (s *NATSSubscription) Delivered() (int64, error) {
	return s.Subscription.Delivered()
}

type NATSConfig struct {
	*client.Config
}

func (c *NATSConfig) GetReadTimeout() time.Duration {
	return c.Config.ReadTimeout
}

func (c *NATSConfig) GetMaxConcurrentSize() uint64 {
	return c.Config.MaxConcurrentSize
}

func (c *NATSConfig) GetConcurrentSize() int {
	return c.Config.ConcurrentSize
}

type NATSSubscriber struct {
	Conn *client.Client
}

func NewNATSSubscriber(conn *client.Client) *NATSSubscriber {
	return &NATSSubscriber{Conn: conn}
}

func (n *NATSSubscriber) WithMeter(m metric.Meter) {
	n.Conn.WithMeter(m)
}

func (n *NATSSubscriber) Meter() metric.Meter {
	return n.Conn.Meter()
}

func (n *NATSSubscriber) Context() context.Context {
	return n.Conn.Context()
}

func (n *NATSSubscriber) Logger() interfaces.Logger {
	return n.Conn.Logger()
}

func (n *NATSSubscriber) Config() interfaces.Config {
	return &NATSConfig{
		Config: n.Conn.Config(),
	}
}

func (n *NATSSubscriber) Close() error {
	return n.Conn.Close()
}

func (n *NATSSubscriber) QueueSubscribeSync(subject, queue string) (interfaces.Subscription, error) {
	sub, err := n.Conn.QueueSubscribeSync(subject, queue)
	if err != nil {
		return nil, err
	}

	return &NATSSubscription{Subscription: sub}, nil
}
