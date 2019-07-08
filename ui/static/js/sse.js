window.addEventListener('load', onLoad);

function onLoad () { 
  connect();
}

let es;

function connect() {
  console.log("connecting");
  es = new EventSource('/sse');
  es.onmessage = (e) => {
    if (e.data.includes('refresh')) {
      es.close();
      document.location.reload(true);
    } else {
      console.log(e.data);
    }
  }

  es.onerror = () => {
    console.log("error");
    connect();
  }
}

/*
function connect() {
  //gotActivity();
  let es = new EventSource('/sse');
  es.onmessage = (e) => {
    if (e.data.includes('refresh')) {
      es.close();
      document.location.reload(true);
    }
  }

  es.onerror = () => {
    console.log("error");
    es = new EventSource('/sse');
  }
}
*/