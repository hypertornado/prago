function bindPlacesView() {
  var els = document.querySelectorAll(".admin_item_view_place");
  for (var i = 0; i < els.length; i++) {
    new PlacesView(<HTMLDivElement>els[i]);
  }
}

class PlacesView {
  constructor(el: HTMLDivElement) {
    var val = el.getAttribute("data-value");
    el.innerText = "";

    var coords = val.split(",");
    if (coords.length != 2) {
      el.innerText = "-";
      el.classList.remove("admin_item_view_place");
      return
    }


    var position = {lat: parseFloat(coords[0]), lng: parseFloat(coords[1])};
    var zoom = 18;

    var map = new google.maps.Map(el, {
      center: position,
      zoom: zoom
    });

    var marker = new google.maps.Marker({
      position: position,
      map: map,
      draggable: true,
      title: ""
    });

  }
}