//var SMap: any;
declare var SMap: any;
declare var JAK: any;
declare var Loader: any;

function initSMap() {
  if (!window.Loader) {
    return;
  }
  Loader.async = true;
  Loader.load(null, null, loadSMap);
}

function loadSMap() {
  /*var viewEls = document.querySelectorAll(".admin_item_view_place");
  viewEls.forEach((el) => {
    new SMapView(<HTMLDivElement>el);
  });*/

  var elements = document.querySelectorAll<HTMLDivElement>(".admin_place");
  elements.forEach((el) => {
    new SMapEdit(el);
  });
}

class SMapView {
  constructor(el: HTMLDivElement) {
    var val = el.getAttribute("data-value");
    el.innerText = "";

    var coords = val.split(",");
    if (coords.length != 2) {
      el.classList.remove("admin_item_view_place");
      return;
    }

    var stred = SMap.Coords.fromWGS84(coords[1], coords[0]);
    var mapa = new SMap(el, stred, 14);
    mapa.addDefaultLayer(SMap.DEF_BASE).enable();
    mapa.addDefaultControls();

    var vrstvaZnacek = new SMap.Layer.Marker(stred);
    mapa.addLayer(vrstvaZnacek);
    vrstvaZnacek.enable();

    var options = {};
    var marker = new SMap.Marker(stred, "myMarker", options);
    vrstvaZnacek.addMarker(marker);
  }
}

class SMapEdit {
  el: HTMLDivElement;
  input: HTMLInputElement;
  marker: any;
  icon: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;

    var mapEl = document.createElement("div");
    mapEl.classList.add("admin_place_map");
    this.el.appendChild(mapEl);

    this.input = this.el.querySelector(".admin_place_value");

    var zoom = 1;

    var coords = SMap.Coords.fromWGS84(14.41854, 50.073658);
    var mapa = new SMap(this.el, coords, 1);
    mapa.addDefaultLayer(SMap.DEF_BASE).enable();
    mapa.addDefaultControls();

    var vrstvaZnacek = new SMap.Layer.Marker(coords);
    mapa.addLayer(vrstvaZnacek);
    vrstvaZnacek.disable();

    this.icon = this.createMarkerIcon();
    this.icon.addEventListener("click", (e) => {
      vrstvaZnacek.disable();
      this.input.value = "";
      e.stopPropagation();
      e.preventDefault();
      return false;
    });

    var options = {
      url: this.icon,
      title: "",
      anchor: { left: 10, top: 10 },
    };

    this.marker = new SMap.Marker(coords, "", options);
    vrstvaZnacek.addMarker(this.marker);

    var inVals = this.input.value.split(",");
    if (inVals.length == 2) {
      var lat = parseFloat(inVals[0]);
      var lon = parseFloat(inVals[1]);
      if (!isNaN(lat) && !isNaN(lon)) {
        coords = SMap.Coords.fromWGS84(lon, lat);
        mapa.setCenterZoom(coords, 10, false);
        this.marker.setCoords(coords);
        vrstvaZnacek.enable();
      }
    }

    mapa.getSignals().addListener(window, "map-click", (e: any, x: any) => {
      var coords = SMap.Coords.fromEvent(e.data.event, mapa);
      this.marker.setCoords(coords);
      vrstvaZnacek.enable();
      this.setValue();
    });
  }

  setValue() {
    let coords = this.marker.getCoords();
    let val = this.stringifyPosition(coords.y, coords.x);
    this.input.value = val;
  }

  stringifyPosition(lat: number, lng: number): string {
    return lat + "," + lng;
  }

  createMarkerIcon(): HTMLDivElement {
    var ret = document.createElement("div");
    ret.classList.add("smap_edit_label");
    ret.setAttribute("style", "");
    ret.innerText = "";
    return ret;
  }
}
