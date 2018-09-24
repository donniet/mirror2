

document.__youtubeAPICallbacks = [];

if (typeof onYouTubeIframeAPIReady === 'function') {
  document.__youtubeAPICallbacks.push(onYouTubeIframeAPIReady);
}
onYouTubeIframeAPIReady = function() {
  console.log('youtube api ready.');
  document.__youtubeAPICallbacks.forEach(callback => callback());
}

var __playerId = 0;

Vue.component('youtube', {
  data: function() {
    return {
      headline: '',
      currentHeadline: 0,
      headlineTimeout: 15 * 1000,
      headliner: null,
      youtubeAPI: 'https://www.youtube.com/iframe_api',
      playerId: 0,
      player: null,
      playerReady: false,
      className: '',
    }
  },
  props: [ 'width', 'height', 'data' ],
  template: '<div :class="className" v-show="!data.hidden"><div :id="playerId"></div></div>',
  methods: {
    ytAPILoaded: function() {
      this.playerStateLookup = {};
      Object.getOwnPropertyNames(YT.PlayerState).forEach(k => {
        this.playerStateLookup[YT.PlayerState[k]] = k;
      });
      this.player = new YT.Player(this.playerId, {
        height: this.height,
        width: this.width,
        playerVars: { 'autoplay': 1, 'controls': 0, 'showinfo': false, 'modestbranding': 1 },
        events: {
          'onReady': () => this.onPlayerReady(),
          'onStateChange': evt => this.onPlayerStateChange(evt),
        }
      });
    },
    onPlayerReady: function() {
      this.playerReady = true;
      console.log('player ready', this.player);

      this.handleDataChange(this.data);
    },
    onPlayerStateChange: function(evt) {
      if (!this.playerReady) return;

      console.log('state change', evt);
      var newState = this.playerStateLookup[evt.data];
      if (this.data.playerState != newState) {
        this.data.playerState = newState;
        this.$emit('player-state-change', this.data);
      }
    },
    handleDataChange: function(data, oldData) {
      if (!data || !this.player) return;

      console.log('data change', data);

      if (data.videoId && (!oldData || data.videoId != oldData.videoId)) {
        this.player.loadVideoById({
          videoId: data.videoId,
        });
      }
      if (!data.paused) {
        this.player.playVideo();
      } else {
        this.player.pauseVideo();
      }
      // this.setFullScreen(data.fullScreen);
      if (data.fullScreen) {
        this.className = 'full-screen';
        // this.player.setSize(1080, 1920);
      } else {
        this.className = 'hidden';
        // this.player.setSize(this.width, this.height);
      }
      if (data.mute) {
        this.player.mute();
      } else {
        this.player.unMute();
      }
      if (typeof data.volume === 'number') {
        this.player.setVolume(data.volume);
      }
    },
  },
  watch: {
    'data': function(data, oldData) {
      this.handleDataChange(data, oldData);
    },
  },
  mounted: function() {

    this.playerId = 'youtube' + (__playerId++);
    if (typeof YT === 'undefined') {
      if (!document.__youtubeAPILoading) {
        document.__youtubeAPILoading = true;
        var s = document.createElement('script');
        s.type = 'text/javascript';
        s.src = this.youtubeAPI;
        var firstScriptTag = document.getElementsByTagName('script')[0];
        firstScriptTag.parentNode.insertBefore(s, firstScriptTag);
      }
      document.__youtubeAPICallbacks.push(() => this.ytAPILoaded());
    } else {
      this.ytAPILoaded();
    }
  },
});
