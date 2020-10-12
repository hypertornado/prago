document.addEventListener("DOMContentLoaded", () => {
  bindStats();
  bindMarkdowns();
  bindTimestamps();
  bindRelations();
  bindImagePickers();
  bindLists();
  bindForm();
  bindImageViews();
  bindFlashMessages();
  bindScrolled();
  bindDatePicker();
  bindDropdowns();
  bindSearch();
  bindEshopControl();
  bindMainMenu();
  bindRelationList();
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

function bindScrolled() {
  var lastScrollPosition = 0;
  var header = <HTMLDivElement>document.querySelector(".admin_header");
  document.addEventListener("scroll", (event) => {
    if (document.body.clientWidth < 1100) {
      return;
    }
    var scrollPosition = window.scrollY;
    if (scrollPosition > 0) {
      header.classList.add("admin_header-scrolled");
    } else {
      header.classList.remove("admin_header-scrolled");
    }
    lastScrollPosition = scrollPosition;
  });
}