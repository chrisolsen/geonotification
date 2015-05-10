package geonotification

import (
	"testing"

	"menteslibres.net/gosexy/redis"
)

func TestMarkComplete(t *testing.T) {
	r := redis.New()
	err := r.Connect("192.168.59.103", 6379)
	if err != nil {
		t.Fatal("Failed to start redis", err)
		return
	}
	defer r.Quit()

	tests := []struct {
		notification *Notification
		sendToIds    []string
		existingIds  []string
	}{
		{
			notification: &Notification{Id: "99"},
			sendToIds:    []string{"1", "2", "3"},
			existingIds:  []string{},
		},
		{
			notification: &Notification{Id: "99"},
			sendToIds:    []string{"1", "2", "3"},
			existingIds:  []string{"4", "5"},
		},

		// duplicate ids...just to make sure things don't blow up
		{
			notification: &Notification{Id: "99"},
			sendToIds:    []string{"1", "2", "3"},
			existingIds:  []string{"3", "4", "5"},
		},
	}

	for _, test := range tests {
		r.FlushAll()
		SetupCache("192.168.59.103", 6379)
		cache := GetCache()

		// fill memcache with the already notified ids
		cache.notifiedRecipients.add(test.notification, test.existingIds...)

		// init count
		nrs, err := cache.notifiedRecipients.get(test.notification)
		if err != nil {
			t.Error("Error when getting initial recipient list", err)
			return
		}
		initCount := len(nrs)

		// perform op
		rp := &recipientProvider{}
		rp.markComplete(test.notification, test.sendToIds)

		// final count
		nrs2, err := cache.notifiedRecipients.get(test.notification)
		if err != nil {
			t.Error("Error when getting final recipient list", err)
			return
		}
		finalCount := len(nrs2)

		// tests
		if initCount+len(test.sendToIds) != finalCount {
			t.Error("Invalid number of ids marked as complete")
			return
		}
	}
}

func TestCacheFetch(t *testing.T) {
	r := redis.New()
	err := r.Connect("192.168.59.103", 6379)
	if err != nil {
		t.Fatal("Failed to start redis", err)
		return
	}
	defer r.Quit()

	tests := []struct {
		notification         *Notification
		newRecipients        map[string][]string
		existingRecipientIds []string
		expectedRecipientIds []string
	}{
		// all new recipients
		{
			notification: &Notification{Id: "99", Message: "test", Keys: []string{"key1", "key2"}},
			newRecipients: map[string][]string{
				"key1": []string{"3", "1", "4"},
				"key2": []string{"5", "9"},
			},
			existingRecipientIds: []string{},
			expectedRecipientIds: []string{"1", "3", "4", "5", "9"},
		},

		// overlapping recipients
		{
			notification: &Notification{Id: "99", Message: "test", Keys: []string{"key1", "key2"}},
			newRecipients: map[string][]string{
				"key1": []string{"3", "1", "4"},
				"key2": []string{"5", "9"},
			},
			existingRecipientIds: []string{"3", "1", "5"},
			expectedRecipientIds: []string{"4", "9"},
		},

		// no new recipients
		{
			notification:         &Notification{Id: "99", Message: "test", Keys: []string{"key1", "key2"}},
			newRecipients:        map[string][]string{},
			existingRecipientIds: []string{"3", "1", "5"},
			expectedRecipientIds: []string{},
		},

		// no new recipients 2
		{
			notification: &Notification{Id: "99", Message: "test", Keys: []string{"key1", "key2"}},
			newRecipients: map[string][]string{
				"key1": []string{"3", "1", "4"},
				"key2": []string{"5", "9"},
			},
			existingRecipientIds: []string{"3", "1", "4", "5", "9"},
			expectedRecipientIds: []string{},
		},
	}

	for _, test := range tests {
		r.FlushAll()
		SetupCache("192.168.59.103", 6379)
		cache := GetCache()

		// insert new recipients into the redis cache based on their current
		// geolocation key
		for geoKey, recips := range test.newRecipients {
			for _, recipId := range recips {
				cache.deviceCoords.addToGeoKey(geoKey, recipId)
			}
		}

		// insert existing ids for the notification keys
		cache.notifiedRecipients.add(test.notification, test.existingRecipientIds...)

		rp := &recipientProvider{}
		ids, err := rp.fetch(test.notification)
		if err != nil {
			t.Error("Error when fetching", err)
			return
		}

		if len(test.expectedRecipientIds) != len(ids) {
			t.Error("Not sent to the proper recipient count", len(test.expectedRecipientIds), len(ids))
			return
		}

		// ensure that each of the recipIds is returned from the call
		for _, id := range ids {
			if !regIds(test.expectedRecipientIds).contains(id) {
				t.Error("resulting regIds should contain", id)
				return
			}
		}
	}
}
