class SearchForm {
  searchForm: HTMLFormElement;
  searchInput: HTMLInputElement;
  suggestionsEl: HTMLDivElement;
  suggestions: any;
  dirty: boolean;
  lastChanged: number;

  constructor(el: HTMLFormElement) {
    this.searchForm = el;
    this.searchInput = <HTMLInputElement>(
      el.querySelector(".admin_header_search_input")
    );
    this.suggestionsEl = <HTMLDivElement>(
      el.querySelector(".admin_header_search_suggestions")
    );

    this.searchInput.value = document.body.getAttribute("data-search-query");

    this.searchInput.addEventListener("input", () => {
      this.suggestions = [];
      this.dirty = true;
      this.lastChanged = Date.now();
      return false;
    });

    window.setInterval(() => {
      if (this.dirty && Date.now() - this.lastChanged > 100) {
        this.loadSuggestions();
      }
    }, 30);

    this.searchInput.addEventListener("keydown", (e) => {
      if (!this.suggestions || this.suggestions.length == 0) {
        return;
      }
      switch (e.keyCode) {
        case 13: //enter
          var i = this.getSelected();
          if (i >= 0) {
            var child = this.suggestions[i];
            if (child) {
              window.location.href = child.getAttribute("href");
            }
            e.preventDefault();
            return true;
          }
          return false;
        case 38: //up
          var i = this.getSelected();
          if (i < 1) {
            i = this.suggestions.length - 1;
          } else {
            i = i - 1;
          }
          this.setSelected(i);
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
          this.setSelected(i);
          e.preventDefault();
          return false;
      }
    });
  }

  loadSuggestions() {
    this.dirty = false;
    var suggestText = this.searchInput.value;
    var request = new XMLHttpRequest();

    var adminPrefix = document.body.getAttribute("data-admin-prefix");

    var url =
      adminPrefix +
      "/api/search-suggest" +
      encodeParams({ q: this.searchInput.value });
    request.open("GET", url);
    request.addEventListener("load", () => {
      if (suggestText != this.searchInput.value) {
        return;
      }
      if (request.status == 200) {
        this.addSuggestions(request.response);
      } else {
        this.suggestionsEl.classList.add("hidden");
        //this.dismissSuggestions();
        console.error("Error while loading item.");
      }
    });
    request.send();
  }

  addSuggestions(content: any) {
    this.suggestionsEl.innerHTML = content;
    this.suggestionsEl.classList.remove("hidden");

    this.suggestions = this.suggestionsEl.querySelectorAll(
      ".admin_search_suggestion"
    );

    for (var i = 0; i < this.suggestions.length; i++) {
      var suggestion = <HTMLAnchorElement>this.suggestions[i];
      suggestion.addEventListener("touchend", (e) => {
        var el = <HTMLDivElement>e.currentTarget;
        window.location.href = el.getAttribute("href");
      });
      suggestion.addEventListener("click", (e) => {
        //this.logClick();
        return false;
      });
      suggestion.addEventListener("mouseenter", (e) => {
        this.deselect();
        var el = <HTMLDivElement>e.currentTarget;
        this.setSelected(parseInt(el.getAttribute("data-position")));
      });
    }
  }

  deselect() {
    var el = this.suggestionsEl.querySelector(
      ".admin_search_suggestion-selected"
    );
    if (el) {
      el.classList.remove("admin_search_suggestion-selected");
    }
  }

  getSelected(): number {
    var el = this.suggestionsEl.querySelector(
      ".admin_search_suggestion-selected"
    );
    if (el) {
      return parseInt(el.getAttribute("data-position"));
    }
    return -1;
  }

  setSelected(position: number) {
    this.deselect();
    if (position >= 0) {
      var els = this.suggestionsEl.querySelectorAll(".admin_search_suggestion");
      els[position].classList.add("admin_search_suggestion-selected");
    }
  }
}
