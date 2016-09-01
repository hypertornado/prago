function bindRelations() {
  function bindRelation(el) {
    var relationName = el.find("input").data("relation");
    var originalValue = el.find("input").val();

    var select = $("<select></select>");
    select.addClass("input form_input");

    select.on("change", function() {
      el.find("input").val(select.val());
    })

    var adminPrefix = $("body").data("admin-prefix");

    $.ajax({
      "url": adminPrefix + "/_api/resource/" + relationName,
      "success": function (responseData) {
        addOption(select, "0", "", false);

        responseData.forEach(function (i) {
          var selected = false;
          if (originalValue == i["id"]) {
            selected = true;
          }

          addOption(select, i["id"], i["name"], selected);
        })

        el.append(select);
        el.find("progress").hide();
      },
      error: function() {
        el.find("progress").hide();
        el.append("Error wile loading relation " + relationName + ".");
      }
    });
  }

  function addOption(select, value, description, selected) {
    var option = $("<option></option>");
    if (selected) {
      option.attr("selected", "selected");
    }
    option.attr("value", value);
    option.text(description);
    select.append(option);
  }

  $(".admin_item_relation").each(
    function() {
      bindRelation($(this));
    }
  );

}