function bindRelations() {
  var elements = document.querySelectorAll(".admin_item_relation");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    new RelationPicker(<HTMLDivElement>el);
  });
}

class RelationPicker {

  input: HTMLInputElement;
  previewContainer: HTMLDivElement;
  progress: HTMLProgressElement;

  changeSection: HTMLDivElement;
  changeButton: HTMLDivElement;
  
  picker: HTMLDivElement;
  pickerInput: HTMLInputElement;

  suggestionsEl: HTMLDivElement;

  suggestions: any;

  relationName: string;
  
  constructor(el: HTMLDivElement) {
    this.input = <HTMLInputElement>el.getElementsByTagName("input")[0];
    this.previewContainer = <HTMLDivElement>el.querySelector(".admin_item_relation_preview");
    this.relationName = el.getAttribute("data-relation");
    this.progress = el.querySelector("progress");

    this.changeSection = <HTMLDivElement>el.querySelector(".admin_item_relation_change");
    this.changeButton = <HTMLDivElement>el.querySelector(".admin_item_relation_change_btn");
    this.changeButton.addEventListener("click", () => {
      this.showSearch();
      this.pickerInput.focus();
    });

    this.suggestionsEl = <HTMLDivElement>el.querySelector(".admin_item_relation_picker_suggestions_content");
    this.suggestions = [];
    
    this.picker = <HTMLDivElement>el.querySelector(".admin_item_relation_picker");
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
    this.pickerInput.addEventListener("keydown", this.suggestionInput.bind(this));

    this.getData();
  }

  getData() {
    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/preview/" + this.relationName + "/" + this.input.value, true);

    request.addEventListener("load", () => {
      this.progress.classList.add("hidden");
      if (request.status == 200) {
        this.showPreview(JSON.parse(request.response));
      } else {
        this.showSearch();
      }
    })
    request.send();
  }

  showPreview(data: any) {
    this.previewContainer.textContent = "";
    this.input.value = data.ID;
    var el = this.createPreview(data, true);
    this.previewContainer.appendChild(el);

    this.previewContainer.classList.remove("hidden");
    this.changeSection.classList.remove("hidden");
    this.picker.classList.add("hidden");
  }

  showSearch() {
    this.previewContainer.classList.add("hidden");
    this.changeSection.classList.add("hidden");
    this.picker.classList.remove("hidden");

    this.suggestions = [];
    this.suggestionsEl.innerText = "";
    this.pickerInput.value = "";
  }

  getSuggestions(q: string) {
    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/search/" + this.relationName + "?q=" + encodeURIComponent(q), true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        if (q != this.pickerInput.value) {
          return;
        }
        var data = JSON.parse(request.response);
        this.suggestions = data;
        this.suggestionsEl.innerText = "";
        for (var i = 0; i < data.length; i++) {
          var item = data[i];
          var el = this.createPreview(item, false);
          el.classList.add("admin_item_relation_picker_suggestion")
          el.setAttribute("data-position", i + "");
          el.addEventListener("mousedown", this.suggestionClick.bind(this));
          el.addEventListener("mouseenter", this.suggestionSelect.bind(this));
          this.suggestionsEl.appendChild(el);
        }
      } else {
        console.log("Error while searching");
      }
    })
    request.send();
  }

  suggestionClick() {
    var selected = this.getSelected();
    if (selected >= 0) {
      this.showPreview(this.suggestions[selected]);
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
    this.suggestionsEl.querySelectorAll(".admin_preview")[i].classList.add(this.selectedClass);
  }


  suggestionInput(e: any) {
    switch (e.keyCode) {
      case 13: //enter
        this.suggestionClick()
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
          i += 1
          i = i % this.suggestions.length;
        } else {
          i = 0;
        }
        this.select(i)
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

    var image = document.createElement("div");
    image.classList.add("admin_preview_image");
    image.setAttribute("style", "background-image: url('" + data.Image + "');");

    var right = document.createElement("div");
    right.classList.add("admin_preview_right");

    var name = document.createElement("div");
    name.classList.add("admin_preview_name");
    name.textContent = data.Name;

    var description = document.createElement("description");
    description.classList.add("admin_preview_description");
    description.textContent = data.Description;


    ret.appendChild(image);
    right.appendChild(name);
    right.appendChild(description);
    ret.appendChild(right);
    return ret;
  }

}

function bindRelationsOLD() {
  function bindRelation(el: HTMLElement) {
    var input: HTMLInputElement = <HTMLInputElement>el.getElementsByTagName("input")[0];

    var relationName = input.getAttribute("data-relation");
    var originalValue = input.value;

    var select = document.createElement("select");
    select.classList.add("input");
    select.classList.add("form_input");

    select.addEventListener("change", function() {
      input.value = select.value;
    });

    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/resource/" + relationName, true);

    var progress = el.getElementsByTagName("progress")[0];

    request.addEventListener("load", () => {
      if (request.status >= 200 && request.status < 400) {
        var resp = JSON.parse(request.response);
        addOption(select, "0", "", false);

        Array.prototype.forEach.call(resp, function (item: any, i: number){
          var selected = false;
          if (originalValue == item["id"]) {
            selected = true;
          }
          addOption(select, item["id"], item["name"], selected);
        });
        el.appendChild(select);
      } else {
        console.error("Error wile loading relation " + relationName + ".");
      }
      progress.style.display = 'none';
    });

    request.onerror = function() {
      console.error("Error wile loading relation " + relationName + ".");
      progress.style.display = 'none';
    };
    request.send();
  }

  function addOption(select: HTMLSelectElement, value: string, description: string, selected: boolean) {
    var option = document.createElement("option");
    if (selected) {
      option.setAttribute("selected", "selected");
    }
    option.setAttribute("value", value);
    option.innerText = description;
    select.appendChild(option);
  }

  var elements = document.querySelectorAll(".admin_item_relation");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    bindRelation(el);
  });
}