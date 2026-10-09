//TODO: does not work with image picker and other hidden elements
class Form {
  private dirty: boolean = false;
  formEl: HTMLFormElement;
  lastChanged: number;
  changeHandler: any;
  willChangeHandler: any;
  formContainer: FormContainer;

  constructor(form: HTMLFormElement, formContainer: FormContainer) {
    this.dirty = false;
    this.formEl = form;
    this.formContainer = formContainer;

    this.fixAufofocus();

    var elements = form.querySelectorAll<HTMLDivElement>(".markdown_input");
    elements.forEach((el) => {
      new MarkdownEditor(el);
    });

    var timestamps = form.querySelectorAll<HTMLDivElement>(".input_timestamp");
    timestamps.forEach((form) => {
      new Timestamp(form);
    });

    var relations = form.querySelectorAll<HTMLDivElement>(
      ".relation_input"
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

    this.initSuggestions();
    this.initConflictCheck();
    this.initPlusMinus();
    this.initCalendar();
    this.initDaterange();
    this.initShowpassword();
    initTooltips();

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

  initConflictCheck() {
    let version = this.formEl.getAttribute("data-item-version");
    if (!version) {
      return;
    }
    if (version == "0") {
      return;
    }
    this.versionCheck(version);
    window.setInterval(() => {
      this.versionCheck(version);
    }, 2000);
  }

  versionCheck(version: string) {
    if (this.formContainer.activeRequest) {
      return;
    }
    var request = new XMLHttpRequest();
    request.open("POST", "/admin/api/_conflict?version="+version, true);
    request.addEventListener("load", (e: any) => {
        if (request.status != 200) {
            console.error("Error while saving order.");
            return;
        }

        var data = JSON.parse(request.response);
        let conflictEl = <HTMLDivElement>this.formEl.querySelector(".form_conflict");

        if (this.formEl.classList.contains("form-loading")) {
          conflictEl.classList.add("hidden");
          return;
        };

        if (data.Show) {
          conflictEl.innerText = data.Text;
          conflictEl.classList.remove("hidden");
        } else {
          conflictEl.classList.add("hidden");
        }
    });
    request.send();
  }

  initSuggestions() {
    let suggestionEls = this.formEl.querySelectorAll(".form_suggestions");
    for (var i = 0; i < suggestionEls.length; i++) {
      let suggestionEl = <HTMLDivElement>suggestionEls[i];
      let parentEl = suggestionEl.parentElement;
      let inputEl = <HTMLInputElement>parentEl.querySelector(".form_input");
      let uuid = suggestionEl.getAttribute("data-suggestions-custom-id");
      let contextFieldsStr = suggestionEl.getAttribute("data-suggestions-custom-context-fields");

      let contextFieldsArr: string[] = [];
      if (contextFieldsStr) {
        contextFieldsArr = contextFieldsStr.split(",")
      }

      new Suggestions(suggestionEl.querySelector(".picker_suggestions_content"), inputEl, (q: string): any => {
        return {
          CustomUUID: uuid,
          Q: q,
          ContextValues: this.getContextMap(contextFieldsArr),
        }
      }, (data: any) => {
        inputEl.value = data.ID;
        if (data.SetValues) {
          for (let key in data.SetValues) {
            this.setFieldValue(key, data.SetValues[key]);
          }
        }
        inputEl.blur();
      })
    }
  }

  getContextMap(fields: string[]): Record<string, any> {
    let ret: Record<string, any> = {};
    for (let i = 0; i < fields.length; i++) {
      let fieldID = fields[i];
      ret[fieldID] = this.getFieldValue(fieldID);
    }
    return ret;
  }

  getFieldInput(id: string): HTMLInputElement {
    let labelEl = this.formEl.querySelector(`.form_label[data-id=${id}]`);
    if (!labelEl) {
      return null;
    }
    return labelEl.querySelector(".form_input");
  }

  getFieldValue(id: string): any {
    let input = this.getFieldInput(id);
    if (!input) {
      return null;
    }
    return input.value;
  }

  setFieldValue(id: string, value: string) {
    let input = this.getFieldInput(id);
    if (!input) {
      return;
    }
    input.value = value;
  }

  initPlusMinus() {
    let els = this.formEl.querySelectorAll(".form_item_plus,.form_item_minus");
    for (var i = 0; i < els.length; i++) {
      let btnEl = els[i];
      btnEl.addEventListener("mousedown", (e: Event) => {
        let input = btnEl.parentElement.querySelector("input");
        if (btnEl.classList.contains("form_item_plus")) {
          input.stepUp();
        } else {
          input.stepDown();
        }
        input.focus();
        e.preventDefault();
      })
    }
  }

  initCalendar() {
    let els = this.formEl.querySelectorAll(".form_item_calendar");
    for (var i = 0; i < els.length; i++) {
      let btnEl = els[i];
      btnEl.addEventListener("mousedown", (e: Event) => {
        let input = <HTMLInputElement>btnEl.parentElement.querySelector("input[type=date]");
        input.showPicker();
        input.focus();
        e.preventDefault();
      })
    }
  }

  initDaterange() {
    let els = this.formEl.querySelectorAll(".form_daterange");
    for (var i = 0; i < els.length; i++) {
      let el = <HTMLDivElement>els[i];
      new FormDateRange(el);
    }
  }

  initShowpassword() {
    let els = this.formEl.querySelectorAll(".form_item_showpassword");
    for (var i = 0; i < els.length; i++) {
      let btnEl = els[i];
      btnEl.addEventListener("mousedown", (e: Event) => {
        let input = <HTMLInputElement>btnEl.parentElement.querySelector("input");
        if (input.getAttribute("type") == "password") {
          input.setAttribute("type", "text");
        } else {
          input.setAttribute("type", "password");
        }
        //input.showPicker();
        input.focus();
        e.preventDefault();
      })
    }
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
