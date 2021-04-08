class Prago {
  private static x: number;

  static start() {
    document.addEventListener("DOMContentLoaded", Prago.init);
  }

  private static init() {
    var listEl = document.querySelector<HTMLDivElement>(".admin_list");
    if (listEl) {
      new List(listEl);
    }

    bindMarkdowns();
    bindTimestamps();
    bindRelations();
    bindImagePickers();
    bindForm();
    bindImageViews();
    bindDatePicker();
    bindDropdowns();
    bindSearch();
    bindMainMenu();
    bindRelationList();
    bindTaskMonitor();
    bindNotifications();

    /*new Confirm("Hello world confirm", () => {
      console.log("ok");
    }, () => {
      console.log("cancel");
    }, ButtonStyle.Delete);

    */

    //var content = document.createElement("div");
    //content.innerHTML = "<h2>hello world</h2><br><textarea rows='10'></textarea>";

    //new ContentPopup("info", content);

    //new Alert("OOO");

    //var loader = new LoadingPopup();
  }
}
Prago.start();
