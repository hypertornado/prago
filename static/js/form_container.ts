class FormContainer {
  formContainer: HTMLDivElement;
  form: Form;
  progress: HTMLProgressElement;
  lastAJAXID: string;
  activeRequest: XMLHttpRequest;

  constructor(formContainer: HTMLDivElement) {
    this.formContainer = formContainer;
    this.progress = formContainer.querySelector(".form_progress");
    var formEl: HTMLFormElement = formContainer.querySelector("form");
    this.form = new Form(formEl);
    this.form.formEl.addEventListener("submit", this.submitFormAJAX.bind(this));

    if (this.isAutosubmitFirstTime()) {
      this.sendForm();
    }

    if (this.isAutosubmit()) {
      this.form.changeHandler = this.formChanged.bind(this);
      this.form.willChangeHandler = this.formWillChange.bind(this);
      this.sendForm();
    }
  }

  isAutosubmitFirstTime(): boolean {
    if (
      this.formContainer.classList.contains(
        "form_container-autosubmitfirsttime"
      )
    ) {
      return true;
    } else {
      return false;
    }
  }

  isAutosubmit(): boolean {
    if (this.formContainer.classList.contains("form_container-autosubmit")) {
      return true;
    } else {
      return false;
    }
  }

  formWillChange() {
    this.progress.classList.remove("hidden");
  }

  formChanged() {
    this.sendForm();
  }

  submitFormAJAX(event: Event) {
    event.preventDefault();
    this.sendForm();
  }

  sendForm() {
    let formData = new FormData(this.form.formEl);
    let request = new XMLHttpRequest();
    request.open("POST", this.form.formEl.getAttribute("action"));

    let requestID: string = this.makeid(10);
    this.lastAJAXID = requestID;

    if (this.activeRequest) {
      if (this.isAutosubmit()) {
        this.activeRequest.abort();
      } else {
        return;
      }
    }
    this.activeRequest = request;

    request.addEventListener("load", (e) => {
      if (requestID != this.lastAJAXID) {
        return;
      }
      this.activeRequest = null;
      if (request.status == 200) {
        let contentType = request.getResponseHeader("Content-Type");
        if (contentType == "application/json") {
          //application/json
          var data = JSON.parse(request.response);
          if (data.RedirectionLocation) {
            window.location = data.RedirectionLocation;
          } else {
            this.progress.classList.add("hidden");
            this.setFormErrors(data.Errors);
            this.setItemErrors(data.ItemErrors);
            if (data.AfterContent) this.setAfterContent(data.AfterContent);
          }
        } else {
          // Step 2: Create a Blob from the response
          var blob = new Blob([request.response], {
            type: "application/octet-stream",
          });

          // Step 3: Create a URL for the Blob
          var downloadUrl = URL.createObjectURL(blob);

          // Step 4: Create an anchor (<a>) element
          var a = document.createElement("a");
          a.href = downloadUrl;
          a.download = "data.xlsx"; // Set the file name

          // Step 5: Simulate a click on the anchor element
          document.body.appendChild(a); // Append the anchor to document
          a.click();

          // Step 6: Cleanup
          document.body.removeChild(a);
          URL.revokeObjectURL(downloadUrl);
          this.progress.classList.add("hidden");
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

  makeid(length: number) {
    var result = "";
    var characters =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    var charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
  }
}
