
var days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
var months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']

function formatDate(d) {
  var hour = d.getHours();
  var minute = d.getMinutes();
  var ampm = 'am';
  if(hour >= 12) {
    ampm = 'pm';
  }
  if(hour == 0) {
    hour = 12;
  } else if(hour > 12) {
    hour -= 12;
  }
  if(minute < 10) {
    minute = '0' + minute;
  }
  return days[d.getDay()] + ' ' + months[d.getMonth()] + ' ' + d.getDate() + numberSuffix(d.getDate()) + ' ' + hour + ':' + minute + ampm;
}

function numberSuffix(num) {
  // num = num << 0;

  var units = num % 10;
  var tens = Math.floor(num / 10) % 10;
  if (units == 1 && tens != 1) return 'st';
  if (units == 2 && tens != 1) return 'nd';
  if (units == 3 && tens != 1) return 'rd';
  return 'th';
}

Vue.component('clock', {
  data: function() {
    return {
      time: new Date(),
      formattedTime: ''
    }
  },
  methods: {
    updateTime: function() {
      this.time = new Date();
      this.formattedTime = formatDate(this.time);
    }
  },
  mounted: function() {
    updater = () => {
      this.updateTime();
      var seconds = new Date().getSeconds();
      setTimeout(updater, 1000*(60 - seconds));
    }
    updater();
  }
});
