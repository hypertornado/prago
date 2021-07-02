class ListMultiple {
  list: List;

  constructor(list: List) {
    this.list = list;
    if (this.hasMultipleActions()) {
      this.bindMultipleActions();
    }
  }

  pseudoCheckboxesAr: NodeListOf<HTMLTableCellElement>;
  lastCheckboxIndexClicked: number;

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

    //this.multipleActionStart("edit", ["2", "16"]);
  }

  multipleActionSelected(e: any) {
    var target: HTMLDivElement = e.target;
    var actionName = target.getAttribute("name");
    var ids = this.multipleGetIDs();
    this.multipleActionStart(actionName, ids);
  }

  multipleActionStart(actionName: string, ids: Array<String>) {
    switch (actionName) {
      case "cancel":
        this.multipleUncheckAll();
        break;
      case "edit":
        new ListMultipleEdit(this, ids);
        break;
      case "delete":
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
    this.lastCheckboxIndexClicked = -1;
    this.pseudoCheckboxesAr = document.querySelectorAll(
      ".admin_table_cell-multiple"
    );
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = <HTMLTableCellElement>this.pseudoCheckboxesAr[i];
      checkbox.addEventListener(
        "click",
        this.multipleCheckboxClicked.bind(this)
      );
    }
    this.multipleCheckboxChanged();
  }

  multipleGetIDs(): Array<String> {
    var ret: Array<String> = [];
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = <HTMLTableCellElement>this.pseudoCheckboxesAr[i];
      if (checkbox.classList.contains("admin_table_cell-multiple-checked")) {
        ret.push(checkbox.getAttribute("data-id"));
      }
    }
    return ret;
  }

  multipleCheckboxClicked(e: MouseEvent) {
    var cell: HTMLTableCellElement = <HTMLTableCellElement>e.currentTarget;
    var index: number = this.indexOfClickedCheckbox(cell);

    if (e.shiftKey && this.lastCheckboxIndexClicked >= 0) {
      var start = Math.min(index, this.lastCheckboxIndexClicked);
      var end = Math.max(index, this.lastCheckboxIndexClicked);
      for (var i = start; i <= end; i++) {
        this.checkPseudocheckbox(i);
      }
    } else {
      this.lastCheckboxIndexClicked = index;
      if (this.isCheckedPseudocheckbox(index)) {
        this.uncheckPseudocheckbox(index);
      } else {
        this.checkPseudocheckbox(index);
      }
    }

    e.preventDefault();
    e.stopPropagation();

    this.multipleCheckboxChanged();

    return false;
  }

  isCheckedPseudocheckbox(index: number): boolean {
    var sb: HTMLTableCellElement = this.pseudoCheckboxesAr[index];
    return sb.classList.contains("admin_table_cell-multiple-checked");
  }

  checkPseudocheckbox(index: number) {
    var sb: HTMLTableCellElement = this.pseudoCheckboxesAr[index];
    sb.classList.add("admin_table_cell-multiple-checked");
  }

  uncheckPseudocheckbox(index: number) {
    var sb: HTMLTableCellElement = this.pseudoCheckboxesAr[index];
    sb.classList.remove("admin_table_cell-multiple-checked");
  }

  multipleCheckboxChanged() {
    var checkedCount = 0;
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = <HTMLTableCellElement>this.pseudoCheckboxesAr[i];
      if (checkbox.classList.contains("admin_table_cell-multiple-checked")) {
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
    this.lastCheckboxIndexClicked = -1;
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = this.pseudoCheckboxesAr[i];
      checkbox.classList.remove("admin_table_cell-multiple-checked");
    }
    this.multipleCheckboxChanged();
  }

  indexOfClickedCheckbox(el: HTMLTableCellElement): number {
    var ret: number = -1;
    this.pseudoCheckboxesAr.forEach((v, k) => {
      if (v == el) {
        ret = k;
      }
    });
    return ret;
  }
}
