
function initGoogleMaps() {
    console.log("google maps init");
    var viewEls = document.querySelectorAll(".admin_item_view_place");
    viewEls.forEach((el) => {
        initGoogleMapView(<HTMLDivElement>el);
    });
}

function initGoogleMapView(el: HTMLDivElement) {
    console.log("initing google maps view");
    console.log(el);

    var val = el.getAttribute("data-value");
    el.innerText = "";

    var coords = val.split(",");
    if (coords.length != 2) {
      el.classList.remove("admin_item_view_place");
      return;
    }

    const location = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) }; // Example: Poděbrady, Czech Republic

    // Create the map centered at the specified location
    //@ts-ignore
    const map = new google.maps.Map(el, {
        zoom: 14, // Adjust zoom level
        center: location,
    });

    // Add a marker at the specified location
    //@ts-ignore
    const marker = new google.maps.Marker({
        position: location,
        map: map,
    });
}