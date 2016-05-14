function bindMarkdowns() {
  function bindMarkdown(el) {
    var textarea = el.find("textarea");
    var lastChanged = Date.now();
    var changed = false;

    setInterval(function(){
      if (changed && (Date.now() - lastChanged > 500)) {
        loadPreview();
      }
    }, 100);

    textarea.on("change keyup", function() {
      changed = true;
      lastChanged = Date.now();
    })

    loadPreview();

    function loadPreview() {
      changed = false;
      console.log("L");
      $.ajax({
          url: '/admin/_api/markdown',
          type: 'POST',
          data: textarea.val(),
          cache: false,
          dataType: 'json',
          processData: false,
          contentType: false,
          success: function(result) {
            el.find(".admin_markdown_preview").html(result);
          },
          error: function() {
              console.log("error while loading markdown preview");
          }
      });
    }
  }

  $(".admin_markdown").each(
    function() {
      bindMarkdown($(this));
    }
  );
}