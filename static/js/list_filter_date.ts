class ListFilterDate {
  hidden: HTMLInputElement;
  from: HTMLInputElement;
  to: HTMLInputElement;

  constructor(el: HTMLDivElement, value: any) {
    this.hidden = <HTMLInputElement>(
      el.querySelector(".admin_table_filter_item")
    );
    this.from = <HTMLInputElement>(
      el.querySelector(".admin_filter_layout_date_from")
    );
    this.to = <HTMLInputElement>(
      el.querySelector(".admin_filter_layout_date_to")
    );

    this.from.addEventListener("input", this.changed.bind(this));
    this.from.addEventListener("change", this.changed.bind(this));
    this.to.addEventListener("input", this.changed.bind(this));
    this.to.addEventListener("change", this.changed.bind(this));

    this.setValue(value);
  }

  setValue(value: any) {
    if (!value) {
      return;
    }
    var splited = value.split(",");
    if (splited.length == 2) {
      this.from.value = splited[0];
      this.to.value = splited[1];
    }
    this.hidden.value = value;
  }

  changed() {
    var val = "";
    if (this.from.value || this.to.value) {
      val = this.from.value + "," + this.to.value;
    }
    this.hidden.value = val;

    var event = new Event("change");
    this.hidden.dispatchEvent(event);
  }
}
