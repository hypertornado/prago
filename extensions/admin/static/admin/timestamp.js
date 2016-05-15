function bindTimestamps() {
  function bindTimestamp(el) {
    var hidden = el.find("input").first();
    var v = hidden.val();

    if (v == "0001-01-01 00:00") {
      var d = new Date();
      var month = d.getMonth() + 1;
      if (month < 10) {
        month = "0" + month;
      }

      var day = d.getUTCDate();
      if (day < 10) {
        day = "0" + day;
      }

      v = d.getFullYear() + "-" + month + "-" + day + " " + d.getHours() + ":" + d.getMinutes();
    }

    var date = v.split(" ")[0];
    var hour = parseInt(v.split(" ")[1].split(":")[0]);
    var minute = parseInt(v.split(" ")[1].split(":")[1]);

    el.find(".admin_timestamp_date").val(date);

    var hourEl = el.find(".admin_timestamp_hour");
    for (var i = 0; i < 24; i++) {
      var newEl = $("<option></option>");
      var addVal = "" + i;
      if (i < 10) {
        addVal = "0" + addVal;
      }
      newEl.text(addVal);
      newEl.attr("value", addVal);

      if (hour == i) {
        newEl.attr("selected","selected");
      }
      hourEl.append(newEl);
    }

    var minEl = el.find(".admin_timestamp_minute");
    for (var i = 0; i < 60; i++) {
      var newEl = $("<option></option>");
      var addVal = "" + i;
      if (i < 10) {
        addVal = "0" + addVal;
      }
      newEl.text(addVal);
      newEl.attr("value", addVal);


      if (minute == i) {
        newEl.attr("selected","selected");
      }
      minEl.append(newEl);
    }

    function saveValue() {
      var date = el.find(".admin_timestamp_date").val();
      var hour = el.find(".admin_timestamp_hour").val();
      var minute = el.find(".admin_timestamp_minute").val();
      var str = date + " " + hour + ":" + minute;
      el.find("input").first().val(str);
    }
    saveValue();

    el.find(".admin_timestamp_date").on("change", saveValue);
    el.find(".admin_timestamp_hour").on("change", saveValue);
    el.find(".admin_timestamp_minute").on("change", saveValue);
  }

  $(".admin_timestamp").each(
    function() {
      bindTimestamp($(this));
    }
  );
}