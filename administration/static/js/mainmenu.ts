function bindMainMenu() {
    var el: HTMLDivElement = document.querySelector(".admin_layout_left");
    if (el) {
        new MainMenu(el);
    }
}


class MainMenu {

    leftEl: HTMLDivElement;
    menuEl: HTMLDivElement;

    constructor(leftEl: HTMLDivElement) {
        this.leftEl = leftEl;
        this.menuEl = document.querySelector(".admin_header_container_menu");
        this.menuEl.addEventListener("click", this.menuClick.bind(this));

        this.scrollTo(this.loadFromStorage());
        this.leftEl.addEventListener("scroll", this.scrollHandler.bind(this))
    }

    scrollHandler() {
        this.saveToStorage(this.leftEl.scrollTop);
    }

    saveToStorage(position: number) {
        window.localStorage["left_menu_position"] = position;
    }

    menuClick() {
        this.leftEl.classList.toggle("admin_layout_left-visible");
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