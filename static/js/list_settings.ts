class ListSettings {
  list: List;

  settingsEl: HTMLDivElement;
  settingsButton: HTMLButtonElement;
  settingsPopup: ContentPopup;

  statsEl: HTMLDivElement;
  statsButton: HTMLButtonElement;
  statsPopup: ContentPopup;

  exportEl: HTMLDivElement;
  exportButton: HTMLButtonElement;
  exportPopup: ContentPopup;

  constructor(list: List) {
    this.list = list;

    this.settingsEl = document.querySelector(".list_settings");
    this.settingsPopup = new ContentPopup("Možnosti", this.settingsEl);
    this.settingsButton = document.querySelector(
      ".list_header_action-settings"
    );
    this.settingsButton.addEventListener("click", () => {
      this.settingsPopup.show();
    });

    this.statsEl = document.querySelector(".list_stats");
    this.statsPopup = new ContentPopup("Statistiky", this.statsEl);
    this.statsPopup.setHiddenHandler(() => {
      this.list.loadStats = false;
    });
    this.statsButton = document.querySelector(".list_header_action-stats");
    this.statsButton.addEventListener("click", () => {
      this.list.loadStats = true;
      this.list.load();
      this.statsPopup.show();
    });

    this.exportEl = document.querySelector(".list_export");
    this.exportPopup = new ContentPopup("Export", this.exportEl);
    this.exportButton = document.querySelector(".list_header_action-export");
    this.exportButton.addEventListener("click", () => {
      this.exportPopup.show();
    });
  }

  bindOptions(visibleColumnsMap: any) {
    var columns: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".list_settings_column"
    );
    for (var i = 0; i < columns.length; i++) {
      let columnName = columns[i].getAttribute("data-column-name");
      console.log(columnName);
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
      if (columns[name] === true) {
        filters[i].classList.remove("hidden");
      }
      if (columns[name] === false) {
        //filters[i].classList.add("hidden");
      }
    }

    this.list.load();
  }

  getSelectedColumnsStr(): string {
    var ret = [];
    var checked: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".list_settings_column:checked"
    );
    for (var i = 0; i < checked.length; i++) {
      ret.push(checked[i].getAttribute("data-column-name"));
    }
    return ret.join(",");
  }

  getSelectedColumnsMap(): any {
    var columns: any = {};
    var inputs: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".list_settings_column"
    );
    for (var i = 0; i < inputs.length; i++) {
      if (inputs[i].checked) {
        columns[inputs[i].getAttribute("data-column-name")] = true;
      } else {
        columns[inputs[i].getAttribute("data-column-name")] = false;
      }
    }
    return columns;
  }
}
