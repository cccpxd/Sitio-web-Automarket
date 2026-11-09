(function () {
  var links = Array.prototype.slice.call(document.querySelectorAll('[data-photo-index]'));
  var viewer = document.getElementById('photo-viewer');
  if (!viewer || !links.length) return;
  var image = viewer.querySelector('[data-photo-image]');
  var counter = viewer.querySelector('[data-photo-counter]');
  var close = viewer.querySelector('.photo-viewer-close');
  var panel = viewer.querySelector('.photo-viewer-panel');
  var previous = viewer.querySelector('[data-photo-prev]');
  var next = viewer.querySelector('[data-photo-next]');
  var index = 0;
  var opener = null;

  function render() {
    var link = links[index];
    image.src = link.getAttribute('data-photo-url');
    image.alt = link.getAttribute('data-photo-alt') || '';
    counter.textContent = (index + 1) + ' de ' + links.length;
    previous.hidden = links.length < 2;
    next.hidden = links.length < 2;
  }

  function openViewer(event) {
    event.preventDefault();
    opener = event.currentTarget;
    index = Number(opener.getAttribute('data-photo-index')) || 0;
    render();
    viewer.hidden = false;
    document.body.classList.add('photo-viewer-open');
    close.focus();
  }

  function closeViewer() {
    viewer.hidden = true;
    document.body.classList.remove('photo-viewer-open');
    if (opener) opener.focus();
  }

  function move(step) {
    index = (index + step + links.length) % links.length;
    render();
  }

  links.forEach(function (link) { link.addEventListener('click', openViewer); });
  close.addEventListener('click', closeViewer);
  previous.addEventListener('click', function () { move(-1); });
  next.addEventListener('click', function () { move(1); });
  viewer.addEventListener('click', function (event) {
    if (event.target.hasAttribute('data-photo-close')) closeViewer();
  });
  viewer.addEventListener('keydown', function (event) {
    if (event.key === 'Escape') { event.preventDefault(); closeViewer(); return; }
    if (event.key === 'ArrowLeft') { event.preventDefault(); move(-1); return; }
    if (event.key === 'ArrowRight') { event.preventDefault(); move(1); return; }
    if (event.key !== 'Tab') return;
    var focusable = [close, previous, next].filter(function (control) { return !control.hidden; });
    if (focusable.length === 0) { panel.focus(); return; }
    var position = focusable.indexOf(document.activeElement);
    if (event.shiftKey && (position <= 0 || position === -1)) {
      event.preventDefault();
      focusable[focusable.length - 1].focus();
    } else if (!event.shiftKey && position === focusable.length - 1) {
      event.preventDefault();
      focusable[0].focus();
    }
  });
})();
