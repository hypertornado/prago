class Menu {
  rootEl: HTMLDivElement;
  rootLeft: HTMLDivElement;
  hamburgerMenuEl: HTMLDivElement;

  search: SearchForm;

  constructor() {
    this.rootEl = document.querySelector(".root");
    this.rootLeft = document.querySelector(".root_left");

    this.hamburgerMenuEl = document.querySelector(".root_hamburger");
    this.hamburgerMenuEl.addEventListener("click", this.menuClick.bind(this));

    var searchFormEl = document.querySelector<HTMLFormElement>(".searchbox");
    if (searchFormEl) {
      this.search = new SearchForm(searchFormEl);
    }

    this.scrollTo(this.loadFromStorage());
    this.rootLeft.addEventListener("scroll", this.scrollHandler.bind(this));

    this.bindSubmenus();
    this.bindResourceCounts();
  }

  scrollHandler() {
    this.saveToStorage(this.rootLeft.scrollTop);
  }

  saveToStorage(position: number) {
    window.localStorage["left_menu_position"] = position;
  }

  menuClick() {
    this.rootEl.classList.toggle("root-visible");
  }

  loadFromStorage(): number {
    var pos = window.localStorage["left_menu_position"];
    if (pos) {
      return parseInt(pos);
    }
    return 0;
  }

  scrollTo(position: number) {
    this.rootLeft.scrollTo(0, position);
  }

  bindSubmenus() {
    let triangleIcons = document.querySelectorAll(".menu_row_icon");

    for (var i = 0; i < triangleIcons.length; i++) {
      let triangleIcon = <HTMLDivElement>triangleIcons[i];
      triangleIcon.addEventListener("click", () => {
        let parent = <HTMLDivElement>triangleIcon.parentElement;
        parent.classList.toggle("menu_row-expanded");
      });
    }
  }

  bindResourceCounts() {
    this.setResourceCountsFromCache();
    new VisibilityReloader(2000, () => {
      this.loadResourceCounts();
    });
  }

  saveCountToStorage(url: string, count: string) {
    if (!window.localStorage) {
      return;
    }
    window.localStorage["left_menu_count-" + url] = count;
  }

  loadCountFromStorage(url: string): string {
    if (!window.localStorage) {
      return "";
    }
    var pos = window.localStorage["left_menu_count-" + url];
    if (pos) {
      return pos;
    }
    return "";
  }

  setResourceCountsFromCache() {
    var items = document.querySelectorAll(".menu_item");
    for (var i = 0; i < items.length; i++) {
      let item = <HTMLDivElement>items[i];
      let url = item.getAttribute("href");
      let count = this.loadCountFromStorage(url);
      if (count) {
        this.setResourceCount(item, count);
      }
    }
  }

  setResourceCounts(data: any) {
    var items = document.querySelectorAll(".menu_item");
    for (var i = 0; i < items.length; i++) {
      let item = <HTMLDivElement>items[i];
      let url = item.getAttribute("href");
      let count = data[url];
      this.setResourceCount(item, count);
    }
  }

  setResourceCount(el: HTMLDivElement, count: string) {
    let countEl = el.querySelector(".menu_item_right");
    if (count) {
      this.saveCountToStorage(el.getAttribute("href"), count);
      countEl.textContent = count;
    }
  }

  loadResourceCounts() {
    var request = new XMLHttpRequest();

    request.open("GET", "/admin/api/resource-counts", true);

    request.addEventListener("load", () => {
      if (request.status == 200) {
        var data = JSON.parse(request.response);
        this.setResourceCounts(data);
      } else {
        console.error("cant load resource counts");
      }
    });
    request.send();
  }
}
