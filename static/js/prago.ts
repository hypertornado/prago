class Prago {
  static notificationCenter: NotificationCenter;
  static shortcuts: Shortcuts;
  static cmenu: CMenu;

  static start() {
    document.addEventListener("DOMContentLoaded", Prago.init);
  }

  private static init() {
    Prago.shortcuts = new Shortcuts(document.body);
    Prago.shortcuts.addRootShortcuts();
    Prago.cmenu = new CMenu();

    var listEl = document.querySelector<HTMLDivElement>(".list");
    if (listEl) {
      new List(listEl);
    }

    var formContainerElements =
      document.querySelectorAll<HTMLDivElement>(".form_container");
    formContainerElements.forEach((el) => {
      new FormContainer(el, (data: any) => {
        if (data.RedirectionLocation) {
          window.location = data.RedirectionLocation;
        }
      });
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

    Prago.notificationCenter = new NotificationCenter(
      document.querySelector(".notification_center")
    );

    var qa: HTMLDivElement = document.querySelector(".quick_actions");
    if (qa) {
      new QuickActions(qa);
    }

    initDashboard();
    initGoogleMaps();

    document.querySelector(".searchbox_button").addEventListener("click", (e: Event) => {
      let input: HTMLInputElement = document.querySelector(".searchbox_input");
      if (!input.value) {
        input.focus();
        e.stopPropagation();
        e.preventDefault();
      }
    })

    //Prago.testPopupForm();
  }

  static testPopupForm() {
    new PopupForm("/admin/packageview/new", (data: any) => {
      console.log("form data");
      console.log(data);
    });
    //new PopupForm("/admin/hotel/new");

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
