package geonotification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type MessagePayloadData struct {
	Message  string `json:"message"`
	ImageUrl string `json:"image_url"`
}

type MessagePayload struct {
	RegistrationIds []string            `json:"registration_ids"`
	Data            *MessagePayloadData `json:"data"`
}

type NotificationSendError struct {
	Errors map[string]error
}

func (n *NotificationSendError) Error() string {
	return fmt.Sprintf("%d errors notification errors have occured", len(n.Errors))
}

func (n *NotificationSendError) Add(notificationType string, err error) {
	if n.Errors == nil {
		n.Errors = make(map[string]error)
	}

	n.Errors[notificationType] = err
}

type NotificationSender interface {
	Send(Notification, []string) *NotificationSendError
	Register(Messenger)
}

type notificationSender struct {
	messengers []Messenger
}

func (n *notificationSender) Register(m Messenger) {
	n.messengers = append(n.messengers, m)
	log.Printf("Adding %v messenger. %v messengers now exist\n", m.Name(), len(n.messengers))
}

/**
 * Calls on all the registered Messenger structs
 * TODO: change recipIds to map[string]string with the keys indicating
 * the service (iOS, GCM,...)
 */
func (n *notificationSender) Send(notification Notification, recipIds []string) *NotificationSendError {

	if len(n.messengers) == 0 {
		log.Fatal("Error: No messengers have been registered")
	}

	e := &NotificationSendError{}

	// FIXME: send only to regIds for the messenger
	for _, msgr := range n.messengers {
		err := msgr.Send(notification, recipIds)
		if err != nil {
			log.Println("Errors sending via " + msgr.Name())
			// on success regIds are marked as complete, so on failure nothing has to be done
			e.Add(msgr.Name(), err)
		}
		log.Printf("sending %v notifications", msgr.Name())
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
	log.Println("GCM Sending...")

	if len(ids) == 0 {
		return errors.New("Exiting GCM Send due to 0 ids")
	}

	payload := &MessagePayload{
		RegistrationIds: ids,
		Data: &MessagePayloadData{
			Message:  notification.Message,
			ImageUrl: "http://goo.gl/iu3w4",
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Println("JSON payload", string(data))

	r, err := http.NewRequest("POST", "https://android.googleapis.com/gcm/send", bytes.NewReader(data))
	if err != nil {
		return err
	}

	fmt.Println("Sending GCM", n.googlePublicKey)

	r.Header.Add("Authorization", "key="+n.googlePublicKey)
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Println("GCM error", string(b))
		return errors.New("Invalid GCM Request: " + resp.Status)
	}

	return nil
}
