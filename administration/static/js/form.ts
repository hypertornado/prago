function bindForm() {
  var els = document.querySelectorAll(".form_leavealert");
  for (var i = 0; i < els.length; i++) {
    new Form(<HTMLFormElement>els[i]);
  }
}

//TODO: does not work with image picker and other hidden elements
class Form {
  dirty: boolean = false;
  constructor(el: HTMLFormElement) {
    el.addEventListener("submit", () => {
      this.dirty = false;
    })

    let els = el.querySelectorAll(".form_watcher");
    for (var i = 0; i < els.length; i++) {
      var input = <HTMLInputElement>els[i];
      input.addEventListener("input", () => {
        this.dirty = true;
      });
      input.addEventListener("change", () => {
        this.dirty = true;
      });
    }

    window.addEventListener("beforeunload", (e) => {
      if (this.dirty) {
        var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
        e.returnValue = confirmationMessage;
        return confirmationMessage;
      }
    });
  }
}