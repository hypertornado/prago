function bindMarkdowns() {
    function bindMarkdown(el) {
        var textarea = el.getElementsByTagName("textarea")[0];
        var lastChanged = Date.now();
        var changed = false;
        setInterval(function () {
            if (changed && (Date.now() - lastChanged > 500)) {
                loadPreview();
            }
        }, 100);
        textarea.addEventListener("change", textareaChanged);
        textarea.addEventListener("keyup", textareaChanged);
        function textareaChanged() {
            changed = true;
            lastChanged = Date.now();
        }
        loadPreview();
        function loadPreview() {
            changed = false;
            var request = new XMLHttpRequest();
            request.open("POST", document.body.getAttribute("data-admin-prefix") + "/_api/markdown", true);
            request.onload = function () {
                if (this.status == 200) {
                    console.log(JSON.parse(this.response));
                    var previewEl = el.getElementsByClassName("admin_markdown_preview")[0];
                    previewEl.innerHTML = JSON.parse(this.response);
                }
                else {
                    console.error("Error while loading markdown preview.");
                }
            };
            request.send(textarea.value);
        }
    }
    var elements = document.querySelectorAll(".admin_markdown");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindMarkdown(el);
    });
}
function bindTimestamps() {
    function bindTimestamp(el) {
        var hidden = el.getElementsByTagName("input")[0];
        var v = hidden.value;
        if (v == "0001-01-01 00:00") {
            var d = new Date();
            var month = d.getMonth() + 1;
            var monthStr = String(month);
            if (month < 10) {
                monthStr = "0" + monthStr;
            }
            var day = d.getUTCDate();
            var dayStr = String(day);
            if (day < 10) {
                dayStr = "0" + dayStr;
            }
            v = d.getFullYear() + "-" + monthStr + "-" + dayStr + " " + d.getHours() + ":" + d.getMinutes();
        }
        var date = v.split(" ")[0];
        var hour = parseInt(v.split(" ")[1].split(":")[0]);
        var minute = parseInt(v.split(" ")[1].split(":")[1]);
        var timestampEl = el.getElementsByClassName("admin_timestamp_date")[0];
        timestampEl.value = date;
        var hourEl = el.getElementsByClassName("admin_timestamp_hour")[0];
        for (var i = 0; i < 24; i++) {
            var newEl = document.createElement("option");
            var addVal = "" + i;
            if (i < 10) {
                addVal = "0" + addVal;
            }
            newEl.innerText = addVal;
            newEl.setAttribute("value", addVal);
            if (hour == i) {
                newEl.setAttribute("selected", "selected");
            }
            hourEl.appendChild(newEl);
        }
        var minEl = el.getElementsByClassName("admin_timestamp_minute")[0];
        for (var i = 0; i < 60; i++) {
            var newEl = document.createElement("option");
            var addVal = "" + i;
            if (i < 10) {
                addVal = "0" + addVal;
            }
            newEl.innerText = addVal;
            newEl.setAttribute("value", addVal);
            if (minute == i) {
                newEl.setAttribute("selected", "selected");
            }
            minEl.appendChild(newEl);
        }
        var elTsDate = el.getElementsByClassName("admin_timestamp_date")[0];
        var elTsHour = el.getElementsByClassName("admin_timestamp_hour")[0];
        var elTsMinute = el.getElementsByClassName("admin_timestamp_minute")[0];
        var elTsInput = el.getElementsByTagName("input")[0];
        function saveValue() {
            var str = elTsDate.value + " " + elTsHour.value + ":" + elTsMinute.value;
            elTsInput.value = str;
        }
        saveValue();
        elTsDate.addEventListener("change", saveValue);
        elTsHour.addEventListener("change", saveValue);
        elTsMinute.addEventListener("change", saveValue);
    }
    var elements = document.querySelectorAll(".admin_timestamp");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindTimestamp(el);
    });
}
function bindRelations() {
    function bindRelation(el) {
        var input = el.getElementsByTagName("input")[0];
        var relationName = input.getAttribute("data-relation");
        var originalValue = input.value;
        var select = document.createElement("select");
        select.classList.add("input");
        select.classList.add("form_input");
        select.addEventListener("change", function () {
            input.value = select.value;
        });
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/resource/" + relationName, true);
        var progress = el.getElementsByTagName("progress")[0];
        request.onload = function () {
            if (this.status >= 200 && this.status < 400) {
                var resp = JSON.parse(this.response);
                addOption(select, "0", "", false);
                Array.prototype.forEach.call(resp, function (item, i) {
                    var selected = false;
                    if (originalValue == item["id"]) {
                        selected = true;
                    }
                    addOption(select, item["id"], item["name"], selected);
                });
                el.appendChild(select);
            }
            else {
                console.error("Error wile loading relation " + relationName + ".");
            }
            progress.style.display = 'none';
        };
        request.onerror = function () {
            console.error("Error wile loading relation " + relationName + ".");
            progress.style.display = 'none';
        };
        request.send();
    }
    function addOption(select, value, description, selected) {
        var option = document.createElement("option");
        if (selected) {
            option.setAttribute("selected", "selected");
        }
        option.setAttribute("value", value);
        option.innerText = description;
        select.appendChild(option);
    }
    var elements = document.querySelectorAll(".admin_item_relation");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindRelation(el);
    });
}
function bindPlaces() {
    function bindPlace(el) {
        var mapEl = document.createElement("div");
        mapEl.classList.add("admin_place_map");
        el.appendChild(mapEl);
        var position = { lat: 50.0796284, lng: 14.4292577 };
        var zoom = 1;
        var visible = false;
        var input = el.getElementsByTagName("input")[0];
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
        });
        map.addListener('click', function (e) {
            position.lat = e.latLng.lat();
            position.lng = e.latLng.lng();
            marker.setPosition(position);
            marker.setVisible(true);
        });
        function stringifyPosition(lat, lng) {
            return lat + "," + lng;
        }
    }
    var elements = document.querySelectorAll(".admin_place");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindPlace(el);
    });
}
window.onload = function () {
    bindMarkdowns();
    bindTimestamps();
    bindRelations();
};
