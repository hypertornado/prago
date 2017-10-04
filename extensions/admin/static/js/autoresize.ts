class Autoresize {
  el: HTMLTextAreaElement;

  constructor(el: HTMLTextAreaElement) {
    //DISABLED
    return;
    /*
    this.el = el;

    this.el.addEventListener('change', this.resizeIt.bind(this));
    this.el.addEventListener('cut', this.delayedResize.bind(this));
    this.el.addEventListener('paste', this.delayedResize.bind(this));
    this.el.addEventListener('drop', this.delayedResize.bind(this));
    this.el.addEventListener('keydown', this.delayedResize.bind(this));

    this.resizeIt();*/
  }

  delayedResize () {
    var self = this;
    setTimeout(function () {self.resizeIt()}, 0);
  }

  resizeIt() {
    this.el.style.height = 'auto';
    this.el.style.height = this.el.scrollHeight+'px';
  }
}