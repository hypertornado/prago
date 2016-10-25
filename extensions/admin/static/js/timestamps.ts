function bindTimestamps() {
  function bindTimestamp(el: HTMLElement) {
    var hidden = el.getElementsByTagName("input")[0];
    var v = hidden.value;

    if (v == "0001-01-01 00:00") {
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
    }

    var date = v.split(" ")[0];
    var hour = parseInt(v.split(" ")[1].split(":")[0]);
    var minute = parseInt(v.split(" ")[1].split(":")[1]);

    var timestampEl = <HTMLInputElement>el.getElementsByClassName("admin_timestamp_date")[0];
    timestampEl.value = date;

    var hourEl = el.getElementsByClassName("admin_timestamp_hour")[0];
    for (var i = 0; i < 24; i++) {
      var newEl = document.createElement("option");
      var addVal = "" + i;
      if (i < 10) {
        addVal = "0" + addVal;
      }
      newEl.innerText = addVal;
      newEl.setAttribute("value", addVal);

      if (hour == i) {
        newEl.setAttribute("selected","selected");
      }
      hourEl.appendChild(newEl);
    }

    var minEl = el.getElementsByClassName("admin_timestamp_minute")[0];
    for (var i = 0; i < 60; i++) {
      var newEl = document.createElement("option");
      var addVal = "" + i;
      if (i < 10) {
        addVal = "0" + addVal;
      }
      newEl.innerText = addVal;
      newEl.setAttribute("value", addVal);


      if (minute == i) {
        newEl.setAttribute("selected","selected");
      }
      minEl.appendChild(newEl);
    }

    var elTsDate = <HTMLInputElement>el.getElementsByClassName("admin_timestamp_date")[0];
    var elTsHour = <HTMLInputElement>el.getElementsByClassName("admin_timestamp_hour")[0];
    var elTsMinute = <HTMLInputElement>el.getElementsByClassName("admin_timestamp_minute")[0];
    var elTsInput = <HTMLInputElement>el.getElementsByTagName("input")[0];

    function saveValue() {
      var str = elTsDate.value + " " + elTsHour.value + ":" + elTsMinute.value;
      elTsInput.value = str;
    }
    saveValue();

    elTsDate.addEventListener("change", saveValue);
    elTsHour.addEventListener("change", saveValue);
    elTsMinute.addEventListener("change", saveValue);
  }

  var elements = document.querySelectorAll(".admin_timestamp");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    bindTimestamp(el);
  });
}