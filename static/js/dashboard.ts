function initDashdoard() {
  var dashboardTables =
    document.querySelectorAll<HTMLDivElement>(".dashboard_table");
  dashboardTables.forEach((el) => {
    new DashboardTable(el);
  });

  var dashboardFigures =
    document.querySelectorAll<HTMLDivElement>(".dashboard_figure");
  dashboardFigures.forEach((el) => {
    new DashboardFigure(el);
  });
}

class DashboardTable {
  constructor(el: HTMLDivElement) {
    let uuid = el.getAttribute("data-uuid");

    var request = new XMLHttpRequest();
    var params: any = {
      uuid: uuid,
    };

    request.addEventListener("load", () => {
      if (request.status == 200) {
        el.innerHTML = request.response;
      } else {
        el.innerText = "Error while loading table";
      }
    });

    request.open(
      "GET",
      "/admin/api/dashboard-table" + encodeParams(params),
      true
    );

    request.send();
  }
}

class DashboardFigure {
  el: HTMLDivElement;
  valueEl: HTMLDivElement;
  descriptionEl: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    this.valueEl = el.querySelector(".dashboard_figure_value");
    this.descriptionEl = el.querySelector(".dashboard_figure_description");

    let uuid = el.getAttribute("data-uuid");

    var request = new XMLHttpRequest();
    var params: any = {
      uuid: uuid,
    };

    request.addEventListener("load", () => {
      this.el.classList.remove("dashboard_figure-loading");
      if (request.status == 200) {
        let data = JSON.parse(request.response);
        this.valueEl.innerText = data["Value"];
        this.valueEl.setAttribute("title", data["Value"]);
        this.descriptionEl.innerText = data["Description"];
        this.descriptionEl.setAttribute("title", data["Description"]);

        if (data["IsRed"]) {
          this.el.classList.add("dashboard_figure-red");
        }
        if (data["IsGreen"]) {
          this.el.classList.add("dashboard_figure-green");
        }
      } else {
        this.valueEl.innerText = "Error while loading item.";
      }
    });

    request.open(
      "GET",
      "/admin/api/dashboard-figure" + encodeParams(params),
      true
    );

    this.el.classList.remove("dashboard_figure-green", "dashboard_figure-red");

    this.el.classList.add("dashboard_figure-loading");

    this.valueEl.innerText = "Loading...";
    this.descriptionEl.innerText = "";

    request.send();
  }
}
