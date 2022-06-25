var nid = "79796f3e-d63d-4608-9ce8-f375eb91129f";

async function noninteractionEvent() {
    const payload = {
        eventType: "page_load",
        clickTarget: null,
        closestLinkHref: null,
        userAgent: window.navigator.userAgent,
        hardwareConcurrency: window.navigator.hardwareConcurrency,
        deviceMemory: window.navigator.deviceMemory,
        maxTouchPoints: window.navigator.maxTouchPoints,
        language: window.navigator.language,
        href: window.location.href,
        referrer: document.referrer,
        nid: nid,
        cookieEnabled: window.navigator.cookieEnabled,
        onLine: window.navigator.onLine,
    }
    const response = await fetch("http://localhost:3106/monitor", {
        method: "POST",
        body: JSON.stringify(payload),
        headers: {
            "Content-Type": "application/json",
            "Access-Control-Allow-Origin": "*",
        }
    })
    if (!response.ok) {
        const message = `An error has occured: ${response.status}`;
        throw new Error(message);
    }
}

async function interceptClickEvent(e) {
    var target = e.target || e.srcElement;
    var closestLink = target.closest('a');
    var closestLinkHref;
    if (closestLink) {
        closestLinkHref = closestLink.getAttribute('href');
    }

    const payload = {
        eventType: "page_click",
        clickTarget: target.tagName,
        closestLinkHref: closestLinkHref,
        userAgent: window.navigator.userAgent,
        hardwareConcurrency: window.navigator.hardwareConcurrency,
        deviceMemory: window.navigator.deviceMemory,
        maxTouchPoints: window.navigator.maxTouchPoints,
        language: window.navigator.language,
        href: window.location.href,
        referrer: document.referrer,
        nid: nid,
        cookieEnabled: window.navigator.cookieEnabled,
        onLine: window.navigator.onLine,
    }
    const response = await fetch("http://localhost:3106/monitor", {
        method: "POST",
        body: JSON.stringify(payload),
        headers: {
            "Content-Type": "application/json",
            "Access-Control-Allow-Origin": "*",
        }
    })
    if (!response.ok) {
        const message = `An error has occured: ${response.status}`;
        throw new Error(message);
    }
}

if (document.addEventListener) {
    document.addEventListener('mousedown', interceptClickEvent);
} else if (document.attachEvent) {
    document.attachEvent('onmousedown', interceptClickEvent);
}

noninteractionEvent().catch(error => { console.log(error) });