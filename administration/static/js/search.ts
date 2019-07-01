function bindSearch() {
  var els = document.querySelectorAll(".admin_header_search");
  for (var i = 0; i < els.length; i++) {
    new SearchForm(<HTMLDivElement>els[i]);
  }
}

class SearchForm {
  searchInput: HTMLInputElement;
  suggestionsEl: HTMLDivElement;
  suggestions: any;
  dirty: boolean;
  lastChanged: number;

  constructor(el: HTMLDivElement) {
    this.searchInput = <HTMLInputElement>el.querySelector(".admin_header_search_input");
    this.suggestionsEl = <HTMLDivElement>el.querySelector(".admin_header_search_suggestions");

    this.searchInput.addEventListener("input", () => {
      this.suggestions = [];
      this.dirty = true;
      this.lastChanged = Date.now();
      return false;
    });

    this.searchInput.addEventListener("blur", () => {
      //this.suggestionsEl.classList.add("hidden");
    });

    window.setInterval(() => {
      if (this.dirty && Date.now() - this.lastChanged > 100) {
        this.loadSuggestions();
      }
    }, 30);
  }

  loadSuggestions() {
    this.dirty = false;
    var suggestText = this.searchInput.value;
    var request = new XMLHttpRequest();
    var url = "/admin/_search_suggest" + encodeParams({"q": this.searchInput.value});
    request.open("GET", url);
    request.addEventListener("load", () => {
      if (suggestText != this.searchInput.value) {
        return;
      }
      console.log(request.response);
      if (request.status == 200) {
        this.addSuggestions(request.response);
      } else {
        this.suggestionsEl.classList.add("hidden");
        //this.dismissSuggestions();
        console.error("Error while loading item.");
      }
    })
    request.send();
  }

  addSuggestions(content: any) {
    console.log(content);
    this.suggestionsEl.innerHTML = content;
    this.suggestionsEl.classList.remove("hidden");
    /*
    this.suggestions = this.searchSuggestions.querySelectorAll(".head_search_suggestion");
    if (this.suggestions.length > 0) {
      this.searchForm.classList.add("head_search-suggestion");
    } else {
      this.searchForm.classList.remove("head_search-suggestion");
    }

    for (var i = 0; i < this.suggestions.length; i++) {
      var suggestion = <HTMLAnchorElement>this.suggestions[i];
      suggestion.addEventListener("touchend", (e) => {
        var el = <HTMLDivElement>e.currentTarget;
        window.location.href = el.getAttribute("href");
      });
      suggestion.addEventListener("click", (e) => {
        this.logClick();
        return false;
      });
      suggestion.addEventListener("mouseenter", (e) => {
        this.deselect();
        var el = <HTMLDivElement>e.currentTarget;
        this.setSelected(parseInt(el.getAttribute("data-position")));
      })
    }*/
  }

}

class SearchLAZNE {
  searchForm: HTMLFormElement;
  searchInput: HTMLInputElement;
  searchSuggestions: HTMLDivElement;
  suggestions: any;
  dirty: boolean;
  lastChanged: number;

  constructor() {
    this.searchForm = <HTMLFormElement>document.querySelector("form.head_search");
    this.searchInput = <HTMLInputElement>document.querySelector(".head_search_input");
    if (!this.searchInput) {
      return;
    }

    this.searchForm.addEventListener("submit", (e) => {
      if (this.searchInput.value == "") {
        this.searchInput.focus();
        e.preventDefault();
        return false;
      }
    });

    this.searchInput.addEventListener("input", (e) => {
      this.suggestions = [];
      this.dirty = true;
      this.lastChanged = Date.now();
      return false;
    });

    this.searchInput.addEventListener("focus", () => {
      this.searchForm.classList.add("head_search-focused");
    });

    /*this.searchInput.addEventListener("blur", () => {
      this.searchForm.classList.remove("head_search-focused");
    });*/

    this.searchSuggestions = <HTMLDivElement>document.querySelector(".head_search_suggestions");


    window.setInterval(() => {
      if (this.dirty && Date.now() - this.lastChanged > 100) {
        this.loadSuggestions();
      }
    }, 30);

    this.searchInput.addEventListener("keydown", (e) => {
      switch (e.keyCode) {
        case 13: //enter
          var i = this.getSelected();
          if (i >= 0) {
            var child = this.suggestions[i];
            if (child) {
              this.logClick();
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
            i += 1
            i = i % this.suggestions.length;
          } else {
            i = 0;
          }
          this.setSelected(i)
          e.preventDefault();
          return false;
      }
    })
  }

  logClick() {
    var selected = this.getSelected();
    if (selected >= 0) {
      var suggestion = <HTMLDivElement>this.suggestions[selected];
      var text = this.searchInput.value + " – " + suggestion.getAttribute("data-position") + " - " + suggestion.getAttribute("data-name");
      //ga('send', 'event', "Našeptávač vybrán", window.location.href, text);
    }
  }

  loadSuggestions() {
    this.dirty = false;
    var suggestText = this.searchInput.value;
    var request = new XMLHttpRequest();
    var url = "/suggest" + encodeParams({"q": this.searchInput.value});
    request.open("GET", url);
    request.addEventListener("load", () => {
      if (suggestText != this.searchInput.value) {
        return;
      }
      if (request.status == 200) {
        this.addSuggestions(request.response);
      } else {
        this.dismissSuggestions();
        console.error("Error while loading item.");
      }
    })
    request.send();
  }

  dismissSuggestions() {
    this.searchForm.classList.remove("head_search-suggestion");
    this.searchSuggestions.innerHTML = "";
  }

  addSuggestions(content: any) {
    this.searchSuggestions.innerHTML = content;
    this.suggestions = this.searchSuggestions.querySelectorAll(".head_search_suggestion");
    if (this.suggestions.length > 0) {
      this.searchForm.classList.add("head_search-suggestion");
    } else {
      this.searchForm.classList.remove("head_search-suggestion");
    }

    for (var i = 0; i < this.suggestions.length; i++) {
      var suggestion = <HTMLAnchorElement>this.suggestions[i];
      suggestion.addEventListener("touchend", (e) => {
        var el = <HTMLDivElement>e.currentTarget;
        window.location.href = el.getAttribute("href");
      });
      suggestion.addEventListener("click", (e) => {
        this.logClick();
        return false;
      });
      suggestion.addEventListener("mouseenter", (e) => {
        this.deselect();
        var el = <HTMLDivElement>e.currentTarget;
        this.setSelected(parseInt(el.getAttribute("data-position")));
      })
    }
  }

  deselect() {
    var el = this.searchSuggestions.querySelector(".head_search_suggestion-selected");
    if (el) {
      el.classList.remove("head_search_suggestion-selected");
    }
  }

  getSelected(): number {
    var el = this.searchSuggestions.querySelector(".head_search_suggestion-selected");
    if (el) {
      return parseInt(el.getAttribute("data-position"));
    }
    return -1;
  }

  setSelected(position: number) {
    this.deselect();
    if (position >= 0) {
       var els = this.searchSuggestions.querySelectorAll(".head_search_suggestion");
       els[position].classList.add("head_search_suggestion-selected");
    }
  }
}

function encodeParams(data: any) {
  var ret = "";
  for (var k in data) {
    if (!data[k]) {
      continue;
    }
    if (ret != "") {
      ret += "&";
    }
    ret += encodeURIComponent(k) + "=" + encodeURIComponent(data[k]);
  }
  if (ret != "") {
    ret = "?" + ret;
  }
  return ret;
}
