package proxy

import (
	"context"

	"github.com/mailru/easyjson"
	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/memory/set"
	"github.com/InsideGallery/core/queue/nats/proxy/storage"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

type Server struct {
	balancer *Balancer
	client   NATSPublisher
	subjects set.GenericDataSet[string]
	subject  string
}

func DefaultIDGetter(msg *nats.Msg) (string, error) {
	if msg.Header == nil {
		return "", nil
	}

	return msg.Header.Get(HeaderID), nil
}

func NewServer(client NATSPublisher, subject string, subjects []string) *Server {
	return NewServerWithStorage(storage.NewMemory(), client, subject, subjects)
}

func NewServerWithStorage(storage Storage, client NATSPublisher, subject string, subjects []string) *Server {
	return &Server{
		subject:  subject,
		subjects: set.NewGenericDataSet(subjects...),
		balancer: NewBalancer(storage),
		client:   client,
	}
}

func (s *Server) Subscribe(_ context.Context, msg *nats.Msg) error {
	var sub Subscribe

	err := easyjson.Unmarshal(msg.Data, &sub)
	if err != nil {
		return err
	}

	if !s.subjects.Contains(sub.Subject) {
		return ErrWrongSubject
	}

	err = s.balancer.AddInstance(sub.Subject, sub.InstanceID)
	if err != nil {
		return err
	}

	return msg.Respond([]byte(GetPodSubject(sub.Subject, sub.InstanceID)))
}

func (s *Server) Unsubscribe(_ context.Context, msg *nats.Msg) error {
	var unsub Unsubscribe

	err := easyjson.Unmarshal(msg.Data, &unsub)
	if err != nil {
		return err
	}

	return s.balancer.RemoveInstance(unsub.Subject, unsub.InstanceID)
}

func (s *Server) Proxy(subject string) func(ctx context.Context, msg *nats.Msg) error {
	return s.ProxyWithGetter(subject, DefaultIDGetter)
}

func (s *Server) ProxyWithGetter(
	subject string,
	idGetter func(msg *nats.Msg) (string, error),
) func(ctx context.Context, msg *nats.Msg) error {
	return func(ctx context.Context, msg *nats.Msg) error {
		sid, err := idGetter(msg)
		if err != nil {
			return err
		}

		if sid == "" {
			return nil
		}

		instanceID, err := s.balancer.Execute(subject, sid)
		if err != nil {
			return err
		}

		return s.client.PublishWithContext(ctx, GetPodSubject(subject, instanceID), msg.Data)
	}
}

func (s *Server) GetInstances() []string {
	return s.balancer.GetAllInstances()
}

func (s *Server) listenSubscribes(natsHandler *subscriber.Subscriber) error {
	natsHandler.Subscribe(GetSubscribeSubject(s.subject), proxyQueue, s.Subscribe)
	natsHandler.Subscribe(GetUnsubscribeSubject(s.subject), proxyQueue, s.Unsubscribe)

	return nil
}

func (s *Server) Process(natsHandler *subscriber.Subscriber) error {
	return s.ProcessWithGetter(natsHandler, DefaultIDGetter)
}

func (s *Server) ProcessWithGetter(
	natsHandler *subscriber.Subscriber,
	idGetter func(msg *nats.Msg) (string, error),
) error {
	err := s.listenSubscribes(natsHandler)
	if err != nil {
		return err
	}

	for _, subject := range s.subjects.ToSlice() {
		natsHandler.Subscribe(subject, proxyQueue, s.ProxyWithGetter(subject, idGetter))
	}

	return nil
}
