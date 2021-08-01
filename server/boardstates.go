package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	redis "github.com/go-redis/redis/v7"
	"github.com/zeebo/errs"
)

// BoardObserver wraps a UserID to a BoardState channel emiter.
type BoardObserver struct {
	// The UserID of the user subscribing to BoardUpdates
	UserID string
	// the channel that BoardUpdates are passed down
	Channel chan *BoardState
}

// FullBoardstate binds a set of observers to a game ID and user ID
type FullBoardstate struct {
	sync.Mutex
	// Game ID of the Boardstate in play
	GameID string
	// User ID of the Boardstate being observed
	UserID string
	// Observers keeps a map of UserID to BoardObservers
	Observers map[string]*BoardObserver
}

// BoardStateKey formats a board state key for boardstate to user mapping.
func BoardStateKey(gameID, userID string) string {
	return fmt.Sprintf("%s:%s", gameID, userID)
}

// BoardstateUpdated returns a channel that emits all *BoardState events.
// If one did not exist before it was queried, it will create a new one.
// If one does exist, it will return the existing boardstate channel.
func (s *graphQLServer) BoardstateUpdated(ctx context.Context,
	obsID string,
	userID string,
) (<-chan *BoardState, error) {
	ch, err := s.registerObserver(ctx, obsID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to register listener: %w", err)
	}

	return ch, nil
}

// UpdateBoardState updates a BoardState in Redis and notifies that BoardState
// into the BoardChannels directory.
// This keys off of *input.User.ID so we want to consider pointer safety here.
func (s *graphQLServer) UpdateBoardState(
	ctx context.Context,
	input InputBoardState,
) (*BoardState, error) {
	if input.User.ID == nil {
		return nil, errs.New("invalid boardstate")
	}

	// get a formatted boardstate from input
	bs, err := boardStateFromInput(input)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	// TODO: Should we merge the two board states here and do a right to left
	// merge? Not sure if that's behavior we want.

	// check if the game exists - if it does then allow the boardstate for it to be set.
	g := &Game{}
	err = s.Get(GameKey(input.GameID), &g)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	// set boardstate into redis
	if err := s.Set(BoardStateKey(input.GameID, *input.User.ID), bs); err != nil {
		return nil, fmt.Errorf("failed to persist boardstate: %s", err)
	}

	go s.publishBoardstate(bs)

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

// inputFromBoardState will return an InputBoardState from a BoardState or
// an error
func inputFromBoardState(bs BoardState) (InputBoardState, error) {
	data, err := json.Marshal(bs)
	if err != nil {
		return InputBoardState{}, errs.New("failed to marshal input game: %s", err)
	}
	new := InputBoardState{}
	err = json.Unmarshal(data, &new)
	if err != nil {
		return InputBoardState{}, errs.New("failed to unmarshal game: %s", err)
	}
	return new, nil
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

// publishes a boardstate update in a threadsafe function.
// * if the userID is not found, it logs an error and returns immediately.
// * if a boardstate for a given userID doesnt' exist, it logs an error and
// returns.
// * it acquires a lock on the *FullBoardState and this must be respected
// or else we'll run into race conditions as well.
func (s *graphQLServer) publishBoardstate(bs *BoardState) {
	s.mutex.Lock()

	if bs.User.ID == "" {
		log.Printf("publishBoardstate error: userID not found: %+v", bs)
		return
	}
	fbs, ok := s.boards[bs.User.ID]
	if !ok {
		log.Printf("pubishBoardState error: could not find boardstate: %s", bs.User.ID)
		return
	}
	obs := fbs.Observers
	s.mutex.Unlock()

	fbs.Mutex.Lock()
	for _, v := range obs {
		v.Channel <- bs
	}
	fbs.Mutex.Unlock()
}

// registerObserver will add an observer with ID obsID to the map of observers
// for userID's BoardState. It returns a channel of BoardState updates or an
// error. This function acquires a lock on the Server so it's threadsafe.
func (s *graphQLServer) registerObserver(ctx context.Context, obsID string, userID string) (chan *BoardState, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// locate if a boardstate exists for that userID already
	fbs, ok := s.boards[userID]
	if !ok {
		// create a fullboardstate since this userID has no current subscribers
		full := &FullBoardstate{
			UserID:    userID,
			Observers: map[string]*BoardObserver{},
		}
		// create and assign observer with obsID to that BoardStates's observers
		obs := &BoardObserver{
			UserID:  obsID,
			Channel: make(chan *BoardState),
		}

		// map observers by ID to the full board state
		full.Observers[obsID] = obs

		// map the fullboardstate by observed boardstate's userID
		s.boards[userID] = full

		// clean up after ourselves
		go func() {
			<-ctx.Done()
			full.Mutex.Lock()
			log.Printf("cleaning up observer [%s] of boardstate [%s]", obsID, userID)
			delete(full.Observers, obsID)
			full.Mutex.Unlock()
		}()

		log.Printf("registered observer [%s] to user [%s] boardstate", obsID, userID)
		return obs.Channel, nil
	}
	// if the fullboardstate exists, then create a new observer
	// and assign it to the fullboardstate
	obs := &BoardObserver{
		UserID:  obsID,
		Channel: make(chan *BoardState),
	}
	if fbs.Observers == nil {
		log.Printf("observers was empty, making a new boardstate observers map")
		fbs.Observers = make(map[string]*BoardObserver)
	}
	fbs.Observers[obsID] = obs
	log.Printf("registered observer [%s] to user [%s] boardstate", obsID, userID)
	log.Printf("list of observers: %+v", fbs.Observers)
	return obs.Channel, nil
}
