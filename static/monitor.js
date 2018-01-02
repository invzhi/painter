window.onload = function() {
  var list = document.getElementById('list');
  var tips = document.getElementById('tips');

  if (window.WebSocket) {
    var conn = new WebSocket('ws://' + document.location.host + document.location.pathname + 'ws');

    conn.onopen = function(event) {
      console.log('websocket opened.');
      document.getElementById('start').onclick = function() { conn.send('start'); };
      document.getElementById('end').onclick = function() {
        conn.send('end');
        conn.onmessage = function(event) {
          console.log(event.data);
        }
      };
    };

    conn.onclose = function(event) {
      console.log('websocket closed');
    };

    conn.onmessage = function(event) {
      console.log(event.data);

      var msg = JSON.parse(event.data);
      switch (msg.code) {
        case 0:
          console.log(msg.username, 'join');
          appendUser(msg.username)
          break;
        case 1:
          console.log(msg.username, 'shake');
          document.getElementById(msg.username).innerHTML++;
          break;
      }
    };

    conn.onerror = function(event) {
      console.log('something error');
    };
  } else {
    tips.innerHTML = 'You browser is holy shit, just delete it and install firefox or chrome!';
  }
}

function appendUser(username) {
  var item = document.createElement("li");
  var label = document.createElement('label');
  var number = document.createElement('span');
  label.innerHTML = username;
  number.id = username;
  number.innerHTML = 0;
  item.appendChild(label);
  item.appendChild(number);
  list.appendChild(item);
}
