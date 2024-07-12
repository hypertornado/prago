class QuickActions {
  constructor(el: HTMLDivElement) {
    var buttons = el.querySelectorAll(".quick_actions_btn");

    for (var i = 0; i < buttons.length; i++) {
      let button = buttons[i];
      button.addEventListener("click", this.buttonClicked.bind(this));
    }
  }

  buttonClicked(e: Event) {
    var btn: HTMLButtonElement = <HTMLButtonElement>e.target;
    let actionURL = btn.getAttribute("data-url");

    new Confirm("Potvrdit akci", () => {
      let lp = new LoadingPopup();
      fetch(actionURL, {
        method: "POST",
      })
        .then((response) => {
          lp.done();
          if (response.ok) {
            return response.text();
          } else {
            throw response.text();
          }
        })
        .then((val) => {
          location.reload();
        })
        .catch((val) => {
          return val;
        })
        .then((val) => {
          if (val) {
            new Alert(val);
          }
        });
    });
  }
}
