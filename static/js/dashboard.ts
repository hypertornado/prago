function initDashdoard() {
  var dashboardTables =
    document.querySelectorAll<HTMLDivElement>(".board_table");
  dashboardTables.forEach((el) => {
    new DashboardTable(el);
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
