//go:build integration
// +build integration

package subscriber

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/InsideGallery/core/queue/generic/subscriber/interfaces"
	"github.com/InsideGallery/core/queue/generic/subscriber/interfaces/mock"
	"github.com/InsideGallery/core/testutils"

	"go.uber.org/mock/gomock"
)

func TestSubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subject1 := "core_subject_1"
	ctx := context.Background()
	clMock := mock.NewMockClient(ctrl)
	subMock := mock.NewMockSubscription(ctrl)
	msgMock := mock.NewMockMsg(ctrl)
	cfgMock := mock.NewMockConfig(ctrl)

	clMock.EXPECT().Context().Return(ctx)
	clMock.EXPECT().Config().Return(cfgMock)
	cfgMock.EXPECT().GetConcurrentSize().Return(1)
	clMock.EXPECT().Config().Return(cfgMock).AnyTimes()
	cfgMock.EXPECT().GetReadTimeout().Return(time.Second)
	clMock.EXPECT().Context().Return(ctx)
	clMock.EXPECT().Context().Return(ctx)
	clMock.EXPECT().Meter().Return(nil)
	clMock.EXPECT().QueueSubscribeSync(gomock.Any(), gomock.Any()).Return(subMock, nil)
	subMock.EXPECT().NextMsg(gomock.Any()).Return(msgMock, nil).AnyTimes()
	msgMock.EXPECT().GetData().Return([]byte("test string")).AnyTimes()
	subMock.EXPECT().Drain().Return(nil)
	cfgMock.EXPECT().GetMaxConcurrentSize().Return(uint64(1)).AnyTimes()
	clMock.EXPECT().Close().Return(nil).AnyTimes()

	ch := make(chan struct{})
	s := NewSubscriber(clMock)

	s.Subscribe(subject1, "test", func(ctx context.Context, msg interfaces.Msg) error {
		slog.Info("Received message", "data", string(msg.GetData()))
		time.Sleep(time.Millisecond * 2)
		ch <- struct{}{}
		return nil
	})

	<-ch
	err := s.Close()
	testutils.Equal(t, err, nil)

	err = s.Wait()
	testutils.Equal(t, err, nil)

	testutils.Equal(t, len(s.GetSubs().GetMap()), 0)
}
