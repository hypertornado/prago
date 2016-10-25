function bindMarkdowns() {
  function bindMarkdown(el: HTMLElement) {
    var textarea = <HTMLTextAreaElement>el.getElementsByTagName("textarea")[0];
    var lastChanged = Date.now();
    var changed = false;

    setInterval(function(){
      if (changed && (Date.now() - lastChanged > 500)) {
        loadPreview();
      }
    }, 100);

    textarea.addEventListener("change", textareaChanged);
    textarea.addEventListener("keyup", textareaChanged);
    function textareaChanged() {
      changed = true;
      lastChanged = Date.now(); 
    }
    loadPreview();

    function loadPreview() {
      changed = false;
      var request = new XMLHttpRequest();
      request.open("POST", document.body.getAttribute("data-admin-prefix") + "/_api/markdown", true);

      request.onload = function() {
        if (this.status == 200) {
          console.log(JSON.parse(this.response));
          var previewEl = el.getElementsByClassName("admin_markdown_preview")[0];
          previewEl.innerHTML = JSON.parse(this.response);
        } else {
          console.error("Error while loading markdown preview.");
        }
      }
      request.send(textarea.value);
    }
  }

  var elements = document.querySelectorAll(".admin_markdown");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    bindMarkdown(el);
  });
}