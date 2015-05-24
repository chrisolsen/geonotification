package geonotification

import (
	"log"
	"time"
)

var queue *Queue
var devCoordMap *deviceCoordMap

type Notification struct {
	Id      string
	Message string
	Keys    []string
	Start   time.Time
	End     time.Time
}

func (n *Notification) expired() bool {
	return n.End.Unix() < time.Now().Unix()
}

func New(notifications []Notification, redisHost string, redisPort int64, repeatDelayInSeconds time.Duration) *Queue {
	err := SetupCache(redisHost, redisPort)
	if err != nil {
		log.Fatal("Failed to create cache", err)
		return nil
	}

	// register all the messengers
	sender := &notificationSender{}
	sender.Register(GoogleCloudMessenger{"AIzaSyDpSfh-xmbiqvCa2I_-pkULpLffS4FkkEo"})

	queue = &Queue{
		notifications:     notifications,
		delay:             repeatDelayInSeconds,
		sender:            sender,
		recipientProvider: &recipientProvider{}, // this also should allow for registering providers for each os type
		ErrorChan:         make(chan error),
		RemoveChan:        make(chan Notification),
		SentChan:          make(chan []string),
	}

	return queue
}

func AddNotification(n Notification) error {
	log.Println("Adding notification", n)
	return queue.AddItem(n)
}

func RemoveNotification(n Notification) error {
	return queue.RemoveItem(n)
}

func UpdateNotification(n Notification) error {
	return queue.UpdateItem(n)
}

func SetDeviceLocation(regId string, oldCoords, newCoords string) {
	c := GetCache()
	c.deviceCoords.setGeoKey(oldCoords, newCoords, regId)
}
