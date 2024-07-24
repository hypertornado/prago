class ListFilterRelations {
  valueInput: HTMLInputElement;
  input: HTMLInputElement;
  preview: HTMLDivElement;
  previewImage: HTMLDivElement;
  previewName: HTMLDivElement;
  previewClose: HTMLDivElement;
  search: HTMLDivElement;
  suggestions: HTMLDivElement;
  relatedResourceName: string;
  dirty: boolean;
  lastChanged: number;

  constructor(el: HTMLDivElement, value: any, list: List) {
    this.valueInput = el.querySelector(".filter_relations_hidden");
    this.input = el.querySelector(".filter_relations_search_input");
    this.search = el.querySelector(".filter_relations_search");
    this.suggestions = el.querySelector(".filter_relations_suggestions");
    this.preview = el.querySelector(".filter_relations_preview");
    this.previewImage = el.querySelector(".filter_relations_preview_image");
    this.previewName = el.querySelector(".filter_relations_preview_name");
    this.previewClose = el.querySelector(".filter_relations_preview_close");

    this.previewClose.addEventListener("click", this.closePreview.bind(this));

    this.preview.classList.add("hidden");

    let hiddenEl: HTMLInputElement = el.querySelector("input");

    this.relatedResourceName = el
      .querySelector(".list_filter_item-relations")
      .getAttribute("data-related-resource");

    this.input.addEventListener("input", () => {
      this.dirty = true;
      this.lastChanged = Date.now();
      return false;
    });

    window.setInterval(() => {
      if (this.dirty && Date.now() - this.lastChanged > 100) {
        this.loadSuggestions();
      }
    }, 30);

    if (this.valueInput.value) {
      this.loadPreview(this.valueInput.value);
    }
  }

  loadPreview(value: string) {
    var request = new XMLHttpRequest();
    let apiURL =
      "/admin/" + this.relatedResourceName + "/api/preview-relation/" + value;

    request.open("GET", apiURL, true);

    request.addEventListener("load", () => {
      if (request.status == 200) {
        let respData = JSON.parse(request.response);
        if (respData.length > 0) {
          this.renderPreview(respData[0]);
        }
      } else {
        console.error("not found");
      }
    });
    request.send();
  }

  renderPreview(item: any) {
    this.valueInput.value = item.ID;
    this.preview.classList.remove("hidden");
    this.search.classList.add("hidden");
    this.preview.setAttribute("title", item.Name);
    if (item.Image) {
      this.previewImage.classList.remove("hidden");
      this.previewImage.setAttribute(
        "style",
        "background-image: url('" + item.Image + "');"
      );
    } else {
      this.previewImage.classList.add("hidden");
    }
    this.previewName.textContent = item.Name;
    this.dispatchChange();
  }

  dispatchChange() {
    var event = new Event("change");
    this.valueInput.dispatchEvent(event);
  }

  closePreview() {
    this.valueInput.value = "";
    this.preview.classList.add("hidden");
    this.search.classList.remove("hidden");
    this.input.value = "";
    this.suggestions.innerHTML = "";
    this.suggestions.classList.add("filter_relations_suggestions-empty");
    this.dispatchChange();
    this.input.focus();
  }

  loadSuggestions() {
    this.getSuggestions(this.input.value);
    this.dirty = false;
  }

  getSuggestions(q: string) {
    var request = new XMLHttpRequest();
    request.open(
      "GET",
      "/admin/" +
        this.relatedResourceName +
        "/api/searchresource" +
        "?q=" +
        encodeURIComponent(q),
      true
    );

    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.renderSuggestions(JSON.parse(request.response));
      } else {
        console.error("not found");
      }
    });
    request.send();
  }

  renderSuggestions(data: any) {
    this.suggestions.innerHTML = "";
    this.suggestions.classList.add("filter_relations_suggestions-empty");
    for (var i = 0; i < data.length; i++) {
      this.suggestions.classList.remove("filter_relations_suggestions-empty");
      let item = data[i];
      let el = this.renderSuggestion(item);
      this.suggestions.appendChild(el);
      let index = i;
      el.addEventListener("mousedown", (e) => {
        this.renderPreview(item);
      });
    }
  }

  renderSuggestion(data: any): HTMLDivElement {
    var ret = document.createElement("div");

    ret.classList.add("list_filter_suggestion");
    ret.setAttribute("href", data.URL);

    var right = document.createElement("div");
    right.classList.add("list_filter_suggestion_right");

    var name = document.createElement("div");
    name.classList.add("list_filter_suggestion_name");
    name.textContent = data.Name;

    var description = document.createElement("div");
    description.classList.add("list_filter_suggestion_description");
    description.textContent = data.Description;

    if (data.Image) {
      var image = document.createElement("div");
      image.classList.add("list_filter_suggestion_image");
      image.setAttribute(
        "style",
        "background-image: url('" + data.Image + "');"
      );
      ret.appendChild(image);
    }

    right.appendChild(name);
    right.appendChild(description);
    ret.appendChild(right);
    return ret;
  }
}
