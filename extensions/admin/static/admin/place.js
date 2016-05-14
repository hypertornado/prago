function bindPlaces() {

  function bindPlace(el) {
    var mapEl = $("<div></div>").addClass("admin_place_map");
    el.append(mapEl);

    var position = {lat: 50.0796284, lng: 14.4292577};
    var zoom = 1;
    var visible = false;

    var input = el.find("input");

    var inVal = input.val();
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

    var map = new google.maps.Map(mapEl[0], {
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
      input.val(str);
    });

    marker.addListener("click", function () {
      marker.setVisible(false);
      input.val("");
    })

    map.addListener('click', function(e) {
      position.lat = e.latLng.lat();
      position.lng = e.latLng.lng();
      marker.setPosition(position);
      marker.setVisible(true);
    });

    function stringifyPosition(lat, lng) {
      return lat + "," + lng;
    }

  }

  $(".admin_place").each(
    function() {
      bindPlace($(this));
    }
  );
}