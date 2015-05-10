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

func New(redisHost string, redisPort int64, repeatDelayInSeconds time.Duration) {
	err := SetupCache(redisHost, redisPort)
	if err != nil {
		log.Fatal("Failed to create cache", err)
		return
	}

	// register all the messengers
	sender := notificationSender{}
	sender.Register(GoogleCloudMessenger{})

	queue = &Queue{
		notifications:     make([]Notification, 0, 100),
		delay:             repeatDelayInSeconds,
		sender:            &sender,
		recipientProvider: &recipientProvider{}, // this also should allow for registering providers for each os type
		ErrorChan:         make(chan error),
		RemoveChan:        make(chan Notification),
	}

	queue.Start()
}

func AddNotification(n Notification) error {
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
