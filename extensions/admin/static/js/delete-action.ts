function bindDelete() {
  var deleteButtons = document.querySelectorAll(".admin-action-delete")
  for (var i = 0; i < deleteButtons.length; i++) {
    bindDeleteButton(<HTMLDivElement>deleteButtons[i]);
  }
}

function bindDeleteButton(btn: HTMLDivElement) {
  btn.addEventListener("click", () => {
    var message = btn.getAttribute("data-confirm-message");
    var url = btn.getAttribute("data-action");

    if (confirm(message)) {
      var request = new XMLHttpRequest();
      request.open("POST", url, true);

      request.onload = function() {
        if (this.status == 200) {
          document.location.reload();
        } else {
          console.error("Error while deleting item");
        }
      }
      request.send();
    }
  });
}