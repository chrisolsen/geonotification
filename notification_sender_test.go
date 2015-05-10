package geonotification

import (
	"testing"
	"time"
)

type mockMessenger struct {
	called bool
}

func (m mockMessenger) Send(n Notification, ids []string) error {
	m.called = true
	return nil
}

func (m mockMessenger) Name() string {
	return "mock"
}

func TestRegisteredMessengerIsCalled(t *testing.T) {
	msgr := mockMessenger{}
	sender := notificationSender{}
	sender.Register(msgr)

	// need to delay the sending by a bit to allow the check loop
	// to start up
	done := false
	go (func() {
		sender.Send(Notification{}, []string{"foo", "bar"})
		done = true
	})()

	for {
		if done && msgr.called == false {
			t.Error("messenger Send method not called")
			return
		}
		time.Sleep(time.Millisecond)
		if done {
			return
		}
	}

}
