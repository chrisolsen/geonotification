package geonotification

import (
	"log"

	"menteslibres.net/gosexy/redis"
)

var cacheInstance *cache

type cache struct {
	deviceCoords       *deviceCoordMap
	notifiedRecipients *notifiedRecipients
}

func SetupCache(host string, port int64) error {
	r := redis.New()

	if err := r.Connect(host, uint(port)); err != nil {
		log.Fatal("Unable to start Redis", err)
		return err
	}

	cacheInstance = &cache{
		deviceCoords:       &deviceCoordMap{r},
		notifiedRecipients: &notifiedRecipients{r},
	}

	return nil
}

func GetCache() *cache {

	if cacheInstance == nil {
		log.Fatal("No cache instance exists. Must call Setup() before calling Get")
		return nil
	}

	return cacheInstance
}

// notifiedRecipients holds keeps track of who has already been sent a notification
type notifiedRecipients struct {
	rdis *redis.Client
}

func (nr *notifiedRecipients) key(n *Notification) string {
	return "notification-" + n.Id
}

func (nr *notifiedRecipients) add(n *Notification, ids ...string) {
	// FIXME: this should be able to be done with one LPUSH method call, but
	// for some reason the values are inserted as a single array instead of a list
	for _, id := range ids {
		nr.rdis.LPush(nr.key(n), id)
	}
	// nr.rdis.LPush(nr.key(n), ids...)
}

func (nr *notifiedRecipients) get(n *Notification) ([]string, error) {
	return nr.rdis.LRange(nr.key(n), 0, -1)
}

// deviceCoordMap maps the coord keys to lists of device registrationIds currently
// within the geoBlock
type deviceCoordMap struct {
	r *redis.Client
}

func (g *deviceCoordMap) addToGeoKey(coords string, regId string) {
	g.r.LPush(g.toKey(coords), regId)
}

func (g *deviceCoordMap) removeFromGeoKey(coords string, regId string) {
	g.r.LRem(g.toKey(coords), 1, regId)
}

func (g *deviceCoordMap) setGeoKey(oldCoords, newCoords string, regId string) {
	if len(oldCoords) > 0 {
		g.removeFromGeoKey(oldCoords, regId)
	}
	g.addToGeoKey(newCoords, regId)
}

func (g *deviceCoordMap) getIds(coords string) ([]string, error) {
	return g.r.LRange(g.toKey(coords), 0, -1)
}

func (g *deviceCoordMap) existsAt(coords string, regId string) bool {
	ids, err := g.getIds(coords)
	if err != nil {
		return false
	}
	for _, id := range ids {
		if id == regId {
			return true
		}
	}
	return false
}

func (g *deviceCoordMap) toKey(coords string) string {
	return "location:" + coords
}
