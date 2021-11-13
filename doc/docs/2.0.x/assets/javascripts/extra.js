(function () {
'use strict';
/**
 * Converts details/summary tags into working elements in browsers that don't yet support them.
 * @return {void}
 */
var details = (function () {

  var isDetailsSupported = function () {
    // https://mathiasbynens.be/notes/html5-details-jquery#comment-35
    // Detect if details is supported in the browser
    var el = document.createElement("details");
    var fake = false;

    if (!("open" in el)) {
      return false;
    }

    var root = document.body || function () {
      var de = document.documentElement;
      fake = true;
      return de.insertBefore(document.createElement("body"), de.firstElementChild || de.firstChild);
    }();

    el.innerHTML = "<summary>a</summary>b";
    el.style.display = "block";
    root.appendChild(el);
    var diff = el.offsetHeight;
    el.open = true;
    diff = diff !== el.offsetHeight;
    root.removeChild(el);

    if (fake) {
      root.parentNode.removeChild(root);
    }

    return diff;
  }();

  if (!isDetailsSupported) {
    var blocks = document.querySelectorAll("details>summary");
    for (var i = 0; i < blocks.length; i++) {
      var summary = blocks[i];
      var details = summary.parentNode;

      // Apply "no-details" to for unsupported details tags
      if (!details.className.match(new RegExp("(\\s|^)no-details(\\s|$)"))) {
        details.className += " no-details";
      }

      summary.addEventListener("click", function (e) {
        var node = e.target.parentNode;
        if (node.hasAttribute("open")) {
          node.removeAttribute("open");
        } else {
          node.setAttribute("open", "open");
        }
      });
    }
  }
});

(function () {
  var onReady = function onReady(fn) {
    if (document.addEventListener) {
      document.addEventListener("DOMContentLoaded", fn);
    } else {
      document.attachEvent("onreadystatechange", function () {
        if (document.readyState === "interactive") {
          fn();
        }
      });
    }
  };

  onReady(function () {
    details();
  });
})();

}());
