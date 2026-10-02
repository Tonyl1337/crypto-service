package telegram

import "sync"

type userState int

const (
	stateNone userState = iota
	stateWaitingRateCoin
	stateWaitingSubscriptionCoin
	stateWaitingSubscriptionInterval
	stateWaitingSubscriptionDelete
)

type userSession struct {
	State userState

	SubscriptionCoinID     string
	SubscriptionCoinSymbol string
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[int64]userSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions: make(map[int64]userSession),
	}
}

func (s *sessionStore) get(chatID int64) userSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessions[chatID]
}

func (s *sessionStore) set(
	chatID int64,
	session userSession,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[chatID] = session
}

func (s *sessionStore) clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, chatID)
}
