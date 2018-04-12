document.addEventListener("DOMContentLoaded", () => {
  //bindOrder();
  bindMarkdowns();
  bindTimestamps();
  bindRelationsView();
  bindRelations();
  bindImagePickers();
  //bindDelete();
  bindClickAndStay();
  bindLists();
  bindForm();
  bindImageViews();
  bindFlashMessages();
  bindFilter();
});

function bindClickAndStay() {
  var els = document.getElementsByName("_submit_and_stay");
  var elsClicked = document.getElementsByName("_submit_and_stay_clicked");

  if (els.length == 1 && elsClicked.length == 1) {
    els[0].addEventListener("click", () => {
      (<HTMLInputElement>elsClicked[0]).value = "true";
    })
  }
}

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