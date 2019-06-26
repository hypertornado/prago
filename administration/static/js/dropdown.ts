function bindDropdowns() {
  var els = document.querySelectorAll(".admin_dropdown");
  for (var i = 0; i < els.length; i++) {
    new Dropdown(<HTMLDivElement>els[i]);
  }
}

class Dropdown {
  targetEl: HTMLDivElement;
  contentEl: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.targetEl = <HTMLDivElement>el.querySelector(".admin_dropdown_target");
    this.contentEl = <HTMLDivElement>el.querySelector(".admin_dropdown_content");

    this.targetEl.addEventListener("mousedown", (e) => {
      if (document.activeElement == el) {
        el.blur();
        e.preventDefault();
        return false;
      }
    });

  }

}