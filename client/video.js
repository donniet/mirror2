
Vue.component('videos', {
  props: ['videolist', 'videowidth', 'videoheight'],
  methods: {},
  render: function(createElement) {
    console.log('videos: ', this.videolist);
    if (!this.videolist) return;

    var elements = this.videolist.map(vid => {
      console.log('adding MJPEG video');
      return createElement('mjpegVideo', {
        props: {
          videoUrl: vid.url,
          width: this.videowidth,
          height: this.videoheight,
          hidden: !vid.visible,
        }
      });
    });

    return createElement('div', {
      'class': {
        'videos': true
      }
    }, elements);
  }
});


Vue.component('mjpegVideo', {
  data: function() {
    return {};
  },
  computed: {
    'display': function() {
      if (!this.hidden) {
        return 'block';
      }
      return 'none';
    },
    'finalUrl': function() {
      if (!this.hidden) {
        return this.videoUrl;
      }
      return '';
    }
  },
  props: ['videoUrl', 'width', 'height', 'hidden' ],
  render: function(createElement) {
    return createElement('div',
      {
        'style': {
          'display': this.display
        }
      },
      [
        createElement('img', {
          attrs: {
            src: this.finalUrl,
            width: this.width,
            height: this.height
          }
        })
      ]
    );
  },
  methods: {
    check: function() {
      checkImageAsync(this.videoUrl).then(ret => {
        console.log('check result', ret);
        this.shouldDisplay = !!ret;
        setTimeout(() => { this.check(); }, 5000);
      });
    }
  },
  mounted: function() {
    // this.check();
  }
});
