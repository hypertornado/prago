class ListMultipleEdit {
  listMultiple: ListMultiple;
  form: HTMLFormElement;
  popup: ContentPopup;

  constructor(multiple: ListMultiple, ids: Array<String>) {
    this.listMultiple = multiple;
    var typeID = document
      .querySelector(".admin_list")
      .getAttribute("data-type");

    var progress = document.createElement("progress");
    this.popup = new ContentPopup(
      `Hromadná úprava položek (${ids.length} položek)`,
      progress
    );
    this.popup.show();

    fetch("/admin/" + typeID + "/api/multiple_edit?ids=" + ids.join(","))
      .then((response) => {
        if (response.ok) {
          return response.text();
        } else {
          this.popup.hide();
          new Alert("Operaci nelze nahrát.");
        }
      })
      .then((val) => {
        var div = document.createElement("div");
        div.innerHTML = val;
        this.popup.setContent(div);
        this.initFormPopup(<HTMLFormElement>div.querySelector("form"));
        this.popup.setConfirmButtons(this.confirm.bind(this));
      });

    //this.initFormPopup(el);
  }

  initFormPopup(form: HTMLFormElement) {
    this.form = form;
    this.form.addEventListener("submit", this.confirm.bind(this));
    new Form(this.form);
    this.initCheckboxes();
  }

  initCheckboxes() {
    var checkboxes = this.form.querySelectorAll<HTMLInputElement>(
      ".multiple_edit_field_checkbox"
    );
    checkboxes.forEach((cb) => {
      cb.addEventListener("change", (e) => {
        var item = cb.parentElement.parentElement;
        if (cb.checked) {
          item.classList.add("multiple_edit_field-selected");
        } else {
          item.classList.remove("multiple_edit_field-selected");
        }
      });
    });
  }

  confirm(e: Event) {
    var typeID = document
      .querySelector(".admin_list")
      .getAttribute("data-type");
    var data = new FormData(this.form);

    var loader = new LoadingPopup();

    fetch("/admin/" + typeID + "/api/multiple_edit", {
      method: "POST",
      body: data,
    }).then((response) => {
      loader.done();
      if (response.ok) {
        this.popup.hide();
        this.listMultiple.list.load();
      } else {
        if (response.status == 403) {
          response.json().then((data) => {
            new Alert(data.error.Text);
          });
          return;
        } else {
          new Alert("Chyba při ukládání.");
        }
      }
    });

    /*
    var req = new XMLHttpRequest();
    req.open("POST", "/admin/" + typeID + "/api/multiple_edit", true);
    req.addEventListener("load", () => {
      console.log("loaded");
    });
    req.send(data);*/

    e.preventDefault();
  }
}
