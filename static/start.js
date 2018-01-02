window.onload = function() {
  var conn, number;
  var form = document.getElementById('form');
  var tips = document.getElementById('tips');

  form.onsubmit = function() { return false; };

  if (window.DeviceMotionEvent && window.WebSocket) {
    conn = new WebSocket('ws://' + document.location.host + document.location.pathname + 'ws');

    conn.onopen = function(event) {
      console.log('websocket opened.');

      form.onsubmit = function() {
        var username = document.getElementById('username');
        if (username.value) {
          conn.send(username.value);
          tips.innerHTML = 'You have a name: ' + username.value;
          number = document.createElement('h2');
          number.innerHTML = 0;
          document.getElementsByTagName('body')[0].insertBefore(number, form);
          form.remove();
          form = number;
        } else {
          tips.innerHTML = 'cannot be a name';
        }
        return false;
      };
    };

    conn.onmessage = function(event) {
      console.log('receive data: ', event.data);
      switch (event.data) {
        case "start":
          window.addEventListener('devicemotion', deviceMotionHandler);
          tips.innerHTML = 'Game Start';
          break;
        case "end":
          window.removeEventListener('devicemotion', deviceMotionHandler);
          tips.innerHTML = 'Game End';
          break;
      }
    };

    conn.onclose = function(event) {
      console.log('websocket closed.');
      tips.innerHTML = 'you are offline.';
    };

    conn.onerror = function(event) {
      console.log('websocket error.');
      tips.innerHTML = 'you have trouble.'
    };
  } else {
    tips.innerHTML = 'You browser is holy shit, just delete it and install better one!';
  }

  var SHAKE_THRESHOLD = 800;
  var last_update = 0;
  var x = y = z = last_x = last_y = last_z = 0;

  function deviceMotionHandler(eventData) {
    var acceleration = eventData.accelerationIncludingGravity;
    var curTime = new Date().getTime();
    
    if (curTime - last_update > 100) {
      var diffTime = curTime - last_update;
      last_update = curTime;
      x = acceleration.x;
      y = acceleration.y;
      z = acceleration.z;
      var speed = Math.abs(x + y + z - last_x - last_y - last_z) / diffTime * 10000;
      
      if (speed > SHAKE_THRESHOLD) {
        conn.send("shake");
        number.innerHTML++;
      }
      last_x = x;
      last_y = y;
      last_z = z;
    }
  }
}
