class ListSettings {
  list: List;

  constructor(list: List) {
    this.list = list;
  }

  bindSettingsBtn(btn: HTMLButtonElement) {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      cmenu({
        Event: e,
        AlignByElement: true,
        Commands: [
          {
            Name: "Počet položek na stránce",
            Icon: "glyphicons-basic-960-files-queue.svg",
            Handler: () => {
              new PopupForm("/admin/_list-items-per-page?count=" + this.list.itemsPerPage +"&resource=" + this.list.typeName, (data: any) => {
                this.list.itemsPerPage = data.Data;
                this.list.load();
              });
            },
          },
          {
            Name: "Viditelné sloupce",
            Icon: "glyphicons-basic-107-text-width.svg",
            Handler: () => {
              new PopupForm("/admin/_list-items-visible?fields=" + this.list.visibleColumnsStr +"&resource=" + this.list.typeName, (data: any) => {
                this.list.visibleColumnsStr = data.Data;
                this.setVisibleColumns();
              });
            },
          },
          {
            Name: "Statistiky",
            Icon: "glyphicons-basic-43-stats-circle.svg",
            Handler: () => {
              var params: any = {};
              params["_resource"] = this.list.typeName;
              let filterData = this.list.filter.getFilterData();
              for (var k in filterData) {
                params[k] = filterData[k];
              }
              new PopupForm("/admin/_list-stats?_resource=" + this.list.typeName + "&_params=" + encodeURIComponent(JSON.stringify(params)), (data: any) => {
              });
            },
          },
          {
            Name: "Export CSV",
            Icon: "glyphicons-basic-302-square-download.svg",
            Handler: () => {
              var params: any = {};
              params["_resource"] = this.list.typeName;
              params["_columns"] = this.list.visibleColumnsStr;

              let filterData = this.list.filter.getFilterData();
              for (var k in filterData) {
                params[k] = filterData[k];
              }

              if (this.list.orderColumn != this.list.defaultOrderColumn) {
                params["_order"] = this.list.orderColumn;
              }
              if (this.list.orderDesc != this.list.defaultOrderDesc) {
                params["_desc"] = this.list.orderDesc + "";
              }

              new PopupForm("/admin/_list-export-csv?_params=" + encodeURIComponent(JSON.stringify(params)), (data: any) => {
                window.open(data.RedirectionLocation);
              });
            }
          },
          {
            Name: "Kontrola konzistence",
            Icon: "glyphicons-basic-322-shield-check.svg",
            Handler: () => {
              new PopupForm("/admin/_validation-consistency?resource=" + this.list.typeName, (data: any) => {
                //this.addUUID(data.Data);
              });
            }
          },
        ],
      })
    });
  }

  setVisibleColumns() {
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

  getSelectedColumnsMap(): any {
    let str = this.list.visibleColumnsStr;
    const map: any = {};
    str.split(",").forEach((key) => { map[key] = true; });
    return map;
  }
}
