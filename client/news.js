
var daysOfWeek = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

function formatTime(d) {
  var c = new Date();
  var midnight = new Date(c.getFullYear(), c.getMonth(), c.getDate());

  var day = "";
  if(d.getTime() > midnight.getTime()) {
    day += "Today";
  } else if (midnight.getTime() - d.getTime() < 1000*60*60*24){
    day += "Yesterday";
  } else if (midnight.getTime() - d.getTime() < 7*1000*60*60*24) {
    day += daysOfWeek[d.getDay()];
  } else if (midnight.getTime() - d.getTime() < 14*1000*60*60*24) {
    day += "Last " + daysOfWeek[d.getDay()];
  } else {
    day += months[d.getMonth()] + " " + d.getDate();
  }

  var hours = d.getHours();
  var minutes = d.getMinutes();
  var ampm = "am";
  if (hours > 12) {
    hours -= 12;
    ampm = "pm";
  }
  if (hours == 0) {
    hours = 12;
  }
  if (minutes < 10) {
    minutes = "0" + minutes;
  }
  return day + " " + hours + ":" + minutes + ampm
}

function formatNews(item) {
  var d = new Date(Date.parse(item.publishDate));
  return formatTime(d) + ": " + item.title + " - " + item.source;
}

Vue.component('news', {
  data: function() {
    return {
      headline: '',
      currentHeadline: 0,
      headlineTimeout: 15 * 1000,
      headliner: null,
    }
  },
  props: [ 'visible', 'items' ],
  watch: {
    items: function(oldValue, newValue) {
      this.nextHeadline();
    }
  },
  methods: {
    nextHeadline: function() {
      if (this.headliner) clearTimeout(this.headliner);

      if (!this.items || this.items.length == 0) {
        this.currentHeadline = 0;
        this.headline = '';
      } else {
        this.currentHeadline = (this.currentHeadline + 1) % this.items.length;
        this.headline = formatNews(this.items[this.currentHeadline]);
      }
      this.headliner = setTimeout(() => this.nextHeadline(), this.headlineTimeout);
    },
  },
  mounted: function() {
    this.nextHeadline();
    // this.refreshResponse();
  }
});
