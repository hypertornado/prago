//TODO: does not work with image picker and other hidden elements
class Form {
  dirty: boolean = false;
  ajax: boolean = false;
  formEl: HTMLFormElement;

  constructor(form: HTMLFormElement) {
    this.formEl = form;

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

    var imagePickers = form.querySelectorAll<HTMLDivElement>(".admin_images");
    imagePickers.forEach((form) => {
      new ImagePicker(form);
    });

    var dateInputs =
      form.querySelectorAll<HTMLInputElement>(".form_input-date");
    dateInputs.forEach((form) => {
      new DatePicker(form);
    });

    var elements = form.querySelectorAll<HTMLDivElement>(".admin_place");
    elements.forEach((form) => {
      new PlacesEdit(form);
    });

    form.addEventListener("submit", () => {
      this.dirty = false;
    });

    let els = form.querySelectorAll(".form_watcher");
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
