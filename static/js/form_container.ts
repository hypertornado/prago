
var primaryFormContainer: any

class FormContainer {
  formContainer: HTMLDivElement;
  form: Form;
  progress: HTMLProgressElement;
  lastAJAXID: string;
  activeRequest: XMLHttpRequest;
  okHandler: Function;
  formTaskUUID: string;
  lastTaskLoad: number;
  tableTbody: HTMLTableSectionElement;
  taskHeader: HTMLDivElement;

  constructor(formContainer: HTMLDivElement, okHandler: Function) {
    if (!window.primaryFormContainer && !formContainer.parentElement.classList.contains("popup_content")) {
      window.primaryFormContainer = this;
    }

    this.formTaskUUID = "";

    this.formContainer = formContainer;
    this.okHandler = okHandler;
    this.progress = formContainer.querySelector(".form_progress");
    var formEl: HTMLFormElement = formContainer.querySelector("form");
    this.form = new Form(formEl);
    this.form.formEl.addEventListener("submit", this.submitFormAJAX.bind(this));

    this.tableTbody = this.formContainer.querySelector(".form_task_table tbody");
    this.taskHeader = this.formContainer.querySelector(".form_task_header");

    if (this.isAutosubmitFirstTime()) {
      this.sendForm();
    }

    if (this.isAutosubmit()) {
      this.form.changeHandler = this.formChanged.bind(this);
      this.form.willChangeHandler = this.formWillChange.bind(this);
      this.sendForm();
    }

    this.initTaskProgressReader();

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
      this.formTaskUUID = "";
      this.activeRequest = null;
      this.cleanTable();
      if (request.status == 200) {
        let contentType = request.getResponseHeader("Content-Type");
        if (contentType == "application/json") {
          //application/json
          var data = JSON.parse(request.response);
          this.progress.classList.add("hidden");
          this.formTaskUUID = data.TaskUUID;
          if (data.RedirectionLocation || data.Preview || data.Data) {
            this.okHandler(data);
            //window.location = data.RedirectionLocation;
          } else {
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

  initTaskProgressReader() {
    this.lastTaskLoad = Date.now();

    let taskEl = this.formContainer.querySelector(".form_task");

    this.formContainer.querySelector(".form_task_stop").addEventListener("click", () => {
      new PopupForm("/admin/_taskstop?uuid=" + this.formTaskUUID, (data: any) => {
        this.setTaskFinished();
        //this.addUUID(data.Data);
      })
    });

    window.setInterval(() => {
      if (this.formTaskUUID) {
        taskEl.classList.remove("hidden");
      }

      if (this.formTaskUUID != "" && Date.now() - this.lastTaskLoad > 1000) {
        this.lastTaskLoad = Date.now();
        this.loadTaskProgress(this.formTaskUUID);
      }
    }, 100);
  }

  loadTaskProgress(uuid: string) {
    let request = new XMLHttpRequest();
    request.open("GET", "/admin/api/_taskview?uuid=" + uuid);

    request.addEventListener("load", (e) => {
      if (this.formTaskUUID != uuid) {
        return;
      }
      if (request.status == 200) {
        var data = JSON.parse(request.response);
        this.setTaskData(data);
      } else {
        this.setTaskFinished();
      }
    });
    request.send();
  }

  setTaskData(data: any) {
    this.taskHeader.classList.remove("hidden");
    let taskEl = this.formContainer.querySelector(".form_task");
    taskEl.querySelector(".form_task_status").textContent = data.Description;
    taskEl.querySelector(".form_task_progress").setAttribute("value", data.Progress);
    taskEl.querySelector(".form_task_progress_text").textContent = data.ProgressText;

    this.tableTbody.insertAdjacentHTML("beforeend", data.TableRows);
    if (data.Finished) {
      this.setTaskFinished();
    }
  }

  setTaskFinished() {
    this.formTaskUUID = "";
    this.taskHeader.classList.add("hidden");
  }

  cleanTable() {
    this.tableTbody.innerHTML = "";
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
