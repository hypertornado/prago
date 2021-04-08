class Autoresize {
  el: HTMLTextAreaElement;

  constructor(el: HTMLTextAreaElement) {
    this.el = el;
    this.el.addEventListener("input", this.resizeIt.bind(this));
    this.resizeIt();
  }

  resizeIt() {
    var height = this.el.scrollHeight + 2;
    this.el.style.height = height + "px";
  }
}
