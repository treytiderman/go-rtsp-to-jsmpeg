async function fetchJson(method, url, body = "") {
    if (body === "") {
        const response = await fetch(url, {
            method: method,
            headers: { 'Accept': 'application/json' }
        })
        const json = await response.json();
        return json
    } else {
        const response = await fetch(url, {
            method: method,
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body)
        })
        const json = await response.json();
        return json
    }
}

async function getStreams() {
    const method = "GET"
    const url = "/api/streams"
    const res = fetchJson(method, url)
    return res
}

async function newStream(name, ffmpeg) {
    const method = "POST"
    const url = "/api/stream/new"
    const body = { name, ffmpeg }
    const res = fetchJson(method, url, body)
    return res
}

async function newStreamStart() { }

async function startStream(id) {
    const method = "GET"
    const url = "/api/stream/start/" + id
    const res = fetchJson(method, url)
    return res
}

async function stopStream(id) {
    const method = "GET"
    const url = "/api/stream/stop/" + id
    const res = fetchJson(method, url)
    return res
}

async function removeStream(id) {
    const method = "GET"
    const url = "/api/stream/remove/" + id
    const res = fetchJson(method, url)
    return res
}

async function clearStreams() {
    const method = "DELETE"
    const url = "/api/streams"
    const res = fetchJson(method, url)
    return res
}

// async function updateStreamName() { }


