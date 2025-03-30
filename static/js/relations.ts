class RelationPicker {
  input: HTMLInputElement;

  previewsContainer: HTMLDivElement;

  //TODO: remove this element

  progress: HTMLProgressElement;

  //changeSection: HTMLDivElement;
  //changeButton: HTMLDivElement;

  picker: HTMLDivElement;
  pickerInput: HTMLInputElement;

  suggestionsEl: HTMLDivElement;

  suggestions: any;

  relationName: string;

  multipleInputs: boolean;
  autofocus: boolean;

  constructor(el: HTMLDivElement) {
    if (el.getAttribute("data-autofocus") == "true") {
      this.autofocus = true
    }

    if (el.getAttribute("data-multiple") == "true") {
      this.multipleInputs = true;
    } else {
      this.multipleInputs = false;
    }

    this.input = <HTMLInputElement>el.getElementsByTagName("input")[0];
    this.previewsContainer = <HTMLDivElement>(
      el.querySelector(".admin_relation_previews")
    );
    this.relationName = el.getAttribute("data-relation");
    this.progress = el.querySelector("progress");

    this.suggestionsEl = <HTMLDivElement>(
      el.querySelector(".admin_item_relation_picker_suggestions_content")
    );
    this.suggestions = [];

    this.picker = <HTMLDivElement>(
      el.querySelector(".admin_item_relation_picker")
    );
    this.pickerInput = this.picker.querySelector("input");
    this.pickerInput.addEventListener("input", () => {
      this.getSuggestions(this.pickerInput.value);
    });
    this.pickerInput.addEventListener("blur", () => {
      this.suggestionsEl.classList.add("hidden");
    });
    this.pickerInput.addEventListener("focus", () => {
      this.suggestionsEl.classList.remove("hidden");
      this.getSuggestions(this.pickerInput.value);
    });
    this.pickerInput.addEventListener(
      "keydown",
      this.suggestionInput.bind(this)
    );

    if (parseInt(this.input.value) > 0) {
      this.getData();
    } else {
      this.progress.classList.add("hidden");
      this.showSearch();
    }
  }

  getData() {
    var request = new XMLHttpRequest();
    request.open(
      "GET",
      "/admin/" +
        this.relationName +
        "/api/preview-relation/" +
        this.input.value,
      true
    );

    request.addEventListener("load", () => {
      this.progress.classList.add("hidden");
      if (request.status == 200) {
        let items = JSON.parse(request.response);
        for (var i = 0; i < items.length; i++) {
          this.addPreview(items[i]);
        }
      } else {
        this.showSearch();
      }
    });
    request.send();
  }

  addPreview(data: any) {
    let previewEl = document.createElement("div");
    previewEl.classList.add("admin_relation_preview");

    var el = this.createPreview(data, true);
    this.previewsContainer.appendChild(previewEl);
    previewEl.appendChild(el);

    let upButton = document.createElement("div");
    upButton.classList.add(
      "admin_relation_preview_action",
      "admin_relation_preview_action-up"
    );
    upButton.innerText = "↑";
    previewEl.appendChild(upButton);
    upButton.addEventListener("click", (e: Event) => {
      this.updateOrder(e, false);
    });

    let downButton = document.createElement("div");
    downButton.classList.add(
      "admin_relation_preview_action",
      "admin_relation_preview_action-down"
    );
    downButton.innerText = "↓";
    previewEl.appendChild(downButton);
    downButton.addEventListener("click", (e: Event) => {
      this.updateOrder(e, true);
    });

    let deleteButton = document.createElement("div");
    deleteButton.classList.add("admin_relation_preview_action");
    deleteButton.innerText = "×";
    previewEl.appendChild(deleteButton);
    deleteButton.addEventListener("click", () => {
      previewEl.remove();
      this.updateLayout();
    });

    previewEl.setAttribute("data-id", data.ID);

    this.pickerInput.value = "";
    this.updateLayout();
  }

  numberOfItems(): number {
    return this.previewsContainer.children.length;
  }

  updateOrder(e: Event, down: boolean) {
    let target = <HTMLDivElement>e.target;
    let previewEl = target.parentElement;
    let sibling: Element;
    if (down) {
      sibling = previewEl.nextElementSibling;
    } else {
      sibling = previewEl.previousElementSibling;
    }
    if (!sibling) {
      return;
    }
    let parent = previewEl.parentElement;
    if (down) {
      parent.insertBefore(sibling, previewEl);
    } else {
      parent.insertBefore(previewEl, sibling);
    }
    this.updateLayout();
  }

  updateLayout() {
    if (this.multipleInputs || this.numberOfItems() == 0) {
      this.picker.classList.remove("hidden");
    } else {
      this.picker.classList.add("hidden");
    }
    this.updateInput();
  }

  updateInput() {
    var valItems = [];
    for (var i = 0; i < this.previewsContainer.children.length; i++) {
      let child = this.previewsContainer.children[i];
      let val = child.getAttribute("data-id");
      valItems.push(val);
    }
    let val = valItems.join(";");
    if (this.multipleInputs) {
      val = ";" + val + ";";
    }
    this.input.value = val;
  }

  showSearch() {
    this.picker.classList.remove("hidden");
    this.suggestions = [];
    this.suggestionsEl.innerText = "";
    this.pickerInput.value = "";
    if (this.autofocus) {
      this.pickerInput.focus();
    }
  }

  getSuggestions(q: string) {
    var request = new XMLHttpRequest();
    request.open(
      "GET",
      "/admin/" +
        this.relationName +
        "/api/searchresource" +
        "?q=" +
        encodeURIComponent(q),
      true
    );
    request.addEventListener("load", () => {
      if (request.status == 200) {
        if (q != this.pickerInput.value) {
          return;
        }
        var data = JSON.parse(request.response);
        this.suggestions = data.Previews;
        this.suggestionsEl.innerText = "";

        if (data.Message) {
          let messageEl = document.createElement("div");
          messageEl.innerText = data.Message;
          messageEl.classList.add("relation_message");
          this.suggestionsEl.appendChild(messageEl);
        }

        for (var i = 0; i < data.Previews.length; i++) {
          var item = data.Previews[i];
          var el = this.createPreview(item, false);
          el.classList.add("admin_item_relation_picker_suggestion");
          el.setAttribute("data-position", i + "");
          el.addEventListener("mousedown", this.suggestionClick.bind(this));
          el.addEventListener("mouseenter", this.suggestionSelect.bind(this));
          this.suggestionsEl.appendChild(el);
        }

        if (data.Button) {
          let buttonEl = document.createElement("a");
          buttonEl.innerText = data.Button.Name;
          buttonEl.setAttribute("href", data.Button.URL);
          buttonEl.classList.add("btn", "relation_button");
          buttonEl.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
          })
          buttonEl.addEventListener("mousedown", (e) => {
            this.suggestionsEl.classList.add("hidden");
            let popupForm = new PopupForm(data.Button.FormURL, (data: any) => {
              this.addPreview(data.Data);
            });
            e.preventDefault();
            e.stopPropagation();
          })
          this.suggestionsEl.appendChild(buttonEl);
        }
      } else {
        console.log("Error while searching");
      }
    });
    request.send();
  }

  suggestionClick() {
    var selected = this.getSelected();
    if (selected >= 0) {
      this.addPreview(this.suggestions[selected]);
    }
  }

  suggestionSelect(e: any) {
    var target = <HTMLDivElement>e.currentTarget;
    var position = parseInt(target.getAttribute("data-position"));
    this.select(position);
  }

  selectedClass = "admin_item_relation_picker_suggestion-selected";

  getSelected(): number {
    var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
    if (!selected) {
      return -1;
    }
    return parseInt(selected.getAttribute("data-position"));
  }

  unselect(): number {
    var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
    if (!selected) {
      return -1;
    }
    selected.classList.remove(this.selectedClass);
    return parseInt(selected.getAttribute("data-position"));
  }

  select(i: number) {
    this.unselect();
    this.suggestionsEl
      .querySelectorAll(".admin_preview")
      [i].classList.add(this.selectedClass);
  }

  suggestionInput(e: any) {
    switch (e.keyCode) {
      case 13: //enter
        this.suggestionClick();
        e.preventDefault();
        return true;
      case 38: //up
        var i = this.getSelected();
        if (i < 1) {
          i = this.suggestions.length - 1;
        } else {
          i = i - 1;
        }
        this.select(i);
        e.preventDefault();
        return false;
      case 40: //down
        var i = this.getSelected();
        if (i >= 0) {
          i += 1;
          i = i % this.suggestions.length;
        } else {
          i = 0;
        }
        this.select(i);
        e.preventDefault();
        return false;
    }
  }

  createPreview(data: any, anchor: boolean): HTMLDivElement {
    var ret = document.createElement("div");
    if (anchor) {
      ret = <any>document.createElement("a");
    }
    ret.classList.add("admin_preview");
    ret.setAttribute("href", data.URL);

    var right = document.createElement("div");
    right.classList.add("admin_preview_right");

    var name = document.createElement("div");
    name.classList.add("admin_preview_name");
    name.textContent = data.Name;

    var description = document.createElement("description");
    description.classList.add("admin_preview_description");
    description.setAttribute("title", data.Description);
    description.textContent = data.Description;

    var image = document.createElement("div");
    image.classList.add("admin_preview_image");
    if (data.Image) {
      image.setAttribute(
        "style",
        "background-image: url('" + data.Image + "');"
      );
    }
    ret.appendChild(image);

    right.appendChild(name);
    right.appendChild(description);
    ret.appendChild(right);
    return ret;
  }
}
