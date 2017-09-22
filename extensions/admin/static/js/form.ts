function bindForm() {
  var els = document.querySelectorAll(".form_leavealert");
  for (var i = 0; i < els.length; i++) {
    new Form(<HTMLFormElement>els[i]);
  }
}

class Form {
  allow: boolean = false;

  constructor(el: HTMLFormElement) {
    el.addEventListener("submit", () => {
      this.allow = true
    })

    window.addEventListener("beforeunload", (e) => {
      if (this.allow) {
        return;
      }
      var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
      e.returnValue = confirmationMessage;
      return confirmationMessage;
    });
  }
}