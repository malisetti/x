"use strict";

var lsTest = function lsTest() {
  var test = 'test';

  try {
    localStorage.setItem(test, test);
    localStorage.removeItem(test);
    return true;
  } catch (e) {
    return false;
  }
};

var pinnedItems = "PINNED_ITEMS";
var pinnedItemsContent = "PINNED_ITEMS_CONTENT";
var noPinsMsg = "Nothing is pinned!";

window.onload = function (e) {
  if (!lsTest()) {
    return;
  }

  var showPins = document.createElement("button");
  showPins.setAttribute("id", "show-pins");
  showPins.innerText = "Show Pins";

  showPins.onclick = function () {
    var lPinnedItems = localStorage.getItem(pinnedItems);
    var pItems = JSON.parse(lPinnedItems) || [];

    if (!(pItems.length > 0)) {
      alert(noPinsMsg);
      return;
    }

    var lPinnedItemsContent = localStorage.getItem(pinnedItemsContent);
    var items = JSON.parse(lPinnedItemsContent) || {};

    if (!(Object.keys(items).length > 0)) {
      alert(noPinsMsg);
      return;
    }

    var container = document.querySelector("ol.items");
    container.innerHTML = "";

    for (var key in items) {
      if (items.hasOwnProperty(key)) {
        (function () {
          var div = document.createElement("div");
          div.innerHTML = items[key].trim();
          var fc = div.firstChild;
          var unpin = fc.lastChild;
          unpin.innerHTML = "unpin";

          var el = function el(ev) {
            // remove it from ls and change the pina to pin
            var id = fc.getAttribute("data-id");
            delete items[id];
            pItems = pItems.filter(function (pi) {
              return pi !== id;
            });
            localStorage.setItem(pinnedItems, JSON.stringify(pItems));
            localStorage.setItem(pinnedItemsContent, JSON.stringify(items));
            fc.remove();
          };

          unpin.addEventListener("click", el);
          container.appendChild(div.firstChild);
        })();
      }
    }
  };

  document.getElementById("controls").appendChild(showPins);
  var lPinnedItems = localStorage.getItem(pinnedItems);
  var lPinnedItemsContent = localStorage.getItem(pinnedItemsContent);
  var pItems = JSON.parse(lPinnedItems) || [];
  var items = JSON.parse(lPinnedItemsContent) || {};
  document.querySelectorAll('ol.items>li').forEach(function (item) {
    var id = item.getAttribute("data-id");
    var pina = document.createElement("button");
    var pinned = pItems.includes(id);

    if (pinned) {
      pina.innerHTML = "unpin";
    } else {
      pina.innerHTML = "pin";
    }

    var el = function el(ev) {
      var pos = pItems.indexOf(id);
      var pinned = pos >= 0;

      if (pinned) {
        // remove it from ls and change the pina to pin
        delete items[id];
        pItems = pItems.filter(function (pi) {
          return pi !== id;
        });
        pina.innerHTML = "pin";
      } else {
        // add to ls and change the pina to unpin
        // remove it from ls and change the pina to pin
        items[id] = item.outerHTML;

        if (pos === -1) {
          pItems.push(id);
        }

        pina.innerHTML = "unpin";
      }

      localStorage.setItem(pinnedItems, JSON.stringify(pItems));
      localStorage.setItem(pinnedItemsContent, JSON.stringify(items));
    };

    pina.addEventListener("click", el);
    item.appendChild(pina);
  });
};
