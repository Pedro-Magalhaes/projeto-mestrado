const carDiv = `<div class="card blue-grey darken-1 log-view">
    <a onclick="remove({{id}})" class="waves-effect waves-light right btn-small red onClick"><i class="material-icons left">remove</i></a>
    <div class="card-content white-text">
      <span class="card-title">{{title}}</span>
      <div class="content"></div>
    </div>
  </div>`
const logsWS = {};
const baseUrl = "localhost:7777"
const wsUrl = 'ws://' + baseUrl + "/ws"
let sequence = 0
function createWs(path, jobid, project, logViewElement) {
    let url = wsUrl + "?jobid=j1&projectid=p1&path=" + path;
    console.log("Trying url: " + url);
    url = encodeURI(url);
    const ws = new WebSocket(url)
    
    const textDiv = logViewElement.getElementsByClassName("content")[0];
    ws.onopen = (ev) => {
        console.log("Open", ev);
    }
    
    ws.onclose = (ev) => {
        console.log(ev);
        if (ev.wasClean) {
            alert(`[close] Connection closed cleanly, code=${ev.code} reason=${ev.reason}`);
          } else {
            // e.g. server process killed or network down
            // event.code is usually 1006 in this case
            alert('[close] Connection died');
          }
          document.getElementById("log-container").removeChild(logViewElement);
    }
    ws.onmessage = (ev) => {
        console.log(ev.data);
        const data = JSON.parse(ev.data)
        textDiv.innerText = textDiv.innerText + data.msg;
    }
    return ws;
}

function remove(id) {
    const logElemenmt = document.getElementById(id);
    console.log(logsWS[id]);
    logsWS[id].close();
    document.getElementById("log-container").removeChild(logElemenmt);
}

function appendLogView(fileName) {
    const newDiv = document.createElement("div");
    newDiv.className = "col s12";
    newDiv.id = "log-" + sequence;
    sequence++;
    let inner = carDiv.replace("{{title}}",fileName);
    inner = inner.replace("{{id}}","'" + newDiv.id + "'");
    newDiv.innerHTML = inner
    document.getElementById("log-container").appendChild(newDiv);
    return newDiv
}

function Submit(e) {
    const projectid = e.projectid.value;
    const jobid = e.jobid.value;
    const fileName = e.file_name.value;
    if (fileName == "" || jobid == "" || projectid == "") {
        alert("Invalid form, something is missing!");
        return;
    }
    logView = appendLogView(fileName)
    const ws = createWs(fileName,jobid,projectid,logView);
    logsWS[logView.id] = ws;
}
