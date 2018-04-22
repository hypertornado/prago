document.addEventListener("DOMContentLoaded", () => {
  //bindOrder();
  bindMarkdowns();
  bindTimestamps();
  bindRelationsView();
  bindRelations();
  bindImagePickers();
  //bindDelete();
  bindLists();
  bindForm();
  bindImageViews();
  bindFlashMessages();
  bindFilter();
});

function bindFlashMessages() {
  var messages = document.querySelectorAll(".flash_message");
  for (var i = 0; i < messages.length; i++) {
    var message = <HTMLDivElement>messages[i];
    message.addEventListener("click", (e) => {
      var target = <HTMLDivElement>e.target;
      if (target.classList.contains("flash_message_close")) {
        var current = <HTMLDivElement>e.currentTarget;
        current.classList.add("hidden");
      }
    })
  }
}