(function () {
  var form = document.getElementById('contact-form');
  var textarea = document.getElementById('message');
  if (!form || !textarea) return;
  var key = 'draft:' + form.getAttribute('data-car-id');
  try {
    var saved = window.sessionStorage.getItem(key);
    if (saved) {
      textarea.value = saved;
    }
  } catch (error) {
    // Storage can be unavailable in private modes; the server-rendered text
    // is still a safe fallback.
  }
  form.addEventListener('submit', function () {
    try {
      if (form.getAttribute('data-authenticated') === 'true') {
        window.sessionStorage.removeItem(key);
      } else {
        window.sessionStorage.setItem(key, textarea.value);
      }
    } catch (error) {
      // Ignore storage failures and let the form submit normally.
    }
  });
})();
