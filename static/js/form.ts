//TODO: does not work with image picker and other hidden elements
class Form {
  dirty: boolean = false;
  constructor(el: HTMLFormElement) {
    var elements = el.querySelectorAll<HTMLDivElement>(".admin_markdown");
    elements.forEach((el) => {
      new MarkdownEditor(el);
    });

    var timestamps = el.querySelectorAll<HTMLDivElement>(".admin_timestamp");
    timestamps.forEach((el) => {
      new Timestamp(el);
    });

    var relations = el.querySelectorAll<HTMLDivElement>(".admin_item_relation");
    relations.forEach((el) => {
      new RelationPicker(el);
    });

    var imagePickers = el.querySelectorAll<HTMLDivElement>(".admin_images");
    imagePickers.forEach((el) => {
      new ImagePicker(el);
    });

    var dateInputs = el.querySelectorAll<HTMLInputElement>(".form_input-date");
    dateInputs.forEach((el) => {
      new DatePicker(el);
    });

    var elements = el.querySelectorAll<HTMLDivElement>(".admin_place");
    elements.forEach((el) => {
      new PlacesEdit(el);
    });

    el.addEventListener("submit", () => {
      this.dirty = false;
    });

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
