class Prago {
  private static x: number;

  static start() {
    document.addEventListener("DOMContentLoaded", Prago.init);
  }

  private static placesEditArr: PlacesEdit[] = [];

  private static init() {
    var listEl = document.querySelector<HTMLDivElement>(".admin_list");
    if (listEl) {
      new List(listEl);
    }

    var formElements = document.querySelectorAll<HTMLFormElement>(
      ".prago_form"
    );
    formElements.forEach((el) => {
      new Form(el);
    });

    var imageViews = document.querySelectorAll<HTMLDivElement>(
      ".admin_item_view_image_content"
    );
    imageViews.forEach((el) => {
      new ImageView(el);
    });

    var mainMenuEl = document.querySelector<HTMLDivElement>(
      ".admin_layout_left"
    );
    if (mainMenuEl) {
      new MainMenu(mainMenuEl);
    }

    var relationListEls = document.querySelectorAll<HTMLDivElement>(
      ".admin_relationlist"
    );
    relationListEls.forEach((el) => {
      new RelationList(el);
    });

    new TaskMonitor();
    new NotificationCenter(document.querySelector(".notification_center"));

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

  static registerPlacesEdit(place: PlacesEdit) {
    Prago.placesEditArr.push(place);
  }

  private static googleMapsInited = false;

  static initGoogleMaps() {
    Prago.googleMapsInited = true;
    Prago.placesEditArr.forEach((placeEdit) => {
      placeEdit.start();
    });
  }
}
Prago.start();

function googleMapsInited() {
  var els = document.querySelectorAll<HTMLDivElement>(".admin_item_view_place");
  els.forEach((el) => {
    new PlacesView(el);
  });

  Prago.initGoogleMaps();
}
