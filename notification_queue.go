package geonotification

import "time"

// Queue holds a list of notifications that need to be sent out
// and every set period calls the `.Send()` method
type Queue struct {
	notifications     []Notification
	delay             time.Duration
	sender            NotificationSender
	recipientProvider RecipientProvider
	ErrorChan         chan error
	RemoveChan        chan Notification
	SentChan          chan []string
}

func (q *Queue) Start() {
	go (func() {
		for {
			for _, n := range q.notifications {
				if n.expired() {
					q.RemoveItem(n)
					q.RemoveChan <- n
					continue
				}
				// TODO: need to group recipients by messenger type
				recipientIds, err := q.recipientProvider.fetch(&n)
				if err != nil {
					q.ErrorChan <- err
					continue
				}

				// FIXME: I don't like how an "error" object is returned that
				// contains more errors.  Should return nil if no errors exist
				ferr := q.sender.Send(n, recipientIds)
				if len(ferr.Errors) > 0 {
					for _, e := range ferr.Errors {
						q.ErrorChan <- e
					}
					continue
				}

				q.SentChan <- recipientIds
				q.recipientProvider.markComplete(&n, recipientIds)
			}
			time.Sleep(q.delay)
		}
	})()
}

func (q *Queue) AddItem(n Notification) error {
	q.notifications = append(q.notifications, n)
	return nil
}

func (q *Queue) RemoveItem(n Notification) error {
	for i, sn := range q.notifications {
		if sn.Id == n.Id {
			q.notifications = append(q.notifications[:i], q.notifications[i+1:]...)
		}
	}
	return nil
}

func (q *Queue) UpdateItem(n Notification) error {
	for i, sn := range q.notifications {
		if sn.Id == n.Id {
			q.notifications[i] = n
		}
	}
	return nil
}
