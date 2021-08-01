<template>
  <div class="join shell">
    <!-- <pre>{{ game }}</pre> -->
    <div class="container">
      <div class="columns is-mobile is-centered">
        <div class="column is-half">
          <h1 class="title is-1">Join Game</h1>
          <p>There are {{ game.PlayerIDs.length }} other players in this game.</p>
          <p class="title is-4">Pick your commander</p>
          <b-field label="Select a Commander">
            <b-autocomplete
              :data="data"
              v-model="deck.commander"
              clearable
              field="name"
              placeholder="e.g. Jarad, Golgari Lich Lord"
              @typing="queryCommanders"
              @select="(option) => (selected = option)"
            >
              <template slot="empty">No results found</template>
            </b-autocomplete>
          </b-field>
          <h3 class="title is-4">Add the 99.</h3>
          <p>
            We recommend using <a href="www.archidekt.com">Archidekt</a> to generate your decklists so that spelling
            errors and quantities aren't an issue. Select "CSV" on Export.
          </p>
          <br />
          <p>
            <b>Note</b>: There must be exactly 99 cards in this list, they need to be spelled exactly correct, and there
            can't be duplicates except for Basic Lands.
          </p>
          <br />
          <p>Enter cards in CSV Format: <code>1, Card Name</code></p>
          <br />
          <b-field label="Paste your decklist here.">
            <b-input maxlength="200000" v-model="deck.library" type="textarea"></b-input>
          </b-field>

          <div v-if="!user.Username">
            <b-field label="Add a username?">
              <b-input v-model="username"></b-input>
            </b-field>
          </div>

          <b-button @click="handleJoinGame()" type="button" class="is-success">Join Game</b-button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import gql from 'graphql-tag';
import { mapState } from 'vuex';
import { debounce } from 'lodash';

export default {
  name: 'join',
  data() {
    return {
      data: [],
      loading: false,
      deck: {
        library: '',
        commander: '',
        decklist: '',
      },
      username: '',
    };
  },
  computed: {
    commanderList() {
      this.queryCommanders()
        .then((data) => {
          return data;
        })
        .catch((err) => console.error(err));
    },
    ...mapState({
      game: (state) => state.Game.game,
      user: (state) => state.User.User,
    }),
  },
  created() {
    this.$store.dispatch('getGame', this.$route.params.id);
  },
  methods: {
    queryCommanders() {
      if (this.deck.commander === '') {
        return;
      }
      this.$apollo
        .query({
          query: gql`
            query ($name: String!) {
              search(name: $name) {
                Name
                ID
                Colors
                ColorIdentity
                CMC
                ManaCost
              }
            }
          `,
          variables: {
            name: this.deck.commander,
          },
        })
        .then((resp) => {
          this.data = resp.data.search.map((item) => {
            return { name: item.Name };
          });
          return this.data;
        })
        .catch((err) => {
          console.error('error querying commanders: ', err);
          return err;
        })
        .finally(() => {
          this.loading = false;
        });
    },
    handleJoinGame() {
      this.$store.dispatch('joinGame', {
        inputGame: {
          ID: this.$route.params.id,
          Decklist: this.deck.library,
          User: {
            ID: this.user.ID,
            Username: this.user.Username,
          },
          BoardState: {
            GameID: this.$route.params.id,
            User: {
              Username: this.user.Username,
              ID: this.user.ID,
            },
            Life: 40,
            Commander: [
              {
                Name: this.deck.commander,
              },
            ],
          },
        },
      });
    },
    uuid() {
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        var r = (Math.random() * 16) | 0,
          v = c == 'x' ? r : (r & 0x3) | 0x8;
        return v.toString(16);
      });
    },
  },
};
</script>
<style>
</style>