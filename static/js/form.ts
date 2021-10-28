//TODO: does not work with image picker and other hidden elements
class Form {
  dirty: boolean = false;
  ajax: boolean = false;
  form: HTMLFormElement;
  progress: HTMLProgressElement;

  constructor(form: HTMLFormElement) {
    this.form = form;
    this.progress = this.form.querySelector(".form_progress");
    if (form.classList.contains("form-ajax")) {
      console.log("ajax");
      form.addEventListener("submit", this.submitFormAJAX.bind(this));
    }

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

  submitFormAJAX(event: Event) {
    event.preventDefault();

    let formData = new FormData(this.form);

    var request = new XMLHttpRequest();
    request.open("POST", this.form.getAttribute("action"));

    request.addEventListener("load", (e) => {
      if (request.status == 200) {
        var data = JSON.parse(request.response);
        if (data.RedirectionLocaliton) {
          window.location = data.RedirectionLocaliton;
        } else {
          this.progress.classList.add("hidden");
          this.setFormErrors(data.Errors);
        }
      } else {
        this.progress.classList.add("hidden");
        new Alert("Chyba při nahrávání souboru.");
        console.error("Error while loading item.");
      }
    });

    this.progress.classList.remove("hidden");
    request.send(formData);
  }

  setFormErrors(errors: any[]) {
    let errorsDiv: HTMLDivElement = this.form.querySelector(".form_errors");
    errorsDiv.innerText = "";
    for (let i = 0; i < errors.length; i++) {
      let errorDiv = document.createElement("div");
      errorDiv.classList.add("form_errors_error");
      errorDiv.innerText = errors[i].Text;
      errorsDiv.appendChild(errorDiv);
    }
    if (errors.length > 0) {
      errorsDiv.classList.remove("hidden");
    } else {
      errorsDiv.classList.add("hidden");
    }
    console.log(errors);
  }
}
