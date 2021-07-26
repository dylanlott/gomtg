package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	redis "github.com/go-redis/redis/v7"
	"github.com/zeebo/errs"
)

// BoardStateKey formats a board state key for boardstate to user mapping.
func BoardStateKey(gameID, userID string) string {
	return fmt.Sprintf("%s:%s", gameID, userID)
}

// BoardstateUpdated returns a channel that emits all *BoardState events.
// If one did not exist before it was queried, it will create a new one.
// If one does exist, it will return the existing boardstate channel.
func (s *graphQLServer) BoardstateUpdated(ctx context.Context, gameID string, userID string) (<-chan *BoardState, error) {
	v, ok := s.boardChannels[BoardStateKey(gameID, userID)]
	if !ok {
		boardstates := make(chan *BoardState, 1)
		s.mutex.Lock()
		s.boardChannels[userID] = boardstates
		s.mutex.Unlock()
		return boardstates, nil
	}

	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		delete(s.boardChannels, userID)
		s.mutex.Unlock()
	}()

	return v, nil
}

// UpdateBoardState updates a BoardState in Redis and notifies that BoardState into the BoardChannels directory
// This keys off of *input.User.ID so we might want to consider some pointer safety in the future here.
func (s *graphQLServer) UpdateBoardState(ctx context.Context, input InputBoardState) (*BoardState, error) {
	if input.User.ID == nil {
		return nil, errs.New("invalid boardstate")
	}
	// get a formatted boardstate from input
	bs, err := boardStateFromInput(input)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	// check if the game exists - if it does then allow the boardstate for it to be set.
	g := &Game{}
	err = s.Get(GameKey(input.GameID), &g)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	if reflect.DeepEqual(g, &Game{}) {
		return nil, errs.New("failed to marshall game: %+v", g)
	}

	// set boardstate into redis
	if err := s.Set(BoardStateKey(input.GameID, *input.User.ID), bs); err != nil {
		return nil, fmt.Errorf("failed to persist boardstate: %s", err)
	}

	// emit channel over boardstate
	if ch, ok := s.boardChannels[*input.User.ID]; !ok {
		return nil, errs.New("userID not found")
	} else {
		ch <- bs
	}

	return bs, nil
}

// Boardstates queries Redis for different boardstates per player or game
func (s *graphQLServer) Boardstates(ctx context.Context, gameID string, userID *string) ([]*BoardState, error) {
	game := &Game{}
	err := s.Get(GameKey(gameID), &game)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("game %s does not exist", gameID)
		}
		return nil, fmt.Errorf("failed to find game %s to update boardstates: %s", gameID, err)
	}
	// if username is not provided, send all
	if userID == nil {
		boardstates := []*BoardState{}
		for _, p := range game.PlayerIDs {
			board := &BoardState{}
			err := s.Get(BoardStateKey(gameID, p.ID), &board)
			if err != nil {
				// NB: Should we fail gracefully here?
				return nil, errs.Wrap(err)
			}
			boardstates = append(boardstates, board)
		}
		return boardstates, nil
	}

	// username provided, return that boardstate
	bs := &BoardState{}
	err = s.Get(BoardStateKey(gameID, *userID), &bs)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return []*BoardState{bs}, nil
}

func boardStateFromInput(bs InputBoardState) (*BoardState, error) {
	data, err := json.Marshal(bs)
	if err != nil {
		return nil, errs.New("failed to marshal input game: %s", err)
	}
	new := &BoardState{}
	err = json.Unmarshal(data, &new)
	if err != nil {
		return nil, errs.New("failed to unmarshal game: %s", err)
	}

	return new, nil
}
