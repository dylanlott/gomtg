package server

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBoardState(t *testing.T) {
	userID := string("0xACAB")

	type args struct {
		ctx   context.Context
		input InputBoardState
	}
	tests := []struct {
		name    string
		args    args
		want    *BoardState
		wantErr bool
	}{
		{
			name: "should update boardstate and notify listeners",
			args: args{
				ctx: context.Background(),
				input: InputBoardState{
					GameID: seedGameID,
					User: &InputUser{
						Username: "shakezula",
						ID:       &userID,
					},
					Life: 38, // test life edits
					Commander: []*InputCard{
						{Name: "Gavi, Nest Warden"},
					},
				},
			},
			wantErr: false,
			want: &BoardState{
				GameID: seedGameID,
				User: &User{
					ID:       userID,
					Username: "shakezula",
				},
				Life: 38,
				Commander: []*Card{
					{Name: "Gavi, Nest Warden"},
				},
			},
		},
		{
			name: "should return an error if game does not exist",
			args: args{
				ctx: context.Background(),
				input: InputBoardState{
					GameID: "doesnotexist",
					User: &InputUser{
						Username: "shakezula",
						ID:       &userID,
					},
					Life: 38, // test life edits
					Commander: []*InputCard{
						{Name: "Gavi, Nest Warden"},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			_, err := s.CreateGame(tt.args.ctx, *seedInputGame)
			if err != nil {
				t.Errorf("failed to create dummy game: %s", err)
			}

			// register the BoardState channel for the userID
			bch, err := s.BoardstateUpdated(
				tt.args.ctx,
				"TestUpdateBoardStateObserver",
				*tt.args.input.User.ID,
			)
			assert.Equal(t, err, nil)
			assert.NotNil(t, bch)
			time.Sleep(time.Second * 1)
			got, err := s.UpdateBoardState(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.UpdateBoardState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("graphQLServer.UpdateBoardState() = %v, want %v", got, tt.want)
			}

			// assert on boardstate notifications if no error is expected
			if tt.wantErr == false {
				boardstate := <-bch
				if !reflect.DeepEqual(boardstate, tt.want) {
					diff := cmp.Diff(boardstate, tt.want)
					t.Logf("DIFF: %+v", diff)
					t.Errorf("UpdateBoardState failed to notify listeners")
				} else {
					t.Logf("successfully received boardstate from update: %+v", boardstate)
				}
			}
		})
	}
}

func TestMultipleObservers(t *testing.T) {
	s := testAPI(t)
	ctx := context.Background()
	ctx, done := context.WithCancel(ctx)
	created, err := s.CreateGame(ctx, *seedInputGame)
	assert.NoError(t, err)
	// NB: these should both be observing the same user from *seedInputGame
	ch1, err := s.BoardstateUpdated(ctx, "testobs1", seedUserID)
	assert.NoError(t, err)
	assert.NotNil(t, ch1)
	ch2, err := s.BoardstateUpdated(ctx, "testobs2", seedUserID)
	assert.NoError(t, err)
	assert.NotNil(t, ch2)

	// game has 1 player, so let's update that player's boardstate and see
	// if the other two channels are alerted to that
	all, err := s.Boardstates(ctx, created.ID, nil)
	assert.NoError(t, err)
	assert.NotNil(t, all)
	p1 := all[0]
	p1.Life = 20
	in, err := inputFromBoardState(*p1)
	assert.NoError(t, err)
	assert.NotNil(t, in)

	// fire off the update to trigger a boardstate update event
	updated, err := s.UpdateBoardState(ctx, in)
	assert.NoError(t, err)
	assert.NotNil(t, updated)

	// listen for results and compare to desired boardstates
	bs1 := <-ch1
	bs2 := <-ch2
	assert.NotNil(t, bs1)
	assert.NotNil(t, bs2)
	assert.Equal(t, bs1, updated)
	assert.Equal(t, bs2, updated)
	done()
}
