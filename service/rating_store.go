package service

import "sync"

type RatingStore interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type Rating struct {
	Count uint32
	Sum   float64
}

type InMomoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

func NewInMomoryRatingStore() *InMomoryRatingStore {
	return &InMomoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

func (store *InMomoryRatingStore) Add(laptopID string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.rating[laptopID]

	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopID] = rating
	return rating, nil
}
