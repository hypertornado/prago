function bindRelationList() {
  var els = document.getElementsByClassName("admin_relationlist");
  for (var i = 0; i < els.length; i++) {
    new RelationList(<HTMLDivElement>els[i]);
  }
}

class RelationList {
  adminPrefix: string;

  targetEl: HTMLDivElement;

  sourceResource: string;
  targetResource: string;
  targetField: string;
  idValue: number;
  count: number;

  offset: number;

  loadingEl: HTMLDivElement;
  moreEl: HTMLDivElement;
  moreButton: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.adminPrefix = document.body.getAttribute("data-admin-prefix");

    this.targetEl = el.querySelector(".admin_relationlist_target");

    this.sourceResource = el.getAttribute("data-source-resource");
    this.targetResource = el.getAttribute("data-target-resource");
    this.targetField = el.getAttribute("data-target-field");
    this.idValue = parseInt(el.getAttribute("data-id-value"));
    this.count = parseInt(el.getAttribute("data-count"));

    this.offset = 0;

    this.loadingEl = el.querySelector(".admin_relationlist_loading");
    this.moreEl = el.querySelector(".admin_relationlist_more");
    this.moreButton = el.querySelector(".admin_relationlist_more .btn");

    this.moreButton.addEventListener("click", this.load.bind(this));

    this.load();
  }

  load() {
    this.loadingEl.classList.remove("hidden");
    this.moreEl.classList.add("hidden");

    var request = new XMLHttpRequest();
    request.open("POST", this.adminPrefix + "/api/relationlist", true);
    request.addEventListener("load", () => {
      this.loadingEl.classList.add("hidden");
      if (request.status == 200) {
        this.offset += 10;

        var parentEl = document.createElement("div");
        parentEl.innerHTML = request.response;

        var parentAr = [];

        for (var i = 0; i < parentEl.children.length; i++) {
          parentAr.push(parentEl.children[i]);
        }

        for (var i = 0; i < parentAr.length; i++) {
          this.targetEl.appendChild(parentAr[i]);
        }

        if (this.offset < this.count) {
          this.moreEl.classList.remove("hidden");
        }
      } else {
        console.error("Error while RelationList request");
      }
    });
    request.send(
      JSON.stringify({
        SourceResource: this.sourceResource,
        TargetResource: this.targetResource,
        TargetField: this.targetField,
        IDValue: this.idValue,
        Offset: this.offset,
        Count: 10,
      })
    );
  }
}
