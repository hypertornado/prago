
var primaryFormContainer: any

class FormContainer {
  formContainer: HTMLDivElement;
  form: Form;
  progress: HTMLProgressElement;
  lastAJAXID: string;
  activeRequest: XMLHttpRequest;
  okHandler: Function;

  constructor(formContainer: HTMLDivElement, okHandler: Function) {
    if (!window.primaryFormContainer && !formContainer.parentElement.classList.contains("popup_content")) {
      window.primaryFormContainer = this;
    }

    this.formContainer = formContainer;
    this.okHandler = okHandler;
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

    let requestID: string = makeid(10);
    this.lastAJAXID = requestID;

    if (this.activeRequest) {
      if (this.isAutosubmit()) {
        this.activeRequest.abort();
      } else {
        return;
      }
    }
    this.activeRequest = request;

    this.form.formEl.classList.remove("form-errors");

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
          if (data.RedirectionLocation || data.Preview || data.Data) {
            this.okHandler(data);
            //window.location = data.RedirectionLocation;
          } else {
            this.progress.classList.add("hidden");
            this.setFormErrors(data.Errors);
            if (data.AfterContent) this.setAfterContent(data.AfterContent);
            initTables();
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
      }
    });

    this.progress.classList.remove("hidden");
    request.send(formData);
  }

  setAfterContent(text: string) {
    this.formContainer.querySelector(".form_after_content").innerHTML = text;
  }

  setFormErrors(errors: any[]) {
    this.deleteItemErrors();

    let errorsDiv: HTMLDivElement =
      this.form.formEl.querySelector(".form_errors");
    errorsDiv.innerText = "";
    errorsDiv.classList.add("hidden");

    var anyError = false;

    if (errors) {
      for (let i = 0; i < errors.length; i++) {
        if (errors[i].Field) {
          anyError = true;
          this.setItemError(errors[i]);
        } else {
          let errorDiv = document.createElement("div");
          errorDiv.classList.add("form_errors_item");
          if (errors[i].OK) {
            errorDiv.classList.add("form_errors_item-ok");
          } else {
            anyError = true;
            errorDiv.classList.add("form_errors_item-error");
          }
          errorDiv.innerText = errors[i].Text;
          errorsDiv.appendChild(errorDiv);
          errorsDiv.classList.remove("hidden");
        }
      }
      if (anyError) {
        this.form.formEl.classList.add("form-errors");
      }
    }
  }

  deleteItemErrors() {
    let labels = this.form.formEl.querySelectorAll(".form_label");
    for (let i = 0; i < labels.length; i++) {
      let label = labels[i];
      label.classList.remove("form_label-errors");
      let labelErrors = label.querySelector(".form_label_errors");
      labelErrors.innerHTML = "";
      labelErrors.classList.add("hidden");
    }
  }

  setItemError(itemError: any) {
    let labels = this.form.formEl.querySelectorAll(".form_label");
    for (let i = 0; i < labels.length; i++) {
      let label = labels[i];
      let id = label.getAttribute("data-id");
      if (label.getAttribute("data-id") == itemError.Field) {
        label.classList.add("form_label-errors");
        let labelErrors = label.querySelector(".form_label_errors");
        labelErrors.classList.remove("hidden");
        let errorDiv = document.createElement("div");
        errorDiv.classList.add("form_label_errors_error");
        errorDiv.innerText = itemError.Text;
        labelErrors.appendChild(errorDiv);
      }
    }
  }
}

function makeid(length: number) {
  var result = "";
  var characters =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  var charactersLength = characters.length;
  for (var i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return result;
}
