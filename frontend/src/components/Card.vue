<template>
  <div
    v-bind:class="{
      'd-none': hidden,
      tapped: Tapped,
      flipped: flipped,
    }"
  >
    <!-- # TODO: Get images working for cards, but for now text will do.
    // Maybe hovering on a card should expose the
    // This should be the right dimensions so we can scale them up or down. 
      <img class="card-img-top" src="https://via.placeholder.com/750x1050.jpg">
    -->

    <div class="card fixed-size">
      <!-- TODO: Show this modal view if in text-only mode -->
      <!-- TODO: Show a card-image class div here with the image -->
      <div class="card-content">
        <p class="title is-5">{{ Name }}</p>
        <p class="is-subtitle">{{ ManaCost }}</p>
      </div>

      <div class="">
        <b-collapse class="card" animation="slide" aria-id="contentIdForA11y3">
          <template #trigger="props">
            <div class="card-header" role="button" aria-controls="contentIdForA11y3">
              <!-- It should go supertypes, types, subtypes. -->
              <p class="card-header-title font-medium">{{ Supertypes }} {{ Types }} {{ Subtypes }}</p>
              <a class="card-header-icon">
              <!-- TODO: Make this icon show correctly -->
                <b-icon :icon="props.open ? 'fa-solid fa-chevron-up' : 'fa-close'"> </b-icon>
              </a>
            </div>
          </template>

          <div class="card-content">
            {{ Text }}
          </div>
        </b-collapse>

        <!-- <div>
          <i class="font-medium card-types"> {{ Types }} {{ Subtypes }} {{ Supertypes }} </i>
        </div> -->
      </div>
      <!-- Collapse ends -->

      <footer class="card-footer">
        <p class="card-footer-item">
          <span>
            <a href="#">Send To <i class=""></i></a>
          </span>
        </p>
        <p class="card-footer-item">
          <span><i class="fas fa-sync-alt"></i> <a href="#">Tap</a></span>
        </p>
      </footer>
    </div>
  </div>
</template>
<script>
export default {
  name: 'Card',
  data() {
    return {
      hidden: false, // if a card can be seen at all - visibility off
      flipped: false, // if a card is upside down or not
      trackers: {}, // player-assigned trackers
      labels: {}, // player assigned labels
      counters: {}, // game-assigned counters such as poison or infect
      reminders: {}, // untap effects, etb effect reminders, etc...
    };
  },
  props: [
    'ID',
    'Name',
    'CMC',
    'ManaCost',
    'colorIdentity',
    'Power',
    'Toughness',
    'Text',
    'Types',
    'Supertypes',
    'Subtypes',
    'ScryfallID',
    'Tapped',
    'Flipped',
  ],
};
</script>
<style scoped>
.fixed-size {
  width: 275px;
  max-width: 275px;
}
.flipped {
  color: #000;
}
.column {
  padding: 0rem;
}
.card-content {
  padding: 0.7rem;
}
</style>