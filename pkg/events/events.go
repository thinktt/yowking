package events

import "strings"

type Publisher struct {
	subSets map[*SubscriptionSet]struct{}
}

func (p *Publisher) AddSubSet(s *SubscriptionSet) {
	p.subSets[s] = struct{}{}
}

func (p *Publisher) RemoveSubSet(s *SubscriptionSet) {
	delete(p.subSets, s)
}

func (p *Publisher) PublishMessage(gameID, msg string) {
	for s := range p.subSets {
		if _, exist := s.gameIDs[gameID]; exist {
			s.Channel <- msg
		}
	}
}

var Pub = &Publisher{
	subSets: make(map[*SubscriptionSet]struct{}),
}

type SubscriptionSet struct {
	gameIDs map[string]struct{}
	Channel chan string
}

// PublishMessage allows you to directly publish messages ot this subscription
func (s *SubscriptionSet) PublishMessage(msg string) {
	s.Channel <- msg
}

func (s *SubscriptionSet) AddGameID(gameID string) {
	s.gameIDs[gameID] = struct{}{}
}

func (s *SubscriptionSet) RemoveGameID(gameID string) {
	delete(s.gameIDs, gameID)
}

func (s *SubscriptionSet) Destroy() {
	close(s.Channel)
	Pub.RemoveSubSet(s)
}

func NewSubscriptionSet(gameIDs string) SubscriptionSet {
	subSet := SubscriptionSet{
		gameIDs: make(map[string]struct{}),
		Channel: make(chan string),
	}

	ids := strings.Split(gameIDs, ",")
	for _, gameID := range ids {
		subSet.gameIDs[gameID] = struct{}{}
	}

	// Add the subscription set to the package global publisher
	Pub.AddSubSet(&subSet)

	return subSet
}
