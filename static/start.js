window.onload = function() {
  var state = 0; // unconnected
  var conn;
  var num = document.getElementById('times')
  var form = document.getElementById('form');

  form.onsubmit = function() { return false; };

  if (window.DeviceMotionEvent && window.WebSocket) {
    conn = new WebSocket('ws://' + document.location.host + document.location.pathname + 'ws');
    conn.onopen = function(event) {
      console.log('websocket opened')
      form.onsubmit = function() {
        var username = document.getElementById('username');
        if (conn && username.value) {
          conn.send(username.value);
        }
        return false;
      };
    };
    conn.onclose = function(event) {
      console.log('websocket closed')
    };
    conn.onmessage = function(event) {
      console.log('receive data: ', event.data)
      switch (event.data) {
        case "start":
          state = 1;
          window.addEventListener('devicemotion', deviceMotionHandler);
          break;
        case "end":
          state = 0;
          window.removeEventListener('devicemotion', deviceMotionHandler);
          break;
      }
    };
    conn.onerror = function(event) {
      console.log('something error')
    };
  } else {
    alert('You browser is holy shit, just delete it and install firefox or chrome!');
  }

  function deviceMotionHandler(event) {
    var speed = 40;
    var x, y, z, lastX, lastY, lastZ;
    x = y = z = lastX = lastY = lastZ = 0;

    var acceleration = event.accelerationIncludingGravity;
    x = acceleration.x;
    y = acceleration.y;
    if(Math.abs(x-lastX) > speed || Math.abs(y-lastY) > speed) {
      conn.send("shake");
      num.innerHTML++;
    }
    lastX = x;
    lastY = y;
  }
}