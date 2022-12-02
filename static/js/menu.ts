class Menu {
  rootEl: HTMLDivElement;
  rootLeft: HTMLDivElement;
  menuEl: HTMLDivElement;

  search: SearchForm;

  constructor() {
    this.rootEl = document.querySelector(".root");
    this.rootLeft = document.querySelector(".root_left");

    this.menuEl = document.querySelector(".root_hamburger");
    this.menuEl.addEventListener("click", this.menuClick.bind(this));

    var searchFormEl = document.querySelector<HTMLFormElement>(".searchbox");
    if (searchFormEl) {
      this.search = new SearchForm(searchFormEl);
    }

    this.scrollTo(this.loadFromStorage());
    this.rootLeft.addEventListener("scroll", this.scrollHandler.bind(this));
  }

  scrollHandler() {
    this.saveToStorage(this.rootLeft.scrollTop);
  }

  saveToStorage(position: number) {
    window.localStorage["left_menu_position"] = position;
  }

  menuClick() {
    console.log("toggle");
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
}
