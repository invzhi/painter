window.onload = function() {
  var num = document.getElementById('times')
  var conn;
  if (window.DeviceMotionEvent && window.WebSocket) {
    conn = new WebSocket('ws://' + document.location.host + document.location.pathname + 'ws');
    conn.onopen = function(event) {
      console.log('websocket open')
      window.addEventListener('devicemotion', deviceMotionHandler);
    }
    conn.onclose = function(event) {
      console.log('websocket close')
    }
    conn.onmessage = function(event) {
      console.log(event.data)
    }
    conn.onerror = function(event) {
      console.log('something error')
    }
  } else {
    alert('You browser is holy shit, just delete it and install firefox or chrome!')
  }

  function deviceMotionHandler(event) {
    var speed = 25;
    var x, y, z, lastX, lastY, lastZ;
    x = y = z = lastX = lastY = lastZ = 0;

    var acceleration = event.accelerationIncludingGravity;
    x = acceleration.x;
    y = acceleration.y;
    if(Math.abs(x-lastX) > speed || Math.abs(y-lastY) > speed) {
      conn.send("shake")
      num.innerHTML++;
    }
    lastX = x;
    lastY = y;
  }
}
