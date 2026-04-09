class ListMultiple {
  list: List;
  lastWasUnchecked: boolean;


  constructor(list: List) {
    this.list = list;
    if (this.hasMultipleActions()) {
      this.bindMultipleActions();
    }
  }

  pseudoCheckboxesAr: NodeListOf<HTMLTableCellElement>;
  lastCheckboxIndexClicked: number;

  listHeaderAllSelect: HTMLDivElement;

  hasMultipleActions(): Boolean {
    if (this.list.listEl.classList.contains("list-hasmultipleactions")) {
      return true;
    }
    return false;
  }

  bindMultipleActions() {
    this.listHeaderAllSelect = this.list.listEl.querySelector(".list_header_multiple");
    this.listHeaderAllSelect.addEventListener("click", () =>{
      if (this.isAllChecked()) {
        this.multipleUncheckAll();
      } else {
        this.multipleCheckAll();
      }
    })

    var actions = this.list.listEl.querySelectorAll(".list_multiple_action");
    for (var i = 0; i < actions.length; i++) {
      actions[i].addEventListener(
        "click",
        this.multipleActionClicked.bind(this)
      );
    }

    this.list.listEl
      .querySelector(".list_multiple_actions_cancel")
      .addEventListener("click", () => {
        this.multipleUncheckAll();
      });
  }

  multipleActionClicked(e: any) {
    let btn = e.currentTarget;
    let actionID = btn.getAttribute("data-id");
    let resourceID = btn.getAttribute("data-resource-id");
    this.multipleActionForm(resourceID, actionID);
  }

  multipleActionForm(resourceID: string, actionID: string) {
    var ids = this.multipleGetIDs();
    let idsStr = ids.join(",");
    let formURL = `/admin/${resourceID}/${idsStr}/${actionID}`;
    //@ts-ignore
    new PopupForm(formURL, (data: any) => {
        if (data.RedirectionLocation) {
          window.location.href = data.RedirectionLocation;
        }
        this.list.load();
    });
  }

  bindMultipleActionCheckboxes() {
    this.lastCheckboxIndexClicked = -1;
    this.pseudoCheckboxesAr = document.querySelectorAll(".list_row_multiple");
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = <HTMLTableCellElement>this.pseudoCheckboxesAr[i];
      checkbox.addEventListener(
        "mousedown",
        this.multipleCheckboxMousedown.bind(this)
      );
      checkbox.addEventListener(
        "mouseenter",
        this.multipleCheckboxMousenter.bind(this)
      );
      checkbox.addEventListener(
        "click",
        (e: MouseEvent) => {
          e.preventDefault();
          e.stopPropagation();
        }
      );
    }
    this.multipleCheckboxChanged();
  }

  multipleGetIDs(): Array<String> {
    var ret: Array<String> = [];
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = <HTMLTableCellElement>this.pseudoCheckboxesAr[i];
      if (checkbox.classList.contains("list_row_multiple-checked")) {
        ret.push(checkbox.getAttribute("data-id"));
      }
    }
    return ret;
  }

  multipleCheckboxMousenter(e: MouseEvent) {
    var cell: HTMLTableCellElement = <HTMLTableCellElement>e.currentTarget;
    var index: number = this.indexOfClickedCheckbox(cell);
    if (e.buttons == 1) {
      if (this.lastWasUnchecked) {
        this.uncheckPseudocheckbox(index);
      } else {
        this.checkPseudocheckbox(index);
      }
      this.multipleCheckboxChanged();
    }
  }

  multipleCheckboxMousedown(e: MouseEvent) {
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
    return sb.classList.contains("list_row_multiple-checked");
  }

  checkPseudocheckbox(index: number) {
    var sb: HTMLTableCellElement = this.pseudoCheckboxesAr[index];
    sb.classList.add("list_row_multiple-checked");
    this.lastWasUnchecked = false;
  }

  uncheckPseudocheckbox(index: number) {
    var sb: HTMLTableCellElement = this.pseudoCheckboxesAr[index];
    sb.classList.remove("list_row_multiple-checked");
    this.lastWasUnchecked = true;
  }

  multipleCheckboxChanged() {
    var checkedCount = 0;
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = <HTMLTableCellElement>this.pseudoCheckboxesAr[i];
      if (checkbox.classList.contains("list_row_multiple-checked")) {
        checkedCount++;
      }
    }

    var multipleActionsPanel: HTMLDivElement = this.list.listEl.querySelector(
      ".list_multiple_actions"
    );
    if (checkedCount > 0) {
      multipleActionsPanel.classList.add("list_multiple_actions-visible");
    } else {
      multipleActionsPanel.classList.remove("list_multiple_actions-visible");
    }

    if (this.isAllChecked()) {
      this.listHeaderAllSelect.classList.add("list_row_multiple-checked");
    } else {
      this.listHeaderAllSelect.classList.remove("list_row_multiple-checked");
    }

    this.list.listEl.querySelector(
      ".list_multiple_actions_description"
    ).textContent = `Vybráno ${checkedCount} položek`;
  }

  multipleUncheckAll() {
    this.lastCheckboxIndexClicked = -1;
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = this.pseudoCheckboxesAr[i];
      checkbox.classList.remove("list_row_multiple-checked");
    }
    this.multipleCheckboxChanged();
  }

  multipleCheckAll() {
    this.lastCheckboxIndexClicked = -1;
    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = this.pseudoCheckboxesAr[i];
      checkbox.classList.add("list_row_multiple-checked");
    }
    this.multipleCheckboxChanged();
  }

  isAllChecked() {
    if (this.pseudoCheckboxesAr.length == 0) {
      return false;
    }

    for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
      var checkbox = this.pseudoCheckboxesAr[i];
      if (!checkbox.classList.contains("list_row_multiple-checked")) {
        return false
      }
    }
    return true;
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
