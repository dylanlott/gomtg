package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

// GameObserver binds a UserID to a Channel.
type GameObserver struct {
	UserID  string
	Channel chan *Game
}

// FullGame wraps a Game with Observers and Players so that we can
// access players boardstates with only a game and a player ID
type FullGame struct {
	sync.Mutex

	GameID    string
	Observers map[string]*GameObserver
}

// Games returns a list of Games.
func (s *graphQLServer) Games(ctx context.Context, gameID *string) ([]*Game, error) {
	if gameID == nil {
		return nil, errors.New("not implemented")
	}
	g := &Game{}
	err := s.Get(GameKey(*gameID), &g)
	if err != nil {
		log.Printf("failed to get game %s: %s", *gameID, err)
		return nil, fmt.Errorf("failed to get game %s: %s", *gameID, err)
	}
	return []*Game{g}, nil
}

// GameUpdated returns a channel for a game or an error.
func (s *graphQLServer) GameUpdated(ctx context.Context, gameID string, userID string) (<-chan *Game, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	g, ok := s.games[gameID]
	if !ok {
		log.Printf("game %s not found - creating a subscription channel for %s", gameID, userID)
		game := &FullGame{
			GameID:    gameID,
			Observers: make(map[string]*GameObserver),
		}

		// add observer to the FullGame
		obs := &GameObserver{
			UserID:  userID,
			Channel: make(chan *Game),
		}

		// clean up the observers channel when we're done with it
		go func() {
			<-ctx.Done()
			game.Mutex.Lock()
			log.Printf("cleaning up observer %s game %s", game.GameID, userID)
			delete(game.Observers, userID)
			game.Mutex.Unlock()
		}()

		// game observers are keyed by userID.
		// only one connection per userID is allowed.
		game.Mutex.Lock()
		game.Observers[userID] = obs
		game.Mutex.Unlock()

		// register the game in the main server directory
		s.games[gameID] = game
		return obs.Channel, nil
	}

	// game exists, so just push user into observers and return their channel
	obs := &GameObserver{
		UserID:  userID,
		Channel: make(chan *Game),
	}
	g.Mutex.Lock()
	g.Observers[userID] = obs
	g.Mutex.Unlock()

	return obs.Channel, nil
}

// UpdateGame is what's used to change the name of the game, format, insert
// or remove players, or change other meta informatin about a game.
func (s *graphQLServer) UpdateGame(ctx context.Context, new InputGame) (*Game, error) {
	game := &Game{}
	b, err := json.Marshal(new)
	if err != nil {
		return nil, errs.New("failed to marshal input game: %s", err)
	}
	err = json.Unmarshal(b, &game)
	if err != nil {
		return nil, errs.New("failed to unmarshal game: %s", err)
	}

	// persist the game into redis
	err = s.Set(GameKey(new.ID), game)
	if err != nil {
		return nil, fmt.Errorf("failed to save updated game state: %s", err)
	}

	go s.publishGame(game.ID, game)

	return game, nil
}

// JoinGame ...
func (s *graphQLServer) JoinGame(ctx context.Context, input *InputJoinGame) (*Game, error) {
	// TODO: Handle rejoins by detecting if that player's user.ID already exists
	// in a given game. If it does, just return that same setup.
	// TODO: Check context for User auth and append user info that way
	// TODO: PUll user boardstate creation out into a function since we do it multiple places
	if input.User.ID == nil {
		return nil, errors.New("must provide a user ID to join a game")
	}
	if input.User.Username == "" {
		return nil, errors.New("must provide a username to join a game")
	}

	game := &Game{}
	err := s.Get(GameKey(input.ID), &game)
	if err != nil {
		return nil, fmt.Errorf("failed to get game to join: %s", err)
	}

	user := &User{
		Username: input.User.Username,
		ID:       *input.User.ID,
	}

	// Init default boardstate minus library and commander
	bs := &BoardState{
		User:       user,
		Life:       input.BoardState.Life,
		GameID:     game.ID,
		Hand:       getCards(input.BoardState.Hand),
		Exiled:     getCards(input.BoardState.Exiled),
		Revealed:   getCards(input.BoardState.Revealed),
		Field:      getCards(input.BoardState.Field),
		Controlled: getCards(input.BoardState.Controlled),
	}

	library, err := s.createLibraryFromDecklist(ctx, *input.Decklist)
	if err != nil {
		// Fail gracefully and still populate basic cards
		log.Printf("error creating library from decklist: %+v", err)
		bs.Library = getCards(input.BoardState.Library)
	} else {
		// Happy path
		bs.Library = library
	}

	// TODO: This will eventually have to check the rules of the game to see if it's a
	// Commander game, but for now this works for EDH MVP.
	if len(input.BoardState.Commander) == 0 {
		return nil, errs.New("must supply a Commander for your deck.")
	}

	// TODO: Make this handle multiple commanders?
	commander, err := s.Card(ctx, input.BoardState.Commander[0].Name, nil)
	if err != nil {
		log.Printf("error getting commander for deck: %+v", err)
		// fail gracefully and use their card name so they can still play a game
		cmdr := getCards(input.BoardState.Commander)
		bs.Commander = []*Card{cmdr[0]}
	} else {
		bs.Commander = []*Card{commander[0]}
	}

	// shuffle their library for the start of the game
	shuff, err := Shuffle(bs.Library)
	if err != nil {
		log.Printf("error shuffling library: %s", err)
		return nil, err
	}
	bs.Library = shuff

	// add them to the game's list of players
	game.PlayerIDs = append(game.PlayerIDs, user)

	// set board state in redis keyed by game.ID and user.ID
	err = s.Set(BoardStateKey(game.ID, user.ID), bs)
	if err != nil {
		log.Printf("error persisting boardstate into redis: %s", err)
		return nil, err
	}

	// persist updated game in Redis
	err = s.Set(GameKey(game.ID), game)
	if err != nil {
		return nil, fmt.Errorf("failed to persist game after join: %w", err)
	}

	go s.publishGame(game.ID, game)

	return game, nil
}

// createGame is untested currently
func (s *graphQLServer) CreateGame(ctx context.Context, inputGame InputCreateGame) (*Game, error) {
	// accept a game ID but create one if it isn't assigned
	if inputGame.ID == "" {
		inputGame.ID = uuid.New().String()
	}

	g := &Game{
		ID:        inputGame.ID,
		CreatedAt: time.Now(),
		PlayerIDs: []*User{},
		Turn: &Turn{
			Player: inputGame.Turn.Player,
			Phase:  inputGame.Turn.Phase,
			Number: inputGame.Turn.Number,
		},
		// NB: We're only supporting EDH at this time. We will add more flexible validation later.
		Rules: []*Rule{
			{
				Name:  "format",
				Value: "EDH",
			},
			{
				Name:  "deck_size",
				Value: "99",
			},
		},
	}

	for _, player := range inputGame.Players {
		// TODO: Deck validation should happen here.
		user := &User{
			ID:       *player.User.ID,
			Username: player.User.Username,
		}
		g.PlayerIDs = append(g.PlayerIDs, user)

		// Set default boardstate, handle library and commander specifically
		bs := &BoardState{
			User:       user,
			Life:       player.Life,
			GameID:     g.ID,
			Hand:       getCards(player.Hand),
			Exiled:     getCards(player.Exiled),
			Revealed:   getCards(player.Revealed),
			Field:      getCards(player.Field),
			Controlled: getCards(player.Controlled),
		}

		var decklist string
		if inputGame.Players[0].Decklist != nil {
			decklist = string(*inputGame.Players[0].Decklist)
		}
		library, err := s.createLibraryFromDecklist(ctx, decklist)
		if err != nil {
			// Fail gracefully and still populate basic cards
			log.Printf("error creating library from decklist: %+v", err)
			bs.Library = getCards(player.Library)
		} else {
			// Happy path
			bs.Library = library
		}

		commander, err := s.Card(ctx, player.Commander[0].Name, nil)
		if err != nil {
			log.Printf("error getting commander for deck: %+v", err)
			// fail gracefully and use their card name so they can still play a game
			inputCard := getCards(player.Commander)
			bs.Commander = []*Card{inputCard[0]}
		} else {
			bs.Commander = []*Card{commander[0]}
		}

		shuff, err := Shuffle(bs.Library)
		if err != nil {
			log.Printf("error shuffling library: %s", err)
			return nil, err
		}
		bs.Library = shuff

		// TODO: Use UpdateBoardState instead and use InputBoardState types
		// so that we can set arbitrary board states at create game time.

		// remember that BoardStates are keyed by User.ID not Username anymore
		err = s.Set(BoardStateKey(g.ID, bs.User.ID), bs)
		if err != nil {
			log.Printf("error persisting boardstate into redis: %s", err)
			return nil, err
		}
	}

	// persist the game in Redis
	err := s.Set(GameKey(g.ID), g)
	if err != nil {
		return nil, fmt.Errorf("failed to save created game to redis: %s", err)
	}

	return g, nil
}

func getCards(inputCards []*InputCard) []*Card {
	cardList := []*Card{}

	for _, card := range inputCards {
		c := &Card{
			Name: card.Name,
		}

		if card.ID != nil {
			c.ID = *card.ID
		}

		cardList = append(cardList, c)
	}

	return cardList
}

func (s *graphQLServer) createLibraryFromDecklist(ctx context.Context, decklist string) ([]*Card, error) {
	if decklist == "" {
		return []*Card{}, errs.New("must provide cards in decklist to create a library")
	}
	trimmed := strings.TrimSpace(decklist)
	r := csv.NewReader(strings.NewReader(trimmed))
	// set lazy quotes for using double quotes in csv files
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	cards := []*Card{}

	for {
		// TODO: Use r.ReadAll() to get the whole decklist and do only one
		// DB lookup for all of the cards instead of one by one.
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// handle error path
			log.Printf("error reading record: %+v", err)
			return nil, errs.New("failed to parse CSV: %s", err)
		}

		trimmed := strings.TrimSpace(record[0])
		quantity, err := strconv.ParseInt(trimmed, 0, 64)
		if err != nil {
			// handle error
			log.Printf("error parsing quantity: %+v\n", err)
			// assume quantity = 1
			quantity = 1
		}

		// NB: In the future, this should be optimized to be one query for all the cards
		// instead of a query for each card in the deck.
		name := record[1]
		card, err := s.Card(ctx, name, nil)
		if err != nil {
			// handle lookup error
			log.Printf("error looking up card: %+v\n", err)
			cards = append(cards, &Card{
				Name: name,
			})
			continue
		}

		if card == nil {
			fmt.Printf("failed to find card: %s", name)
		}

		// happy path
		var num int64 = 1
		for num <= quantity {
			// Fail gracefully if we can't find the card
			if len(card) == 0 {
				fmt.Printf("failed to find card- adding dummy card instead")
				cards = append(cards, &Card{
					Name: name,
				})
				num++
			} else {
				// add the first card that's returned from the database
				// NB: This is going to need to be handled eventually
				cards = append(cards, card[0])
				num++
			}
		}

		continue
	}

	return cards, nil
}

// GameKey formats the keys for Games in our Directory
func GameKey(gameID string) string {
	return fmt.Sprintf("%s", gameID)
}

// publish a game update to each Observer of the game
func (s *graphQLServer) publishGame(gameID string, g *Game) {
	s.mutex.Lock()
	if fullgame, ok := s.games[gameID]; ok {
		s.mutex.Unlock()
		for _, v := range fullgame.Observers {
			v.Channel <- g
		}
		return
	} else {
		s.mutex.Unlock()
		log.Printf("published update for game that does not exist: %s", gameID)
	}
}

// publish a boardstate update
func (s *graphQLServer) publishBoardstate(gameID string, userID string, bs *BoardState) {
	s.boardChannels[userID] <- bs
}
