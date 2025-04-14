
async function initGoogleMaps() {

    var viewElements = document.querySelectorAll(".admin_item_view_place");
    var pickerElements = document.querySelectorAll<HTMLDivElement>(".map_picker");

    if (viewElements.length == 0 && pickerElements.length == 0) {
        return;
    }


    //@ts-ignore
    const { Map } = await google.maps.importLibrary("maps");
    //@ts-ignore
    const { AdvancedMarkerElement, PinElement } = await google.maps.importLibrary("marker");

    viewElements.forEach((el) => {
        initGoogleMapView(<HTMLDivElement>el);
    });

    pickerElements.forEach((el) => {
        new GoogleMapEdit(el);
    });

}

function initGoogleMapView(el: HTMLDivElement) {
    var val = el.getAttribute("data-value");
    el.innerText = "";

    var coords = val.split(",");
    if (coords.length != 2) {
      el.classList.remove("admin_item_view_place");
      return;
    }

    const location = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) };

    //@ts-ignore
    const map = new google.maps.Map(el, {
        zoom: 14, // Adjust zoom level
        center: location,
        mapId: "3dab4a498a2dadb",
    });

    //@ts-ignore
    const marker = new google.maps.marker.AdvancedMarkerElement({
        position: location,
        map: map,
    });
}

class GoogleMapEdit {
    el: HTMLDivElement;
    input: HTMLInputElement;
    map: any;
    marker: any;
    icon: HTMLDivElement;
    statusEl: HTMLDivElement;
    deleteButton: HTMLButtonElement;
  
    constructor(el: HTMLDivElement) {
        this.el = el;

        this.statusEl = el.querySelector(".map_picker_description");
        var mapEl = el.querySelector(".map_picker_map");
        this.input = this.el.querySelector(".map_picker_value");
        this.deleteButton = el.querySelector(".map_picker_delete");

        const location = { lng: 14.41854, lat: 50.073658 };

        //@ts-ignore
        this.map = new google.maps.Map(mapEl, {
            zoom: 1, // Adjust zoom level
            center: location,
            mapId: "3dab4a498a2dadb",
        });


        //@ts-ignore
        this.marker = new google.maps.marker.AdvancedMarkerElement({
            position: location,
            map: null,
            gmpDraggable: true,
        });

        this.marker.addListener("gmp-click", (e: any) => {
            this.deleteValue();
        })

        this.marker.addListener("drag", (e: any) => {
            this.setValue(e.latLng.lat(), e.latLng.lng());
            //this.deleteValue();
        })

        this.map.addListener("click", (e: any) => {
            this.setValue(e.latLng.lat(), e.latLng.lng());
        })

        this.deleteButton.addEventListener("click", () => {
            this.deleteValue();
        })

        var inVals = this.input.value.split(",");
        if (inVals.length == 2) {
            let lat = parseFloat(inVals[0]);
            let lng = parseFloat(inVals[1]);
            this.setValue(lat, lng);
            this.centreMap(lat, lng);
        } else {
            this.deleteValue();
        }
    }

    centreMap(lat: number, lng: number) {
        let location = {
            lat: lat,
            lng: lng,
        }
        this.map.setCenter(location);
        this.map.setZoom(14);
    }
  
    setValue(lat: number, lng: number) {
        let location = {
            lat: lat,
            lng: lng,
        }
        this.marker.position = location;
        this.marker.map = this.map;
        this.input.value = lat + "," + lng;
        this.statusEl.textContent = "Latitude: " + lat + ", Longitude: " + lng;
        this.deleteButton.classList.remove("hidden");
    }

    deleteValue() {
        this.marker.map = null;
        this.input.value = "";
        this.statusEl.textContent = "Polohu vyberete kliknutím na mapu";
        this.deleteButton.classList.add("hidden");
    }

  }