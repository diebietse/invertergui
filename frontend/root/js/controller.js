function loadContent() {
  var conn;

  var app = new Vue({
    el: "#app",
    data: {
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
      led_mains: "dot-off",
      led_absorb: "dot-off",
      led_bulk: "dot-off",
      led_float: "dot-off",
      led_inverter: "dot-off",
      led_overload: "dot-off",
      led_bat_low: "dot-off",
      led_over_temp: "dot-off"
    }
  });

  if (window["WebSocket"]) {
    conn = new WebSocket("ws://" + document.location.host + "/ws");
    conn.onclose = function(evt) {
      var item = document.createElement("a");
      item.innerHTML = "<b>Connection closed.</b>";
    };

    conn.onmessage = function(evt) {
      var update = JSON.parse(evt.data);
      console.log(update);
      app.output_current = update.output_current;
      app.output_voltage = update.output_voltage;
      app.output_frequency = update.output_frequency;
      app.output_power = update.output_power;

      app.input_current = update.input_current;
      app.input_voltage = update.input_voltage;
      app.input_frequency = update.input_frequency;
      app.input_power = update.input_power;

      app.battery_charge = update.battery_charge;
      app.battery_voltage = update.battery_voltage;
      app.battery_power = update.battery_power;
      app.battery_current = update.battery_current;

      leds = update.led_map;
      app.led_mains = leds.led_mains;
      app.led_absorb = leds.led_absorb;
      app.led_bulk = leds.led_bulk;
      app.led_float = leds.led_float;
      app.led_inverter = leds.led_inverter;
      app.led_overload = leds.led_overload;
      app.led_bat_low = leds.led_bat_low;
      app.led_over_temp = leds.led_over_temp;
    };
  } else {
    var item = document.createElement("a");
    item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
  }
}
