"use strict";

function _toConsumableArray(arr) { return _arrayWithoutHoles(arr) || _iterableToArray(arr) || _nonIterableSpread(); }

function _nonIterableSpread() { throw new TypeError("Invalid attempt to spread non-iterable instance"); }

function _iterableToArray(iter) { if (Symbol.iterator in Object(iter) || Object.prototype.toString.call(iter) === "[object Arguments]") return Array.from(iter); }

function _arrayWithoutHoles(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = new Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } }

var lsTest = function () {
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

window.onload = function () {
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

          var el = function () {
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
  var reverseList = document.createElement("button");
  reverseList.setAttribute("id", "reverse-list");
  reverseList.innerText = "Reverse";

  reverseList.onclick = function () {
    var items = _toConsumableArray(document.querySelectorAll('ol.items>li'));

    var j = items.length;

    for (var i = 0; i < j; i++) {
      var startItem = items[i];
      var tstartItem = startItem.cloneNode(true);
      var endItem = items[j - 1];
      var tendItem = endItem.cloneNode(true);
      endItem.replaceWith(tstartItem);
      startItem.replaceWith(tendItem);
      j--;
    }
  };

  document.getElementById("controls").appendChild(reverseList);
  var initpItems = JSON.parse(localStorage.getItem(pinnedItems)) || [];
  document.querySelectorAll('ol.items>li').forEach(function (item) {
    var id = item.getAttribute("data-id");
    var pina = document.createElement("button");
    var pinned = initpItems.includes(id);

    if (pinned) {
      pina.innerHTML = "unpin";
    } else {
      pina.innerHTML = "pin";
    }

    pina.addEventListener("click", function el() {
      var lPinnedItems = localStorage.getItem(pinnedItems);
      var lPinnedItemsContent = localStorage.getItem(pinnedItemsContent);
      var pItems = JSON.parse(lPinnedItems) || [];
      var items = JSON.parse(lPinnedItemsContent) || {};
      var pos = pItems.indexOf(id);

      if (pos >= 0) {
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
    });
    item.appendChild(pina);
  });
};
