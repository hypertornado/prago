class Prago {
  //private static x: number;

  static start() {
    document.addEventListener("DOMContentLoaded", Prago.init);
  }

  private static init() {
    Prago.heightListenerSetup();

    var listEl = document.querySelector<HTMLDivElement>(".list");
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

    var menuEl = document.querySelector<HTMLDivElement>(".root_left");
    if (menuEl) {
      new Menu();
    }

    var relationListEls = document.querySelectorAll<HTMLDivElement>(
      ".admin_relationlist"
    );
    relationListEls.forEach((el) => {
      new RelationList(el);
    });

    new NotificationCenter(document.querySelector(".notification_center"));

    var qa: HTMLDivElement = document.querySelector(".quick_actions");
    if (qa) {
      new QuickActions(qa);
    }

    initDashdoard();

    initSMap();
  }

  private static heightListenerSetup() {
    let appHeight = () => {
      const doc = document.documentElement;
      doc.style.setProperty("--app-height", `${window.innerHeight}px`);
    };
    window.addEventListener("resize", appHeight);
    appHeight();
  }
}
Prago.start();

class VisibilityReloader {
  lastRequestedTime: number;

  constructor(reloadIntervalMilliseconds: number, handler: any) {
    this.lastRequestedTime = 0;
    window.setInterval(() => {
      if (
        document.visibilityState == "visible" &&
        Date.now() - this.lastRequestedTime >= reloadIntervalMilliseconds
      ) {
        this.lastRequestedTime = Date.now();
        handler();
      }
    }, 100);
  }
}
