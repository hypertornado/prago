class ListSettings {
  list: List;

  settingsRow: HTMLTableRowElement;
  settingsRowColumn: HTMLTableElement;
  settingsEl: HTMLDivElement;
  settingsCheckbox: HTMLInputElement;

  settingsButton: HTMLButtonElement;
  settingsPopup: ContentPopup;

  constructor(list: List) {
    this.list = list;

    this.settingsRow = document.querySelector(".admin_list_settingsrow");
    this.settingsRowColumn = document.querySelector(
      ".admin_list_settingsrow_column"
    );
    this.settingsEl = document.querySelector(".admin_tablesettings");

    this.settingsPopup = new ContentPopup("Možnosti", this.settingsEl);
    this.settingsButton = document.querySelector(".admin_list_settings");
    this.settingsButton.addEventListener("click", () => {
      this.settingsPopup.show();
    });
  }

  settingsCheckboxChange() {
    if (this.settingsCheckbox.checked) {
      this.settingsRow.classList.add("admin_list_settingsrow-visible");
    } else {
      this.settingsRow.classList.remove("admin_list_settingsrow-visible");
    }
  }

  bindOptions(visibleColumnsMap: any) {
    var columns: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".admin_tablesettings_column"
    );
    for (var i = 0; i < columns.length; i++) {
      let columnName = columns[i].getAttribute("data-column-name");
      if (visibleColumnsMap[columnName]) {
        columns[i].checked = true;
      }
      columns[i].addEventListener("change", () => {
        this.changedOptions();
      });
    }
    this.changedOptions();
  }

  changedOptions() {
    var columns: any = this.getSelectedColumnsMap();

    var headers: NodeListOf<HTMLDivElement> =
      document.querySelectorAll(".list_header_item");
    for (var i = 0; i < headers.length; i++) {
      var name = headers[i].getAttribute("data-name");
      if (columns[name]) {
        headers[i].classList.remove("hidden");
      } else {
        headers[i].classList.add("hidden");
      }
    }

    var filters: NodeListOf<HTMLDivElement> = document.querySelectorAll(
      ".list_header_item_filter"
    );
    for (var i = 0; i < filters.length; i++) {
      var name = filters[i].getAttribute("data-name");
      if (columns[name]) {
        filters[i].classList.remove("hidden");
      } else {
        filters[i].classList.add("hidden");
      }
    }

    this.settingsRowColumn.setAttribute(
      "colspan",
      Object.keys(columns).length + ""
    );

    this.list.load();
  }

  getSelectedColumnsStr(): string {
    var ret = [];
    var checked: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".admin_tablesettings_column:checked"
    );
    for (var i = 0; i < checked.length; i++) {
      ret.push(checked[i].getAttribute("data-column-name"));
    }
    return ret.join(",");
  }

  getSelectedColumnsMap(): any {
    var columns: any = {};
    var checked: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".admin_tablesettings_column:checked"
    );
    for (var i = 0; i < checked.length; i++) {
      columns[checked[i].getAttribute("data-column-name")] = true;
    }
    return columns;
  }
}
