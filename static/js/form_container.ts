class FormContainer {
  formContainer: HTMLDivElement;
  form: Form;
  progress: HTMLProgressElement;

  constructor(formContainer: HTMLDivElement) {
    this.formContainer = formContainer;
    this.progress = formContainer.querySelector(".form_progress");
    var formEl: HTMLFormElement = formContainer.querySelector("form");
    this.form = new Form(formEl);
    this.form.formEl.addEventListener("submit", this.submitFormAJAX.bind(this));
  }

  submitFormAJAX(event: Event) {
    event.preventDefault();

    let formData = new FormData(this.form.formEl);

    var request = new XMLHttpRequest();
    request.open("POST", this.form.formEl.getAttribute("action"));

    request.addEventListener("load", (e) => {
      if (request.status == 200) {
        var data = JSON.parse(request.response);
        if (data.RedirectionLocaliton) {
          window.location = data.RedirectionLocaliton;
        } else {
          this.progress.classList.add("hidden");
          this.setFormErrors(data.Errors);
          this.setItemErrors(data.ItemErrors);
          if (data.AfterContent) this.setAfterContent(data.AfterContent);
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

  setAfterContent(text: string) {
    this.formContainer.querySelector(".form_after_content").innerHTML = text;
  }

  setFormErrors(errors: any[]) {
    let errorsDiv: HTMLDivElement =
      this.form.formEl.querySelector(".form_errors");
    errorsDiv.innerText = "";
    errorsDiv.classList.add("hidden");

    if (errors) {
      for (let i = 0; i < errors.length; i++) {
        let errorDiv = document.createElement("div");
        errorDiv.classList.add("form_errors_error");
        errorDiv.innerText = errors[i].Text;
        errorsDiv.appendChild(errorDiv);
      }
      if (errors.length > 0) {
        errorsDiv.classList.remove("hidden");
      }
    }
  }

  setItemErrors(itemErrors: any) {
    let labels = this.form.formEl.querySelectorAll(".form_label");
    for (let i = 0; i < labels.length; i++) {
      let label = labels[i];
      let id = label.getAttribute("data-id");
      label.classList.remove("form_label-errors");
      let labelErrors = label.querySelector(".form_label_errors");
      labelErrors.innerHTML = "";
      labelErrors.classList.add("hidden");
      if (itemErrors[id]) {
        label.classList.add("form_label-errors");
        labelErrors.classList.remove("hidden");
        for (let j = 0; j < itemErrors[id].length; j++) {
          let errorDiv = document.createElement("div");
          errorDiv.classList.add("form_label_errors_error");
          errorDiv.innerText = itemErrors[id][j].Text;
          labelErrors.appendChild(errorDiv);
        }
      }
    }
  }
}
