package geonotification

import (
	"testing"
	"time"
)

type mockSender struct{}

func (m mockSender) Register(msg Messenger) {
	// do nothing for the mock
}

func (m mockSender) Send(n Notification, regIds []string) *NotificationSendError {
	return &NotificationSendError{}
}

type mockRecipProvider struct{}

func (m mockRecipProvider) fetch(n *Notification) ([]string, error) {
	return []string{"rId_1", "rId_2"}, nil
}

func (m mockRecipProvider) markComplete(n *Notification, rIds []string) {
	// do nothing in the mock
}

func NewTestQueue(
	sender NotificationSender,
	recipientProvider RecipientProvider,
	delay time.Duration) Queue {

	return Queue{
		notifications:     make([]Notification, 0, 100),
		delay:             delay,
		sender:            sender,
		recipientProvider: recipientProvider,
		ErrorChan:         make(chan error),
		RemoveChan:        make(chan Notification),
		SentChan:          make(chan []string),
	}
}

func TestSending(t *testing.T) {
	// ch := make(chan bool)
	queue := NewTestQueue(mockSender{}, mockRecipProvider{}, time.Second)

	n1 := Notification{
		Id:      "1",
		Message: "the Message",
		Keys:    []string{},
		Start:   time.Now(),
		End:     time.Now().Add(time.Hour),
	}

	queue.AddItem(n1)
	queue.Start()

	for {
		select {
		case <-queue.SentChan:
			return
		case <-time.After(time.Second):
			t.Error("Message not sent")
			return
		}
	}
}

func TestAutoRemovalofItems(t *testing.T) {
	queue := NewTestQueue(mockSender{}, mockRecipProvider{}, 50*time.Millisecond)

	n1 := Notification{
		Id:      "1",
		Message: "the Message",
		Keys:    []string{},
		Start:   time.Now(),
		End:     time.Now().Add(time.Hour),
	}

	n2 := Notification{
		Id:      "2",
		Message: "expires soon",
		Keys:    []string{},
		Start:   time.Now(),
		End:     time.Now().Add(time.Millisecond * 100),
	}

	queue.AddItem(n1)
	queue.AddItem(n2)
	queue.Start()

	for {
		select {
		case n := <-queue.RemoveChan:
			if n.Id != n2.Id {
				t.Error("Fail on removal")
			}
			return
		case <-queue.SentChan:
			// do nothing
		case <-time.After(2 * time.Second):
			t.Error("Test timeout")
			return
		}
	}
}

func TestAddItem(t *testing.T) {

	queue := NewTestQueue(mockSender{}, mockRecipProvider{}, time.Second)

	n1 := Notification{
		Id:      "1",
		Message: "the Message",
		Keys:    []string{},
		Start:   time.Now(),
		End:     time.Now().Add(time.Hour),
	}

	n2 := Notification{
		Id:      "1",
		Message: "check Message",
		Keys:    []string{},
		Start:   time.Now(),
		End:     time.Now().Add(time.Hour),
	}

	queue.AddItem(n1)

	if len(queue.notifications) != 1 {
		t.Error("add: invalid queue count of", len(queue.notifications))
		return
	}

	queue.UpdateItem(n2)
	if nCheck := queue.notifications[0]; nCheck.Message != "check Message" {
		t.Error("update: failed")
		return
	}

	queue.RemoveItem(n1)
	if len(queue.notifications) != 0 {
		t.Error("remove: invalid queue count of", len(queue.notifications))
		return
	}
}
