package publisher

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/queue/nats/subscriber"
)

const DefaultPublishTimeout = 10 * time.Second

type Publisher struct {
	Client
}

func New(client Client) *Publisher {
	return &Publisher{
		Client: client,
	}
}

func (p *Publisher) prepareHeaders(keyValues ...string) (map[string][]string, error) {
	l := len(keyValues)
	if l == 0 {
		return nil, nil
	}

	if l%2 != 0 {
		return nil, ErrWrongCountOfArguments
	}

	attributes := map[string][]string{}

	for i := 0; i < len(keyValues)-1; i += 2 {
		key, value := keyValues[i], keyValues[i+1]
		attributes[key] = append(attributes[key], value)
	}

	return attributes, nil
}

func (p *Publisher) Publish(subj string, msg []byte, headers ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultPublishTimeout)
	defer cancel()

	return p.PublishWithContext(ctx, subj, msg, headers...)
}

func (p *Publisher) Requester(subj string, data []byte, timeout time.Duration, headers ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return p.RequesterWithContext(ctx, subj, data, headers...)
}

func (p *Publisher) PublishWithContext(ctx context.Context, subj string, msg []byte, headers ...string) error {
	rawHeaders, err := p.prepareHeaders(headers...)
	if err != nil {
		return err
	}

	msgEntity := &nats.Msg{
		Subject: subj,
		Data:    msg,
		Header:  rawHeaders,
	}

	conn := p.Conn()

	err = conn.PublishMsg(msgEntity)
	if err != nil {
		return err
	}

	_, ok := ctx.Deadline()
	if ok {
		if err = conn.FlushWithContext(ctx); err != nil {
			return err
		}
	} else {
		if err = conn.FlushTimeout(DefaultPublishTimeout); err != nil {
			return err
		}
	}

	return conn.LastError()
}

func (p *Publisher) RequesterWithContext(
	ctx context.Context,
	subj string,
	msg []byte,
	headers ...string,
) ([]byte, error) {
	rawHeaders, err := p.prepareHeaders(headers...)
	if err != nil {
		return nil, err
	}

	msgEntity := &nats.Msg{
		Subject: subj,
		Data:    msg,
		Header:  rawHeaders,
	}

	resp, err := p.Conn().RequestMsgWithContext(ctx, msgEntity)
	if err != nil {
		return nil, err
	}

	return resp.Data, messageError(resp)
}

func messageError(msg *nats.Msg) error {
	errHeader := msg.Header.Get(subscriber.HeaderConsumerError)
	if len(errHeader) == 0 {
		return nil
	}

	return handlerErr(errHeader)
}
