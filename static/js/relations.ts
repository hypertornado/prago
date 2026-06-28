class RelationPicker {
  el: HTMLDivElement;
  input: HTMLInputElement;

  previewsContainer: HTMLDivElement;
  progress: HTMLProgressElement;

  picker: HTMLDivElement;

  relationName: string;
  filterID: string;

  multipleInputs: boolean;
  autofocus: boolean;

  suggestionsObject: Suggestions;

  constructor(el: HTMLDivElement) {
    this.el = el;

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
      el.querySelector(".relation_input_previews")
    );
    this.relationName = el.getAttribute("data-relation");
    this.filterID = el.getAttribute("data-filter");
    this.progress = el.querySelector("progress");

    this.picker = <HTMLDivElement>(
      el.querySelector(".picker")
    );
    
    this.suggestionsObject = new Suggestions(
      this.el.querySelector(".picker_suggestions_content"),
      this.picker.querySelector("input"),
      this.getSearchURL.bind(this),
      (data: any) => {
        this.addPreview(data, true);
      }
    );

    this.makeReordable();

    if (this.multipleInputs || parseInt(this.input.value) > 0) {
      this.getData();
    } else {
      this.progress.classList.add("hidden");
      this.showSearch();
    }
  }

  makeReordable() {
    makeReorderable(this.previewsContainer, () => {
      this.updateLayout();
    });
  }

  getSearchURL(q: string): string {
    var encoded = encodeParams({
      q: q,
      filter: this.filterID,
      resource: this.relationName,
    });

    return "/admin/api/_suggestionsresource" + encoded;
  }

  getData() {
    if (!this.input.value) {
      this.progress.classList.add("hidden");
      this.showSearch();
      return;
    }

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
          this.addPreview(items[i], false);
        }
        if (items.length == 0) {
          this.showSearch();
        }
        this.updateInput();
      } else {
        this.showSearch();
      }
    });
    request.send();
  }

  addPreview(data: any, animate: boolean) {
    let previewEl = document.createElement("div");
    previewEl.classList.add("relation_input_preview");
    if (animate) {
      previewEl.classList.add("relation_input_preview-insert")
    }

    var el = createSuggestionsPreviewEl(data, true);
    this.previewsContainer.appendChild(previewEl);
    previewEl.appendChild(el);

    let deleteButton = document.createElement("div");
    deleteButton.classList.add("btn");
    deleteButton.classList.add("btn-formitem");
    deleteButton.classList.add("relation_input_preview_action");

    deleteButton.innerHTML = `
      <img src="/admin/api/icons?file=glyphicons-basic-599-menu-close.svg&color=base" class="btn_icon">
    `
    
    previewEl.appendChild(deleteButton);
    deleteButton.addEventListener("click", () => {
      el.classList.add("relation_input_preview-remove");
      setTimeout(() => {
        previewEl.remove();
        this.updateLayout();
        this.suggestionsObject.focus();
      }, 200);
    });

    previewEl.setAttribute("data-id", data.ID);
    this.suggestionsObject.clear();
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
    if (this.multipleInputs && val != "") {
      val = ";" + val + ";";
    }
    this.input.value = val;
  }

  showSearch() {
    this.picker.classList.remove("hidden");
    this.suggestionsObject.clear();
    if (this.autofocus) {
      this.suggestionsObject.focus();
    }
  }
}
