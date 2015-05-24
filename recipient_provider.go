package geonotification

import (
	"fmt"
	"log"
)

// RecipientProvider provides the ability to fetch a list of geoCoord blocks
// based on the notification
//
// The interface is provided to allow for easy test mocking.
type RecipientProvider interface {
	fetch(*Notification) ([]string, error)
	markComplete(*Notification, []string)
}

type regIds []string

func (r regIds) contains(s string) bool {
	for _, str := range []string(r) {
		if str == s {
			return true
		}
	}
	return false
}

type recipientProvider struct{}

// fetch returns a list of recipientId strings
func (r recipientProvider) fetch(n *Notification) ([]string, error) {
	// FIXME: testing would be a lot easier if a cache pointer was passed in and
	// the notifiedRecients was an interface
	c := GetCache()
	ids := make([]string, 0)

	for _, key := range n.Keys {
		rIds, err := c.deviceCoords.getIds(key)
		if len(rIds) > 0 {
			fmt.Println("IDs: ", rIds)
		}
		if err != nil {
			return nil, err
		}
		ids = append(ids, rIds...)
	}

	// ids that have already been sent a notification
	existingIds, err := c.notifiedRecipients.get(n)
	if err != nil {
		return nil, err
	}

	// don't send to devices that have already recieved a notification
	unmatchedIds := make([]string, 0)
	for _, id := range ids {
		if !regIds(existingIds).contains(id) {
			unmatchedIds = append(unmatchedIds, id)
		}
	}

	return unmatchedIds, nil
}

func (r recipientProvider) markComplete(n *Notification, recipientIds []string) {
	c := GetCache()
	c.notifiedRecipients.add(n, recipientIds...)
	log.Printf("Sent to %v recipients\n", len(recipientIds))
}
