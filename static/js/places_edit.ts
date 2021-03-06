class PlacesEdit {
  private el: HTMLDivElement;

  constructor(el: any) {
    this.el = el;
    Prago.registerPlacesEdit(this);
  }

  start() {
    var mapEl = document.createElement("div");
    mapEl.classList.add("admin_place_map");
    this.el.appendChild(mapEl);

    var position = { lat: 0, lng: 0 };
    var zoom = 1;
    var visible = false;

    var input = this.el.getElementsByTagName("input")[0];

    var inVal = input.value;
    var inVals = inVal.split(",");
    if (inVals.length == 2) {
      var lat = parseFloat(inVals[0]);
      var lon = parseFloat(inVals[1]);
      if (!isNaN(lat) && !isNaN(lon)) {
        position.lat = lat;
        position.lng = lon;
        zoom = 11;
        visible = true;
      }
    }

    var map = new google.maps.Map(mapEl, {
      center: position,
      zoom: zoom,
    });

    var marker = new google.maps.Marker({
      position: position,
      map: map,
      draggable: true,
      title: "",
      visible: visible,
    });

    var searchInput: HTMLInputElement = document.createElement("input");
    searchInput.classList.add("input", "input-placesearch");
    var searchBox = new google.maps.places.SearchBox(searchInput);
    map.controls[google.maps.ControlPosition.LEFT_TOP].push(searchInput);
    searchBox.addListener("places_changed", () => {
      var places = searchBox.getPlaces();
      if (places.length > 0) {
        map.fitBounds(places[0].geometry.viewport);
        marker.setPosition({
          lat: places[0].geometry.location.lat(),
          lng: places[0].geometry.location.lng(),
        });
        marker.setVisible(true);
      }
    });

    searchInput.addEventListener("keydown", (e) => {
      if (e.keyCode == 13) {
        e.preventDefault();
        return false;
      }
    });

    marker.addListener("position_changed", function () {
      var p = marker.getPosition();
      var str = stringifyPosition(p.lat(), p.lng());
      input.value = str;
    });

    marker.addListener("click", function () {
      marker.setVisible(false);
      input.value = "";
    });

    map.addListener("click", function (e: any) {
      position.lat = e.latLng.lat();
      position.lng = e.latLng.lng();
      marker.setPosition(position);
      marker.setVisible(true);
    });

    function stringifyPosition(lat: number, lng: number) {
      return lat + "," + lng;
    }
  }
}
