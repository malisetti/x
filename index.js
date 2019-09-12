"use strict";

const lsTest = () => {
    const test = 'test'
    try {
        localStorage.setItem(test, test)
        localStorage.removeItem(test)
        return true
    } catch (e) {
        return false
    }
}
const pinnedItems = "PINNED_ITEMS"
const pinnedItemsContent = "PINNED_ITEMS_CONTENT"
const noPinsMsg = "Nothing is pinned!"

window.onload = (e) => {
    if (!lsTest()) {
        return
    }
    const showPins = document.createElement("button");
    showPins.setAttribute("id", "show-pins")
    showPins.innerText = "Show Pins"
    showPins.onclick = () => {
        const lPinnedItems = localStorage.getItem(pinnedItems)
        let pItems = JSON.parse(lPinnedItems) || []
        if (!(pItems.length > 0)) {
            alert(noPinsMsg)
            return
        }
        const lPinnedItemsContent = localStorage.getItem(pinnedItemsContent)
        const items = JSON.parse(lPinnedItemsContent) || {}
        if (!(Object.keys(items).length > 0)) {
            alert(noPinsMsg)
            return
        }
        const container = document.querySelector("ol.items")
        container.innerHTML = ""
        for (let key in items) {
            if (items.hasOwnProperty(key)) {
                const div = document.createElement("div")
                div.innerHTML = items[key].trim()
                const fc = div.firstChild
                const unpin = fc.lastChild
                unpin.innerHTML = "unpin"
                const el = (ev) => {
                    // remove it from ls and change the pina to pin
                    const id = fc.getAttribute("data-id")
                    delete items[id]
                    pItems = pItems.filter(pi => pi !== id)

                    localStorage.setItem(pinnedItems, JSON.stringify(pItems))
                    localStorage.setItem(pinnedItemsContent, JSON.stringify(items))
                    fc.remove()
                }
                unpin.addEventListener("click", el)

                container.appendChild(div.firstChild)
            }
        }
    }

    document.getElementById("controls").appendChild(showPins)

    const initpItems = JSON.parse(localStorage.getItem(pinnedItems)) || []

    document.querySelectorAll('ol.items>li').forEach((item) => {
        const id = item.getAttribute("data-id")
        const pina = document.createElement("button")
        const pinned = initpItems.includes(id)
        if (pinned) {
            pina.innerHTML = "unpin"
        } else {
            pina.innerHTML = "pin"
        }
        const el = (ev) => {
            const lPinnedItems = localStorage.getItem(pinnedItems)
            const lPinnedItemsContent = localStorage.getItem(pinnedItemsContent)
            let pItems = JSON.parse(lPinnedItems) || []
            let items = JSON.parse(lPinnedItemsContent) || {}

            const pos = pItems.indexOf(id)
            const pinned = pos >= 0
            if (pinned) {
                // remove it from ls and change the pina to pin
                delete items[id]
                pItems = pItems.filter(pi => pi !== id)
                pina.innerHTML = "pin"
            } else {
                // add to ls and change the pina to unpin
                // remove it from ls and change the pina to pin
                items[id] = item.outerHTML
                if (pos === -1) {
                    pItems.push(id)
                }
                pina.innerHTML = "unpin"
            }

            localStorage.setItem(pinnedItems, JSON.stringify(pItems))
            localStorage.setItem(pinnedItemsContent, JSON.stringify(items))
        }
        pina.addEventListener("click", el)
        item.appendChild(pina)
    })
}
