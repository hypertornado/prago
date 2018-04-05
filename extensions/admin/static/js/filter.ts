function bindFilter() {
  var els = document.querySelectorAll(".admin_filter_layout_date");
  for (var i = 0; i < els.length; i++) {
    new FilterDate(<HTMLDivElement>els[i]);
  }
}

class FilterDate {

  hidden: HTMLInputElement;
  from: HTMLInputElement;
  to: HTMLInputElement;

  constructor(el: HTMLDivElement) {
    this.hidden = <HTMLInputElement>el.querySelector(".admin_table_filter_item");
    this.from = <HTMLInputElement>el.querySelector(".admin_filter_layout_date_from");
    this.to = <HTMLInputElement>el.querySelector(".admin_filter_layout_date_to");

    this.from.addEventListener("input", this.changed.bind(this));
    this.to.addEventListener("input", this.changed.bind(this));

  }

  changed() {
    var val = "";
    if (this.from.value && this.to.value) {
      val = this.from.value + " - " + this.to.value;
    }
    this.hidden.value = val;

    var event = new Event('change');
    this.hidden.dispatchEvent(event);
  }

}