class ListSettings {
  list: List;

  settingsEl: HTMLDivElement;
  settingsPopup: ContentPopup;

  statsEl: HTMLDivElement;
  statsPopup: ContentPopup;

  statsCheckboxSelectCount: HTMLSelectElement;

  statsContainer: HTMLDivElement;

  constructor(list: List) {
    this.list = list;

    this.settingsEl = document.querySelector(".list_settings");
    this.settingsPopup = new ContentPopup("Nastavení", this.settingsEl);
    this.settingsPopup.setIcon("glyphicons-basic-137-cogwheel.svg");
    this.statsContainer = document.querySelector(".list_stats_container");
    
    this.statsEl = document.querySelector(".list_stats");
    this.statsPopup = new ContentPopup("Statistiky", this.statsEl);
    this.statsPopup.setIcon("glyphicons-basic-43-stats-circle.svg");

    this.statsCheckboxSelectCount = document.querySelector(".list_stats_limit");
    this.statsCheckboxSelectCount.addEventListener("change", () => {
      this.loadStats();
    });
  }

  bindSettingsBtn(btn: HTMLButtonElement) {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      cmenu({
        Event: e,
        AlignByElement: true,
        Commands: [
          {
            Name: "Nastavení",
            Icon: "glyphicons-basic-137-cogwheel.svg",
            Handler: () => {
              this.settingsPopup.show();
            },
          },
          {
            Name: "Statistiky",
            Icon: "glyphicons-basic-43-stats-circle.svg",
            Handler: () => {
              this.loadStats();
              this.statsPopup.show();
            },
          },
          {
            Name: "Export CSV",
            Icon: "glyphicons-basic-302-square-download.svg",
            Handler: () => {
              window.open("/admin/" + this.list.typeName +"/api/export.csv")
            }
          },
          {
            Name: "Kontrola konzistence",
            Icon: "glyphicons-basic-322-shield-check.svg",
            Handler: () => {
              new PopupForm("/admin/_validation-consistency?resource=" + this.list.typeName, (data: any) => {
                //this.addUUID(data.Data);
              })
            }
          },
        ],
      })
    });
  }

  loadStats() {
    let filterData = this.list.getFilterData();

    var params: any = {};

    params["_statslimit"] = this.statsCheckboxSelectCount.value;

    for (var k in filterData) {
      params[k] = filterData[k];
    }

    var request = new XMLHttpRequest();

    var encoded = encodeParams(params);

    request.open(
      "GET",
      "/admin/" + this.list.typeName + "/api/list-stats" + encoded,
      true
    );

    this.statsContainer.innerHTML = "Loading...";

    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.statsContainer.innerHTML = request.response;
      }
    })
    request.send();

  }

  bindOptions(visibleColumnsMap: any) {
    var columns: NodeListOf<HTMLInputElement> = document.querySelectorAll(
      ".list_settings_column"
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
