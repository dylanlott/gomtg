<template>
  <div class="container is-fluid" v-if="user && game">
    <!-- TURN TRACKER -->
    <div class="box"><TurnTracker :game="game" /></div>
    <!-- END TURN TRACKER -->

    <!-- ### OPPONENTS BOARDSTATES ### -->
    <div class="box" :key="player.ID" v-for="player in bs">
      <div class="columns" v-if="user.ID !== player.User.ID">
        <div class="title">{{ player.User.Username }}</div>
        <div class="battlefield">
          <div class="columns" v-if="player">
            <div class="column">
              <draggable
                class="columns is-mobile"
                v-model="player.Field"
                group="people"
                @start="drag = true"
                @end="drag = false"
              >
                <div class="column" v-for="card in player.Field" :key="card.id">
                  <Card v-bind="card"></Card>
                </div>
              </draggable>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- ### END OF OPPONENTS BOARDSTATES ### -->

    <div class="tabs">
      <ul>
        <li class="is-active"><a>Battlefield</a></li>
        <li><a>Music</a></li>
        <li><a>Videos</a></li>
        <li><a>Documents</a></li>
      </ul>
    </div>

    <!-- ### SELF BOARDSTATE - PUBLIC SECTION ### -->
    <div class="columns is-flex" v-if="self">
      <!-- SELF - BATTLEFIELD -->
      <div class="column box is-align-content-flex-start">
        <p class="title is-4">Battlefield</p>
        <draggable
          class="columns is-mobile"
          @change="handleChange()"
          v-model="self.Field"
          group="people"
          @start="drag = true"
          @end="drag = false"
        >
          <!-- Note: Cards can only be tapped on the battlefield -->
          <div v-on:dblclick="handleTap(card)" class="column" v-for="card in self.Field" :key="card.id">
            <Card v-bind="card"></Card>
          </div>
        </draggable>
      </div>
    </div>
    <!-- END SELF BATTLEFIELD -->

    <!-- ### TOOLBAR START  -->
    <template>
      <b-navbar>
        <template #start>
          <b-navbar-item @click="handleDraw()" href="#">
            <button class="button is-primary">
              <span class="icon">
                <i class="fa fa-book"></i>
              </span>
              <span>Draw</span>
            </button>
          </b-navbar-item>
        </template>

        <template #end>
          <b-navbar-item tag="div">
            <div class="buttons">
              <a @click="handleTapAll()" class="button is-dark"> <strong>Tap All</strong> </a>
              <a @click="handleUntapAll()" class="button is-light"> Untap All </a>
            </div>
          </b-navbar-item>
        </template>
      </b-navbar>
    </template>

    <!-- FOOTER  -->
    <footer class="footer">
      <div class="content">
        <div class="column is-full">
          <!-- SELF - HAND-->
          <p class="title is-5">Hand</p>
          <draggable
            class="columns is-flex is-multiline is-mobile is-align-items-flex-start"
            v-model="self.Hand"
            group="people"
            @change="handleChange()"
            @start="drag = true"
            @end="drag = false"
          >
            <div class="column" v-for="card in self.Hand" :key="card.id">
              <Card v-bind="card"></Card>
            </div>
          </draggable>
        </div>
      </div>
    </footer>
  </div>
</template>
  </div>
</template>
<script>
import _ from 'lodash';
import draggable from 'vuedraggable';
import Card from '@/components/Card';
import TurnTracker from '@/components/TurnTracker.vue';
import { mapState } from 'vuex';

export default {
  name: 'board',
  data() {
    return {};
  },
  created() {
    // load and sub to game
    // this.$store.dispatch('getGame', this.$route.params.id)
    this.$store.dispatch('subscribeToGame', {
      gameID: this.$route.params.id,
      userID: this.user.ID,
    });
    // load and sub to all boardstates
    this.$store.dispatch('subAllBoardstates', {
      gameID: this.$route.params.id,
      obsID: this.user.ID,
    });
  },
  computed: mapState({
    game: (state) => state.Game.game,
    bs: (state) => state.BoardStates.boardstates,
    self: (state) => state.BoardStates.self,
    user: (state) => state.User.User,
  }),
  methods: {
    handleDraw() {
      this.$store.dispatch('draw', this.self);
    },
    handleChange() {
      this.$store.dispatch('mutateBoardState', this.self);
    },
    handleTap(card) {
      // TODO: Make this a vuex boardstate action
      card.Tapped = !card.Tapped;
      this.handleChange();
    },
    handleTapAll() {
      this.$store.dispatch('tapAll', this.self);
    },
    handleUntapAll() {
      this.$store.dispatch('untapAll', this.self);
    },
  },
  components: {
    draggable,
    Card,
    TurnTracker,
  },
};
</script>
<style media="screen" scoped>
.shell {
  padding: 0.5rem;
  border: 1px solid #efefef;
  margin: 0.25rem 0rem;
}

.bordered {
  border: 1px #000;
}
</style>
