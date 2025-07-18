//TODO: does not work with image picker and other hidden elements
class Form {
  private dirty: boolean = false;
  formEl: HTMLFormElement;
  lastChanged: number;
  changeHandler: any;
  willChangeHandler: any;

  constructor(form: HTMLFormElement) {
    this.dirty = false;
    this.formEl = form;

    this.fixAufofocus();

    var elements = form.querySelectorAll<HTMLDivElement>(".admin_markdown");
    elements.forEach((el) => {
      new MarkdownEditor(el);
    });

    var timestamps = form.querySelectorAll<HTMLDivElement>(".admin_timestamp");
    timestamps.forEach((form) => {
      new Timestamp(form);
    });

    var relations = form.querySelectorAll<HTMLDivElement>(
      ".admin_item_relation"
    );
    relations.forEach((form) => {
      new RelationPicker(form);
    });

    var imagePickers = form.querySelectorAll<HTMLDivElement>(".imagepicker");
    imagePickers.forEach((form) => {
      new ImagePicker(form);
    });

    var textovers = form.querySelectorAll<HTMLDivElement>(".form_label_textover");
    textovers.forEach((textover) => {
      textover.addEventListener("click", () => {
        textover.parentElement.classList.add("form_label-textoverexpanded");
        let inputs = textover.parentElement.querySelectorAll(".input");
        if (inputs) {
          let input = <HTMLDivElement>inputs[0];
          input.focus();
        }
      })
    });

    form.addEventListener("submit", () => {
      this.dirty = false;
    });

    let els = form.querySelectorAll(".form_watcher");
    for (var i = 0; i < els.length; i++) {
      var input = <HTMLInputElement>els[i];
      input.addEventListener("keyup", this.messageChanged.bind(this));
      input.addEventListener("change", this.changed.bind(this));
    }

    window.setInterval(() => {
      if (this.dirty && Date.now() - this.lastChanged > 500) {
        this.changed();
      }
    }, 100);

    //TODO enable this when it works with new change watcher
    /*
    window.addEventListener("beforeunload", (e) => {
      if (this.dirty) {
        var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
        e.returnValue = confirmationMessage;
        return confirmationMessage;
      }
    });*/
  }

  messageChanged() {
    if (this.willChangeHandler) {
      this.willChangeHandler();
    }
    this.dirty = true;
    this.lastChanged = Date.now();
  }

  changed() {
    if (this.changeHandler) {
      this.dirty = false;
      this.changeHandler();
    } else {
      this.dirty = true;
    }
  }

  fixAufofocus() {
    let input: HTMLInputElement = this.formEl.querySelector('[autofocus]');
    if (input) {
      let value = input.value;
      let typ = input.getAttribute("type");
      if (input.nodeName == "TEXTAREA" || typ == "text" || typ == "password" || typ == "tel" || typ == "search" || typ == "url") {
        input.focus();
        input.setSelectionRange(value.length, value.length);
      }
    }
  }
}
