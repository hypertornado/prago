class Timestamp {
  elTsInput: HTMLInputElement;
  elTsDate: HTMLInputElement;
  elTsHour: HTMLInputElement;
  elTsMinute: HTMLInputElement;

  constructor(el: HTMLDivElement) {
    this.elTsInput = <HTMLInputElement>el.getElementsByTagName("input")[0];
    this.elTsDate = <HTMLInputElement>(
      el.getElementsByClassName("admin_timestamp_date")[0]
    );
    this.elTsHour = <HTMLInputElement>(
      el.getElementsByClassName("admin_timestamp_hour")[0]
    );
    this.elTsMinute = <HTMLInputElement>(
      el.getElementsByClassName("admin_timestamp_minute")[0]
    );

    this.initClock();

    var v = this.elTsInput.value;

    /*if (v == "0001-01-01 00:00") {
      var d = new Date();
      var month = d.getMonth() + 1;
      var monthStr = String(month);
      if (month < 10) {
        monthStr = "0" + monthStr;
      }

      var day = d.getUTCDate();
      var dayStr = String(day);
      if (day < 10) {
        dayStr = "0" + dayStr;
      }

      v = d.getFullYear() + "-" + monthStr + "-" + dayStr + " " + d.getHours() + ":" + d.getMinutes();
    }*/

    this.setTimestamp(v);

    this.elTsDate.addEventListener("change", this.saveValue.bind(this));
    this.elTsHour.addEventListener("change", this.saveValue.bind(this));
    this.elTsMinute.addEventListener("change", this.saveValue.bind(this));

    this.saveValue();
  }

  setTimestamp(v: string) {
    if (v == "") {
      return;
    }
    var date = v.split(" ")[0];
    var hour = parseInt(v.split(" ")[1].split(":")[0]);
    var minute = parseInt(v.split(" ")[1].split(":")[1]);

    this.elTsDate.value = date;

    var minuteOption: HTMLOptionElement = <HTMLOptionElement>(
      this.elTsMinute.children[minute]
    );
    minuteOption.selected = true;

    var hourOption: HTMLOptionElement = <HTMLOptionElement>(
      this.elTsHour.children[hour]
    );
    hourOption.selected = true;
  }

  initClock() {
    for (var i = 0; i < 24; i++) {
      var newEl = document.createElement("option");
      var addVal = "" + i;
      if (i < 10) {
        addVal = "0" + addVal;
      }
      newEl.innerText = addVal;
      newEl.setAttribute("value", addVal);
      this.elTsHour.appendChild(newEl);
    }

    for (var i = 0; i < 60; i++) {
      var newEl = document.createElement("option");
      var addVal = "" + i;
      if (i < 10) {
        addVal = "0" + addVal;
      }
      newEl.innerText = addVal;
      newEl.setAttribute("value", addVal);
      this.elTsMinute.appendChild(newEl);
    }
  }

  saveValue() {
    var str =
      this.elTsDate.value +
      " " +
      this.elTsHour.value +
      ":" +
      this.elTsMinute.value;
    if (this.elTsDate.value == "") {
      str = "";
    }
    this.elTsInput.value = str;
  }
}
