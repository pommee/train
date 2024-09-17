import './style.css';
import './app.css';

import { SendRequest } from '../wailsjs/go/main/App';

const btnSendRequest = document.getElementById("btn-send-request");
const routeInput = document.getElementById("route-input")
const jsonOutput = document.getElementById("json-output");
const lineNumbers = document.getElementById("line-numbers");
const status = document.getElementById("status-code");
const responseTime = document.getElementById("response-time");
const responseSize = document.getElementById("response-size");
const requestMethod = document.getElementById("route-options")
const btnHeaders = document.getElementById("btn-headers")
const btnBody = document.getElementById("btn-body");
const btnCookies = document.getElementById("btn-cookies");
const responseWrapper = document.getElementById("response-wrapper");
const headersTable = document.getElementById("headers-table-container");
const headersBody = document.getElementById("headers-body");
const resizer = document.getElementById('resizer');
const response = document.getElementById('response');
const responseFormatButtons = document.querySelectorAll("#response-format button");

let lastResponse = null
let startY, startHeight;

btnSendRequest.addEventListener("click", function (e) {
    console.log("Sending request to", routeInput.value)
    try {
        SendRequest(routeInput.value, requestMethod.value)
            .then((response) => {
                lastResponse = response
                displayResponse()

                let headerKeysLen = Object.keys(lastResponse.Headers).length;
                btnHeaders.innerText = `Headers (${headerKeysLen})`;
            })
            .catch((err) => {
                console.error(err);
            });
    } catch (err) {
        console.error(err);
    }
});

btnBody.addEventListener("click", () => {
    responseWrapper.style.display = "flex";
    headersTable.style.display = "none";
    btnBody.classList.add("btn-active");
    btnHeaders.classList.remove("btn-active");
    btnCookies.classList.remove("btn-active");
    displayResponse()
});

btnHeaders.addEventListener("click", () => {
    responseWrapper.style.display = "none";
    headersTable.style.display = "block";
    btnBody.classList.remove("btn-active");
    btnHeaders.classList.add("btn-active");
    btnCookies.classList.remove("btn-active");
    displayHeaders()
});

document.getElementById('json-output').addEventListener('scroll', function () {
    lineNumbers.scrollTop = jsonOutput.scrollTop;
});

responseFormatButtons.forEach(button => {
    button.addEventListener("click", () => {
        responseFormatButtons.forEach(btn => btn.classList.remove("active-response-format"));
        button.classList.add("active-response-format");
    });
});

function displayResponse() {

    jsonOutput.innerHTML = ""
    let currLastResponse = lastResponse
    const lines = currLastResponse.Response.split('\n');

    lineNumbers.innerHTML = lines.map((_, index) => index + 1).join('<br>');
    jsonOutput.innerHTML = syntaxHighlight(currLastResponse.Response);

    status.innerText = currLastResponse.Status
    responseTime.innerText = currLastResponse.TotalTime + "ms"
    responseSize.innerText = currLastResponse.Size
}

function displayHeaders() {

    headersBody.innerHTML = "";
    Object.entries(lastResponse.Headers).forEach(([key, value]) => {
        const row = document.createElement("tr");
        const headerCell = document.createElement("td");
        const valueCell = document.createElement("td");

        headerCell.textContent = key;
        valueCell.textContent = value;

        row.appendChild(headerCell);
        row.appendChild(valueCell);
        headersBody.appendChild(row);
    });
}

function syntaxHighlight(json) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        let cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}

const startResize = (e) => {
    startY = e.clientY;
    startHeight = parseFloat(getComputedStyle(response).height);

    document.addEventListener('mousemove', resize);
    document.addEventListener('mouseup', stopResize);
};

const resize = (e) => {
    const newHeight = startHeight + (startY - e.clientY);
    response.style.height = `${newHeight}px`;
    response.style.top = 'auto';
};

const stopResize = () => {
    document.removeEventListener('mousemove', resize);
    document.removeEventListener('mouseup', stopResize);
};

resizer.addEventListener('mousedown', startResize);
