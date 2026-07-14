class Suggestions {
  pickerInput: HTMLInputElement;
  suggestionsEl: HTMLDivElement;
  suggestions: any;
  searchURL: (q: string) => string;
  returnData: (data: any) => void;

  constructor(
    suggestionEl: HTMLDivElement,
    pickerInput: HTMLInputElement,
    searchURL: (q: string) => string,
    returnData: (data: any) => void
  ) {
    this.suggestionsEl = suggestionEl;

    this.suggestions = [];

    this.pickerInput = pickerInput;
    this.searchURL = searchURL;
    this.returnData = returnData;

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
      this.suggestionInput.bind(this),
    );
  }

  getSuggestions(q: string) {
    var request = new XMLHttpRequest();
    request.open("GET", this.searchURL(q), true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        if (q != this.pickerInput.value) {
          return;
        }
        var data = JSON.parse(request.response);
        this.suggestions = data.Suggestions;
        this.suggestionsEl.innerText = "";

        if (data.Message) {
          let messageEl = document.createElement("div");
          messageEl.innerText = data.Message;
          messageEl.classList.add("picker_message");
          this.suggestionsEl.appendChild(messageEl);
        }

        for (var i = 0; i < data.Suggestions.length; i++) {
          var item = data.Suggestions[i];
          var el = createSuggestionsPreviewEl(item, false);
          el.classList.add("picker_suggestion");
          /*el.addEventListener("mouseleave", () => {
            this.unselect();
          });*/
          el.setAttribute("data-position", i + "");
          el.addEventListener("mousedown", (e: Event) => {
            e.preventDefault();
          });
          el.addEventListener("click", this.suggestionClick.bind(this));
          el.addEventListener("mouseenter", this.suggestionSelect.bind(this));
          this.suggestionsEl.appendChild(el);
        }

        if (data.Button) {
          let buttonEl = document.createElement("a");

          let buttonElIcon = document.createElement("img");
          buttonElIcon.setAttribute(
            "src",
            "/admin/api/icons?file=glyphicons-basic-371-plus.svg&color=${getBaseColor}",
          );
          buttonElIcon.classList.add("btn_icon");

          let buttonElText = document.createElement("span");
          buttonElText.innerText = data.Button.Name;

          buttonEl.appendChild(buttonElIcon);
          buttonEl.appendChild(buttonElText);

          buttonEl.classList.add("btn", "picker_button");
          buttonEl.addEventListener("click", (e) => {
            this.suggestionsEl.classList.add("hidden");
            let popupForm = new PopupForm(data.Button.FormURL, (data: any) => {
              this.returnData(data.Data);
            });
            e.preventDefault();
            e.stopPropagation();
          });
          buttonEl.addEventListener("mousedown", (e) => {
            e.preventDefault();
            e.stopPropagation();
          });
          this.suggestionsEl.appendChild(buttonEl);
        }

        initTooltips();

        this.scrollTop();
      } else {
        console.log("Error while searching");
      }
    });
    request.send();
  }

  suggestionClick() {
    var selected = this.getSelected();
    if (selected >= 0) {
      this.returnData(this.suggestions[selected]);
    }
  }

  suggestionSelect(e: any) {
    var target = <HTMLDivElement>e.currentTarget;
    var position = parseInt(target.getAttribute("data-position"));
    this.select(position);
  }

  selectedClass = "picker_suggestion-selected";
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
    let selectEl = <HTMLDivElement>this.suggestionsEl.querySelectorAll(".preview")[i];
    selectEl.classList.add(this.selectedClass);
    scrollToChild(selectEl);
  }

  suggestionInput(e: any) {
    switch (e.keyCode) {
      case 13: //enter
        this.suggestionClick();
        e.preventDefault();
        return true;
      case 27: //enter
        //this.suggestionClick();
        this.clear();
        e.preventDefault();
        return true;
      case 38: //up
        if (this.suggestionsCount() == 0) {
          return
        }
        var i = this.getSelected();
        if (i < 1) {
          //i = this.suggestions.length - 1;
          i = 0;
        } else {
          i = i - 1;
        }
        this.select(i);
        e.preventDefault();
        return false;
      case 40: //down
        if (this.suggestionsCount() == 0) {
          return
        }
        var i = this.getSelected();
        if (i >= 0) {
          i += 1;
          if (i > this.suggestionsCount() - 1) {
            i = this.suggestionsCount() - 1;
          }
        } else {
          i = 0;
        }
        this.select(i);
        e.preventDefault();
        return false;
    }
  }

  suggestionsCount(): number {
    return this.suggestions.length;
  }

  clear() {
    this.suggestions = [];
    this.suggestionsEl.innerText = "";
    this.pickerInput.value = "";
  }

  focus() {
    this.pickerInput.focus();
  }

  scrollTop() {
    this.suggestionsEl.scrollTo({top: 0});
  }
}

function createSuggestionsPreviewEl(
  data: any,
  anchor: boolean,
): HTMLDivElement {
  var ret = document.createElement("div");
  if (anchor) {
    ret = <any>document.createElement("a");
  }
  ret.classList.add("preview");
  ret.setAttribute("href", data.URL);

  var right = document.createElement("div");
  right.classList.add("preview_right");

  var name = document.createElement("div");
  name.classList.add("preview_name");
  name.textContent = data.Name;

  var description = document.createElement("div");
  description.classList.add("preview_description");
  description.setAttribute("title", data.Description);
  description.textContent = data.Description;

  if (data.Image) {
    let image = document.createElement("img");
    image.classList.add("preview_image");
    image.setAttribute("src", data.Image);
    image.setAttribute("loading", "lazy");
    ret.appendChild(image);
  } else {
    //let imageDiv = document.createElement("div");
    //imageDiv.classList.add("preview_image");
    //ret.appendChild(imageDiv);
  }

  right.appendChild(name);
  right.appendChild(description);
  ret.appendChild(right);
  return ret;
}
