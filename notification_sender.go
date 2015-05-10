package geonotification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type MessagePayload struct {
	RegistrationIds []string     `json:"registration_ids"`
	Data            Notification `json:"data"`
}

type NotificationSender interface {
	Send(Notification, []string) *NotificationSendError
	Register(Messenger)
}

type notificationSender struct {
	messengers []Messenger
}

func (n notificationSender) Register(m Messenger) {
	if n.messengers == nil {
		n.messengers = make([]Messenger, 0)
	}
	n.messengers = append(n.messengers, m)
}

type NotificationSendError struct {
	Errors map[string]error
}

func (n *NotificationSendError) Error() string {
	return fmt.Sprintf("%d errors notification errors have occured", len(n.Errors))
}

func (n *NotificationSendError) Add(notificationType string, err error) {
	n.Errors[notificationType] = err
}

/**
 * Calls on all the registered Messenger structs
 * TODO: change recipIds to map[string]string with the keys indicating
 * the service (iOS, GCM,...)
 */
func (n notificationSender) Send(notification Notification, recipIds []string) *NotificationSendError {

	e := &NotificationSendError{}

	// FIXME: send only to regIds for the messenger
	for _, msgr := range n.messengers {
		err := msgr.Send(notification, recipIds)
		if err != nil {
			// on success regIds are marked as complete, so on failure nothing has to be done
			e.Add(msgr.Name(), err)
		}
	}

	return e
}

type Messenger interface {
	Send(Notification, []string) error
	Name() string
}

type GoogleCloudMessenger struct {
	googlePublicKey string
}

func (g GoogleCloudMessenger) Name() string {
	return "GCM"
}

func (n GoogleCloudMessenger) Send(notification Notification, ids []string) error {
	payload := &MessagePayload{
		RegistrationIds: ids,
		Data:            notification,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	r, err := http.NewRequest("POST", "https://android.googleapis.com/gcm/send", bytes.NewReader(data))
	if err != nil {
		return err
	}

	r.Header.Add("Authorization", "key="+n.googlePublicKey)
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Invalid GCM Request: " + resp.Status)
	}

	return nil
}
