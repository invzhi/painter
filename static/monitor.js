window.onload = function() {
  var list = document.getElementById('list');

  if (window.WebSocket) {
    var conn = new WebSocket('ws://' + document.location.host + document.location.pathname + 'ws');

    conn.onopen = function(event) {
      console.log('websocket opened')
      document.getElementById('start').onclick = function() {
        conn.send('start');
      };
      document.getElementById('end').onclick = function() {
        conn.send('end');
      };
    }
    conn.onclose = function(event) {
      console.log('websocket closed');
    }
    conn.onmessage = function(event) {
      console.log(event.data);
      var msg = JSON.parse(event.data);
      switch (msg.code) {
        case 0: // join
          item = document.createElement("li");
          item.innerHTML = msg.username + ': <span id="' + msg.username + '">0</span>';
          list.appendChild(item);
          break;
        case 1: // shake
          document.getElementById(msg.username).innerHTML++;
          break;
      }
    }
    conn.onerror = function(event) {
      console.log('something error');
    }
  } else {
    alert('You browser is holy shit, just delete it and install firefox or chrome!')
  }
}
