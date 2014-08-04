// HTTP URI new data is requested from
var uri = "http://127.0.0.1:3000/api/robots/brain/devices/neurosky/events/eeg";

// placeholder data to populate the graph when we start
var data = {
  delta: 1079410,
  theta: 3226112,
  loAlpha: 3550720,
  hiAlpha: 3225856,
  loBeta: 6103040,
  hiBeta: 4804096,
  loGamma: 5185536,
  midGamma: 2957370
};
 
var convertDataToArray = function(data) {
  var arr = [];
  for (var i in data) {
    if (i.toLowerCase() === "delta") {
      arr.push([i, data[i].toString().substring(2)])
    } else {
      arr.push([i, data[i]]);
    }
  }
  return arr;
};
 
data = convertDataToArray(data);
 
// graph config
var config = {
  barWidth: 140,
  width: 1280,
  height: 720,
  animationTime: 300
};
 
var x = d3.scale.linear().domain([0, data.length]).range([0, config.width]);
var y = d3.scale.linear().domain([0, 20000000]).range([0, config.height]);
 
// we treat the delta wave different so it actually shows movement
var deltaY = d3.scale.linear().domain([70000, 90000]).range([0, config.height]);
 
var chart = d3.select(".chart")
  .attr("width", config.width)
  .attr("height", config.height);
 
var bar = chart.selectAll("g")
    .data(data)
  .enter().append("g")
    .attr("transform", function(d, i) { return "translate(" + x(i) + ",0)"; });
 
bar.append("rect")
  .attr("class", function(datum) { return datum[0]; })
  .attr("width", config.barWidth);
 
bar.append("text")
  .attr("x", config.barWidth * 0.5)
  .attr("text-anchor", "middle")
  .attr("dy", ".35em")
  .text(function(datum) { return datum[0]; })
  .attr("fill", "white")
 
// function to render the parts of the chart that change with new data
var render = function(data) {
  var bar = chart.selectAll("g").data(data);
 
  bar.select("rect")
    .transition()
    .duration(config.animationTime)
    .attr("y", function(datum) {
      // special handling of delta waves
      if (datum[0].toLowerCase() === "delta") {
        return config.height - deltaY(datum[1]);
      }
 
      return config.height - y(datum[1]);
    })
    .attr("height", function(datum) {
 
      // special handling of delta waves
      if (datum[0].toLowerCase() === "delta") {
        return deltaY(datum[1]);
      }
 
      return y(datum[1]);
    });
 
  bar.select("text")
    .transition()
    .duration(config.animationTime)
    .attr("y", function(datum) {
      // special handling of delta waves
      if (datum[0].toLowerCase() === "delta") {
        return config.height - deltaY(datum[1]) + 20;
      }
 
      return config.height - y(datum[1]) + 20;
    })
};
 
render(data);


var es = new EventSource(uri)
                                                                                                                                                                              
es.addEventListener('message', function(message) {                                                                                                                        
  var newData = convertDataToArray(JSON.parse(message.data));
  render(newData);
}, false);  
