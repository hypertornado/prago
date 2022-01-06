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

    var formContainerElements =
      document.querySelectorAll<HTMLDivElement>(".form_container");
    formContainerElements.forEach((el) => {
      new FormContainer(el);
    });

    var imageViews = document.querySelectorAll<HTMLDivElement>(
      ".admin_item_view_image_content"
    );
    imageViews.forEach((el) => {
      new ImageView(el);
    });

    var mainMenuEl =
      document.querySelector<HTMLDivElement>(".admin_layout_left");
    if (mainMenuEl) {
      new MainMenu(mainMenuEl);
    }

    var relationListEls = document.querySelectorAll<HTMLDivElement>(
      ".admin_relationlist"
    );
    relationListEls.forEach((el) => {
      new RelationList(el);
    });

    new NotificationCenter(document.querySelector(".notification_center"));

    /*new Confirm("Hello world confirm", () => {
      console.log("ok");
    }, () => {
      console.log("cancel");
    }, ButtonStyle.Delete);

    */

    /*var content = document.createElement("div");
    content.innerHTML =
      "<h2>hello world</h2><br><textarea rows='10'></textarea>";

    var cp = new ContentPopup("info", content);
    cp.show();*/

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
