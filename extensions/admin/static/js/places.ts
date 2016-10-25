function bindPlaces() {
  function bindPlace(el: any) {
    var mapEl = document.createElement("div")
    mapEl.classList.add("admin_place_map");
    el.appendChild(mapEl);

    var position = {lat: 50.0796284, lng: 14.4292577};
    var zoom = 1;
    var visible = false;

    var input = el.getElementsByTagName("input")[0]

    var inVal = input.value;
    var inVals = inVal.split(",");
    if (inVals.length == 2) {
      inVals[0] = parseFloat(inVals[0]);
      inVals[1] = parseFloat(inVals[1]);
      if (!isNaN(inVals[0]) && !isNaN(inVals[1])) {
        position.lat = inVals[0];
        position.lng = inVals[1];
        zoom = 11;
        visible = true;
      }
    }

    var map = new google.maps.Map(mapEl, {
      center: position,
      zoom: zoom
    });

    var marker = new google.maps.Marker({
      position: position,
      map: map,
      draggable: true,
      title: "",
      visible: visible
    });

    marker.addListener("position_changed", function () {
      var p = marker.getPosition();
      var str = stringifyPosition(p.lat(), p.lng());
      input.value = str;
    });

    marker.addListener("click", function () {
      marker.setVisible(false);
      input.value = "";
    })


    map.addListener('click', function(e: any) {
      position.lat = e.latLng.lat();
      position.lng = e.latLng.lng();
      marker.setPosition(position);
      marker.setVisible(true);
    });

    function stringifyPosition(lat: number, lng: number) {
      return lat + "," + lng;
    }

  }

  var elements = document.querySelectorAll(".admin_place");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    bindPlace(el);
  });
}