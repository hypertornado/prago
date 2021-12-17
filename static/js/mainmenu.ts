class MainMenu {
  leftEl: HTMLDivElement;
  menuEl: HTMLDivElement;

  search: SearchForm;

  constructor(leftEl: HTMLDivElement) {
    this.leftEl = leftEl;
    this.menuEl = document.querySelector(".admin_mobile_menu");
    this.menuEl.addEventListener("click", this.menuClick.bind(this));

    var searchFormEl = leftEl.querySelector<HTMLFormElement>(
      ".admin_header_search"
    );
    if (searchFormEl) {
      this.search = new SearchForm(searchFormEl);
    }

    this.scrollTo(this.loadFromStorage());
    this.leftEl.addEventListener("scroll", this.scrollHandler.bind(this));
  }

  scrollHandler() {
    this.saveToStorage(this.leftEl.scrollTop);
  }

  saveToStorage(position: number) {
    window.localStorage["left_menu_position"] = position;
  }

  menuClick() {
    this.leftEl.classList.toggle("admin_layout_left-visible");
    this.menuEl.classList.toggle("admin_mobile_menu-selected");
  }

  loadFromStorage(): number {
    var pos = window.localStorage["left_menu_position"];
    if (pos) {
      return parseInt(pos);
    }
    return 0;
  }

  scrollTo(position: number) {
    this.leftEl.scrollTo(0, position);
  }
}
