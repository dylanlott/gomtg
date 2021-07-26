//go:generate go run github.com/99designs/gqlgen

package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/dylanlott/edh-go/persistence"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/segmentio/ksuid"
	"github.com/tinrab/retry"
	"github.com/zeebo/errs"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// graphQLServer binds the whole app together and implements the GraphQL interfac
type graphQLServer struct {
	mutex sync.RWMutex

	// TODO: Remove redisClient and switch over to our Persistence interface
	redisClient *redis.Client

	// Persistence layers
	kv     persistence.Persistence
	db     *sql.DB
	cardDB persistence.Database

	// games holds a reference to *FullGames in the server.
	games map[string]*FullGame

	// Channels per resource to achieve realtime.
	// Game and Board channels are required for a Game to be running.
	// TODO: Channels are only stored in memory.
	// If the server crashes, we'll need to setup all the channels for
	// existing games again.
	boardChannels   map[string]chan *BoardState
	messageChannels map[string]chan *Message
	userChannels    map[string]chan string
}

// Conf takes configuration values and loads them from the environment into our struct.
type Conf struct {
	RedisURL    string `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
	PostgresURL string `envconfig:"DATABASE_URL"`
	DefaultPort int    `envconfig:"PORT" default:"8080"`
}

// NewGraphQLServer creates a new server to attach the database, game engine,
// and graphql connections together
func NewGraphQLServer(
	kv persistence.KV,
	appDB *sql.DB,
	cardDB *sql.DB,
	cfg Conf,
) (*graphQLServer, error) {
	// TODO: Remove this redis client and wire chat up to KV interface instead
	log.Printf("attempting to connecto to redis at %s", cfg.RedisURL)
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("failed to get redis options: %s", err)
		return nil, errs.Wrap(err)
	}
	client := redis.NewClient(opts)

	retry.ForeverSleep(2*time.Second, func(_ int) error {
		_, err := client.Ping().Result()
		if err != nil {
			log.Printf("error connecting to redis: %+v\n", err)
		}
		return err
	})

	return &graphQLServer{
		mutex:           sync.RWMutex{},
		cardDB:          cardDB,
		db:              appDB,
		kv:              kv,
		redisClient:     client,
		games:           map[string]*FullGame{},
		messageChannels: map[string]chan *Message{},
		userChannels:    map[string]chan string{},
		boardChannels:   map[string]chan *BoardState{},
	}, nil
}

func (s *graphQLServer) Serve(route string, port int) error {
	mux := http.NewServeMux()
	mux.Handle(
		route,
		handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),
			handler.WebsocketUpgrader(websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}),
			handler.WebsocketKeepAliveDuration(time.Second*10),
		),
	)
	h := cors.AllowAll().Handler(s.auth(mux))
	mux.Handle("/playground", handler.Playground("GraphQL", route))
	log.Printf("serving graphiql at localhost:%d/playground", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), h)
}

// auth is a middleware responsible for passing header and cookie info to context
func (s *graphQLServer) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		// TODO: all of this code below needs to be vetted to find where
		// Gorilla's websocket is getting clobbered and fix that

		// username, err := r.Cookie("username")
		// if err != nil {
		// 	// log.Printf("username not present: %s", err)
		// 	// This means they're not attempting to auth, so we need to let them through
		// 	// to either sign up or login.
		// 	next.ServeHTTP(w, r)
		// 	return
		// }
		// token, err := r.Cookie("token")
		// if err != nil {
		// 	log.Printf("error parsing token: %s", err)
		// 	next.ServeHTTP(w, r)
		// }
		// if token == nil {
		// 	// allow unauthed users in for login/signup
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// // If they have a token, we have to compare it.
		// log.Printf("token: %s", token)
		// // find and authenticate
		// rows, err := s.db.Query("SELECT * FROM users WHERE username = ?", username)
		// if err != nil {
		// 	log.Printf("error querying users: %s", err)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// for rows.Next() {
		// 	cols, err := rows.Columns()
		// 	if err != nil {
		// 		log.Printf("error getting rows: %s", err)
		// 	}
		// 	log.Printf("columns: %s", cols)
		// 	continue
		// }

		// // TODO: Assign the found user to the request
		// // ctx := context.WithValue(r.Context(), userCtxKey, user)
		// // r = r.WithContext(ctx)
		// next.ServeHTTP(w, r)
	})
}

func (s *graphQLServer) PostMessage(ctx context.Context, user string, text string) (*Message, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	// Create message
	m := &Message{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		Text:      text,
		User:      user,
	}
	mj, _ := json.Marshal(m)

	// TODO: Update messages to key off of `message:<game_id>` and `message:<room_id>`
	if err := s.redisClient.LPush("messages", mj).Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	// Notify new message
	s.mutex.Lock()
	for _, ch := range s.messageChannels {
		ch <- m
	}
	s.mutex.Unlock()
	return m, nil
}

func (s *graphQLServer) Messages(ctx context.Context) ([]*Message, error) {
	cmd := s.redisClient.LRange("messages", 0, -1)
	if cmd.Err() != nil {
		log.Println(cmd.Err())
		return nil, cmd.Err()
	}
	res, err := cmd.Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	messages := []*Message{}
	for _, mj := range res {
		m := &Message{}
		err = json.Unmarshal([]byte(mj), &m)
		if err != nil {
			log.Printf("error unmarhsaling json Message: %s", err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *graphQLServer) Users(ctx context.Context, id *string) ([]string, error) {
	return nil, errors.New("not impl")
}

func (s *graphQLServer) MessagePosted(ctx context.Context, user string) (<-chan *Message, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	fmt.Printf("\n received messagePosted: %+v\n", user)

	// Create new channel for request
	messages := make(chan *Message, 1)
	s.mutex.Lock()
	s.messageChannels[user] = messages
	s.mutex.Unlock()

	// Delete channel when done
	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		delete(s.messageChannels, user)
		s.mutex.Unlock()
	}()

	return messages, nil
}

func (s *graphQLServer) UserJoined(ctx context.Context, user string, gameID string) (<-chan string, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	fmt.Printf("userJoined: %s", user)

	// Create new channel for request
	users := make(chan string, 1)
	s.mutex.Lock()

	// userChannels is a map of usernames to the channel we just created
	s.userChannels[user] = users
	s.mutex.Unlock()

	// Wait for the Done event to fire, then clean up.
	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		delete(s.userChannels, user)
		s.mutex.Unlock()
	}()

	// Return the channel we created loaded with its cleanup instructions.
	return users, nil
}

func (s *graphQLServer) createUser(user string) error {
	// Upsert user
	if err := s.redisClient.SAdd("users", user).Err(); err != nil {
		return err
	}
	// Notify new user joined
	s.mutex.Lock()
	for _, ch := range s.userChannels {
		ch <- user
	}
	s.mutex.Unlock()
	return nil
}

func (s *graphQLServer) Mutation() MutationResolver {
	return s
}

func (s *graphQLServer) Query() QueryResolver {
	return s
}

func (s *graphQLServer) Subscription() SubscriptionResolver {
	return s
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}
