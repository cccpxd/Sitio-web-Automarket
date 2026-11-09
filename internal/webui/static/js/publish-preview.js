(function () {
  var input = document.getElementById('photos');
  var preview = document.getElementById('photo-preview');
  if (!input || !preview) return;
  input.addEventListener('change', function () {
    preview.innerHTML = '';
    var files = Array.prototype.slice.call(input.files);
    files.forEach(function (file, index) {
      var item = document.createElement('div');
      item.className = 'preview-item';
      var reader = new FileReader();
      reader.onload = function (event) {
        var img = document.createElement('img');
        img.src = event.target.result;
        img.alt = file.name;
        var remove = document.createElement('button');
        remove.type = 'button';
        remove.className = 'btn btn-small btn-danger preview-remove';
        remove.textContent = 'Quitar';
        remove.addEventListener('click', function () {
          files.splice(index, 1);
          var transfer = new DataTransfer();
          files.forEach(function (selected) { transfer.items.add(selected); });
          input.files = transfer.files;
          input.dispatchEvent(new Event('change'));
        });
        item.appendChild(img);
        item.appendChild(remove);
        preview.appendChild(item);
      };
      reader.readAsDataURL(file);
    });
  });
})();
