class ListMultiple {
  list: List;

  constructor(list: List) {
    //console.log("list");
    this.list = list;

    if (this.hasMultipleActions()) {
      this.bindMultipleActions();
    }
  }

  checkboxesAr: NodeListOf<HTMLInputElement>;

  hasMultipleActions(): Boolean {
    if (this.list.el.classList.contains("admin_list-hasmultipleactions")) {
      return true;
    }
    return false;
  }

  bindMultipleActions() {
    var actions = this.list.el.querySelectorAll(".admin_list_multiple_action");
    for (var i = 0; i < actions.length; i++) {
      actions[i].addEventListener(
        "click",
        this.multipleActionSelected.bind(this)
      );
    }
  }

  multipleActionSelected(e: any) {
    var target: HTMLDivElement = e.target;
    var actionName = target.getAttribute("name");

    switch (actionName) {
      case "cancel":
        this.multipleUncheckAll();
        break;
      case "delete":
        var ids = this.multipleGetIDs();
        new Confirm(
          `Opravdu chcete smazat ${ids.length} položek?`,
          () => {
            var loader = new LoadingPopup();
            var params: any = {};
            params["action"] = "delete";
            params["ids"] = ids.join(",");
            var url =
              this.list.adminPrefix +
              "/" +
              this.list.typeName +
              "/api/multipleaction" +
              encodeParams(params);
            fetch(url, {
              method: "POST",
            }).then((e) => {
              loader.done();
              if (e.status != 200) {
                new Alert("Error while doing multipleaction delete");
                return;
              }
              this.list.load();
            });
          },
          Function(),
          ButtonStyle.Delete
        );
        break;
      default:
        console.log("other");
    }
  }

  bindMultipleActionCheckboxes() {
    this.checkboxesAr = document.querySelectorAll(
      ".admin_table_cell-multiple_checkbox"
    );
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      checkbox.addEventListener(
        "change",
        this.multipleCheckboxChanged.bind(this)
      );
    }
    this.multipleCheckboxChanged();
  }

  multipleGetIDs(): Array<String> {
    var ret: Array<String> = [];
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      if (checkbox.checked) {
        ret.push(checkbox.getAttribute("data-id"));
      }
    }
    return ret;
  }

  multipleCheckboxChanged() {
    var checkedCount = 0;
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      if (checkbox.checked) {
        checkedCount++;
      }
    }

    var multipleActionsPanel: HTMLDivElement = this.list.el.querySelector(
      ".admin_list_multiple_actions"
    );
    if (checkedCount > 0) {
      multipleActionsPanel.classList.add("admin_list_multiple_actions-visible");
    } else {
      multipleActionsPanel.classList.remove(
        "admin_list_multiple_actions-visible"
      );
    }
    this.list.el.querySelector(
      ".admin_list_multiple_actions_description"
    ).textContent = `Vybráno ${checkedCount} položek`;
  }

  multipleUncheckAll() {
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      checkbox.checked = false;
    }
    this.multipleCheckboxChanged();
  }
}
