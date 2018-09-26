
var frown = `<?xml version="1.0" ?>
<svg height="200px" viewBox="0 0 1792 1792" width="200px" xmlns="http://www.w3.org/2000/svg"><path d="M1262 1229q8 25-4 48.5t-37 31.5-49-4-32-38q-25-80-92.5-129.5t-151.5-49.5-151.5 49.5-92.5 129.5q-8 26-31.5 38t-48.5 4q-26-8-38-31.5t-4-48.5q37-121 138-195t228-74 228 74 138 195zm-494-589q0 53-37.5 90.5t-90.5 37.5-90.5-37.5-37.5-90.5 37.5-90.5 90.5-37.5 90.5 37.5 37.5 90.5zm512 0q0 53-37.5 90.5t-90.5 37.5-90.5-37.5-37.5-90.5 37.5-90.5 90.5-37.5 90.5 37.5 37.5 90.5zm256 256q0-130-51-248.5t-136.5-204-204-136.5-248.5-51-248.5 51-204 136.5-136.5 204-51 248.5 51 248.5 136.5 204 204 136.5 248.5 51 248.5-51 204-136.5 136.5-204 51-248.5zm128 0q0 209-103 385.5t-279.5 279.5-385.5 103-385.5-103-279.5-279.5-103-385.5 103-385.5 279.5-279.5 385.5-103 385.5 103 279.5 279.5 103 385.5z"/></svg>
`

var regexSequences = [
    // Remove XML stuffs and comments
    [/<\?xml[\s\S]*?>/gi, ""],
    [/<!doctype[\s\S]*?>/gi, ""],
    [/<!--.*-->/gi, ""],

    // SVG XML -> HTML5
    [/\<([A-Za-z]+)([^\>]*)\/\>/g, "<$1$2></$1>"], // convert self-closing XML SVG nodes to explicitly closed HTML5 SVG nodes
    [/\s+/g, " "],                                 // replace whitespace sequences with a single space
    [/\> \</g, "><"]                               // remove whitespace between tags
];

Vue.component('weather', {
  data: function() {
    return {
      svgContent: '<svg>blah</svg>',
    };
  },
  props: ['weather'],
  computed: {
    formattedLow: function() {
      return 'low ' + this.lowTemp + '\u00B0';
    },
    formattedHigh: function() {
      return 'high ' + this.highTemp + '\u00B0';
    },
    lowTemp: function() {
      return this.weather ? this.weather.low : -459.67;
    },
    highTemp: function() {
      return this.weather ? this.weather.high : -459.67;
    },
    icon: function() {
      return this.weather ? this.weather.icon : '';
    },
  },
  watch: {
    icon: function(newIcon, oldIcon) {
      if (newIcon == '') {
        this.svgContent = frown;
        return;
      }

      getXMLAsync('/client/SVG/' + newIcon + '.svg').then(svg => {
        try {
          svg.documentElement.setAttribute('width', '200px');
          svg.documentElement.setAttribute('height', '133px');
          svg.documentElement.setAttribute('viewBox', '12 15 88 65');
          this.setSVGContent((new XMLSerializer()).serializeToString(svg));
        } catch(ex) {
          this.setSVGContent(frown);
        }
      });
    },
  },
  methods: {
    setSVGContent: function(svgStr) {
      this.svgContent = regexSequences.reduce(function (prev, regexSequence) {
        return ''.replace.apply(prev, regexSequence);
      }, svgStr).trim();
    },
  },
});
