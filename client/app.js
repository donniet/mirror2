
document.addEventListener('DOMContentLoaded', function() {
  document.app = new Vue({
    el: '#vue',
    data: {
      weatherUrl: null,
      lastModified: '',
      backgroundInterval: null,
      videos: [],
      news: {},
      weather: {},
      youtube: {},
      bg: '',
      socket: null,
      clientWidth: 1920,
      clientHeight: 1080,
      socketUrl: "",
    },
    methods: {
      updateYoutubePlayerState: function(data) {
        console.log('updating youtube player state', data);
        this.socket.send(JSON.stringify({youtube: data}));
      },
      setVideoHidden: function(video, hidden) {
        console.log('setVideoHidden', video, hidden);
        postAsync('/api/activeSources', {
          url: video.url,
          hidden: hidden,
        }).then((m) => console.log('output:', m));
      },
      setWeatherHidden: function(hidden) {
        console.log('setWeatherHidden', hidden);
        postAsync('/api/weather', {
          hidden: hidden,
        }).then((m) => console.log('output:', m));
      },
      setNewsHidden: function(hidden) {
        console.log('setNewsHidden', hidden);
        postAsync('/api/news', {
          hidden: hidden,
        }).then((m) => console.log('output:', m));
      },
      sendYoutube: function(videoId, paused) {
        console.log('sendYoutube', videoId, !!paused);

        let y = JSON.parse(JSON.stringify(this.youtube));
        y.videoId = videoId;
        y.pause = !!paused;
        postAsync('/api/youtube', y).then(m => console.log('output:', m));
      },
      setYoutubeVolume: function(volume) {
        console.log('setYoutubeVolume', volume);
        let y = JSON.parse(JSON.stringify(this.youtube));
        y.volume = volume >>> 0;
        postAsync('/api/youtube', y).then((m) => console.log('output:', m));
      },
      setYoutubeMute: function(mute) {
        console.log('setYoutubeMute', mute);
        let y = JSON.parse(JSON.stringify(this.youtube));
        y.mute = !!mute;
        postAsync('/api/youtube', y).then((m) => console.log('output:', m));
      },
      setYoutubeHidden: function(hidden) {
        console.log('setYoutubeHidden', hidden);
        let y = JSON.parse(JSON.stringify(this.youtube));
        y.paused = !!hidden;
        y.fullScreen = !hidden;
        y.hidden = !!hidden;
        postAsync('/api/youtube', y).then(m => console.log('output:', m));
      },
      setYoutubeFullScreen: function(fullScreen) {
        console.log('setYoutubeFullScreen', fullScreen);
        let y = JSON.parse(JSON.stringify(this.youtube));
        y.fullScreen = !!fullScreen;
        y.hidden = y.hidden && !fulScreen;

        postAsync('/api/youtube', y).then(m => console.log('output:', m));
      },
      setYoutubePaused: function(paused) {
        console.log('setYoutubePaused', paused);
        let y = JSON.parse(JSON.stringify(this.youtube));
        y.paused = !!paused;
        postAsync('/api/youtube', y).then(m => console.log('output:', m));
      },
      updateBackground: function() {
        this.bg = this.backgrounds[this.currentBackground];
        this.currentBackground = (this.currentBackground + 1) % this.backgrounds.length;
      },
      socketMessage: function(msg) {
        let o;
        try {
          o = JSON.parse(msg.data);
        } catch(ex) {
          console.log(ex);
          return
        }

        for (var k in o) {
          switch(k) {
          case "activeVideos":
            if (JSON.stringify(this.videos) != JSON.stringify(o[k])) {
              this.videos = o[k];
            }
            break;
          case "news":
            if (JSON.stringify(this.news) != JSON.stringify(o[k])) {
              this.news = o[k];
            }
            break;
          case "weather":
            if (JSON.stringify(this.weather) != JSON.stringify(o[k])) {
              this.weather = o[k];
            }
            break;
          case "youtube":
            if (JSON.stringify(this.youtube) != JSON.stringify(o[k])) {
              this.youtube = o[k];
            }
            break;
          case "config":
            break;
          }
        }
      },
      socketClose: function(notimeout) {
        this.socketOpen = false;
        if (this.socket) {
          this.socket.onclose = null;
          this.socket.onmessage = null;
          this.socket.onerror = null;
          if (this.socket.readyState == WebSocket.CONNECTING || this.socket.readyState == WebSocket.OPEN)
            this.socket.close();
        }
        setTimeout(() => {
          this.socket = new WebSocket(this.socketUrl);

          console.log('setting timeout');
          setTimeout(() => {
            console.log('ws readyState', this.socket);
            if (this.socket.readyState != 1) {
              console.log('retrying');
              this.socketClose();
            }
          }, 5000);

          this.socket.onopen = () => {
            this.socket.onmessage = msg => this.socketMessage(msg);
            this.socket.onerror = err => this.socketError(err);
            this.socket.onclose = () => this.socketClose();

            this.socketOpen = true;

            this.socket.send(JSON.stringify({Path: "/"}))
          };
        }, notimeout ? 0 : 1000);
      },
      socketError: function(err) {
        console.log('socket error', err);
        this.socketClose();
      },

    },
    mounted: function() {
      console.log('mounted');
      this.socketUrl = this.$el.attributes.getNamedItem("socket-url").value;
      this.clientWidth = window.innerWidth;
      this.clientHeight = window.innerHeight;
      this.socketClose(true);
    }
  });
});
