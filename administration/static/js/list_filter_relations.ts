class ListFilterRelations {
  valueInput: HTMLInputElement;
  input: HTMLInputElement
  preview: HTMLDivElement;
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
    this.previewName = el.querySelector(".filter_relations_preview_name");
    this.previewClose = el.querySelector(".filter_relations_preview_close");

    this.previewClose.addEventListener("click", this.closePreview.bind(this));


    this.preview.classList.add("hidden");

    let hiddenEl: HTMLInputElement = el.querySelector("input");

    this.relatedResourceName = el.querySelector(".admin_table_filter_item-relations").getAttribute("data-related-resource");

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
    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    request.open("GET", adminPrefix + "/_api/preview/" + this.relatedResourceName + "/" + value, true);

    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.renderPreview(JSON.parse(request.response));
      } else {
        console.error("not found");
      }
    })
    request.send();
  }

  renderPreview(item: any) {
    this.valueInput.value = item.ID;
    this.preview.classList.remove("hidden");
    this.search.classList.add("hidden");
    this.previewName.textContent = item.Name;
    this.dispatchChange();
  }

  dispatchChange() {
    var event = new Event('change');
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
    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    request.open("GET", adminPrefix + "/_api/search/" + this.relatedResourceName + "?q=" + encodeURIComponent(q), true);

    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.renderSuggestions(JSON.parse(request.response));
      } else {
        console.error("not found");
      }
    })
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

    var image = document.createElement("div");
    image.classList.add("list_filter_suggestion_image");
    image.setAttribute("style", "background-image: url('" + data.Image + "');");

    var right = document.createElement("div");
    right.classList.add("list_filter_suggestion_right");

    var name = document.createElement("div");
    name.classList.add("list_filter_suggestion_name");
    name.textContent = data.Name;

    var description = document.createElement("div");
    description.classList.add("list_filter_suggestion_description");
    description.textContent = data.Description;


    ret.appendChild(image);
    right.appendChild(name);
    right.appendChild(description);
    ret.appendChild(right);
    return ret;
  }

}