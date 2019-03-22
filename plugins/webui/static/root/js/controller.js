var app;
const timeoutMax = 30000;
const timeoutMin = 1000;
var timeout = timeoutMin;

function loadContent() {
  app = new Vue({
    el: "#app",
    data: {
      error: {
        has_error: false,
        error_message: ""
      },
      state: {
        output_current: null,
        output_voltage: 0,
        output_frequency: 0,
        output_power: 0,
        input_current: 0,
        input_voltage: 0,
        input_frequency: 0,
        input_power: 0,
        battery_current: 0,
        battery_voltage: 0,
        battery_charge: 0,
        battery_power: 0,
        led_map: [
          { led_mains: "dot-off" },
          { led_absorb: "dot-off" },
          { led_bulk: "dot-off" },
          { led_float: "dot-off" },
          { led_inverter: "dot-off" },
          { led_overload: "dot-off" },
          { led_bat_low: "dot-off" },
          { led_over_temp: "dot-off" }
        ]
      }
    }
  });

  connect();
}

function connect() {
  if (window["WebSocket"]) {
    var conn = new WebSocket(getURI());
    conn.onclose = function(evt) {
      app.error.has_error = true;
      app.error.error_message =
        "Server not reachable. Trying to reconnect in " +
        timeout / 1000 +
        " second(s).";

      console.log(app.error.error_message, evt.reason);
      setTimeout(function() {
        connect();
      }, timeout);
      timeout = timeout * 2;
      if (timeout > timeoutMax) {
        timeout = timeoutMax;
      }
    };

    conn.onopen = function(evt) {
      timeout = timeoutMin;
      app.error.has_error = false;
    };

    conn.onmessage = function(evt) {
      var update = JSON.parse(evt.data);
      app.state = update;
    };
  } else {
    app.error.has_error = true;
    app.error.error_message = "Our browser does not support WebSockets.";
  }
}

function getURI() {
  var loc = window.location,
    new_uri;
  if (loc.protocol === "https:") {
    new_uri = "wss:";
  } else {
    new_uri = "ws:";
  }
  new_uri += "//" + loc.host;
  new_uri += loc.pathname + "ws";
  return new_uri;
}
