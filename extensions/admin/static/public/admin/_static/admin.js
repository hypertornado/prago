function DOMinsertChildAtIndex(parent, child, index) {
    if (index >= parent.children.length) {
        parent.appendChild(child);
    }
    else {
        parent.insertBefore(child, parent.children[index]);
    }
}
function bindImagePicker() {
    var popup = document.getElementById("admin_images_popup");
    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    popup.addEventListener("click", function (e) {
        if (e.target == popup) {
            hidePopup();
        }
    });
    var loadedContainer = popup.getElementsByClassName("admin_images_popup_box_new_list")[0];
    var selectedContainer = popup.getElementsByClassName("admin_images_popup_box_content")[0];
    var doFilter = function () {
        loadedContainer.textContent = "Loading...";
        var popupFilter = popup.getElementsByClassName("admin_images_popup_filter")[0];
        loadImages("", popupFilter.value, function (items) {
            loadedContainer.textContent = "";
            items.forEach(function (item) {
                var img = itemToImg(item);
                img.addEventListener("click", function (e) {
                    var currentTarget = e.currentTarget;
                    var cloned = currentTarget.cloneNode(true);
                    cloned.addEventListener("click", function (event) {
                        var currentTarget = event.currentTarget;
                        currentTarget.remove();
                    });
                    bindDraggableEvents(cloned);
                    selectedContainer.appendChild(cloned);
                });
                loadedContainer.appendChild(img);
            });
        });
    };
    var draggedElement;
    function bindDraggableEvents(el) {
        el.addEventListener("dragstart", function () {
            draggedElement = this;
        });
        el.addEventListener("drop", function (e) {
            if (this != draggedElement) {
                var uid = this.getAttribute("data-uid");
                var src = this.getAttribute("src");
                this.setAttribute("data-uid", draggedElement.getAttribute("data-uid"));
                this.setAttribute("src", draggedElement.getAttribute("src"));
                draggedElement.setAttribute("data-uid", uid);
                draggedElement.setAttribute("src", src);
            }
        });
        el.addEventListener("dragover", function (e) {
            e.preventDefault();
        });
    }
    function itemToImg(item) {
        var img = document.createElement("img");
        img.classList.add("admin_images_img");
        img.setAttribute("src", item.Thumb);
        img.setAttribute("data-uid", item.UID);
        return img;
    }
    function createDraggableImg(item) {
        var img = itemToImg(item);
        img.setAttribute("draggable", "true");
        bindDraggableEvents(img);
        img.addEventListener("click", function (event) {
            var element = event.currentTarget;
            element.remove();
        });
        return img;
    }
    function loadImageToPopup(value) {
        if (value.length > 0) {
            selectedContainer.textContent = "Loading...";
            loadImages(value, "", function (items) {
                selectedContainer.textContent = "";
                items.forEach(function (item) {
                    var img = createDraggableImg(item);
                    selectedContainer.appendChild(img);
                });
            });
        }
    }
    var connectedItem;
    popup.getElementsByClassName("admin_images_popup_save")[0].addEventListener("click", function () {
        hidePopup();
        var items = [];
        for (var i = 0; i < selectedContainer.children.length; i++) {
            items.push(selectedContainer.children[i].getAttribute("data-uid"));
        }
        var str = items.join(",");
        connectedItem.getElementsByTagName("input")[0].value = str;
        showPreview(connectedItem);
    });
    popup.getElementsByClassName("admin_images_popup_cancel")[0].addEventListener("click", hidePopup);
    popup.getElementsByClassName("admin_images_popup_filter_button")[0].addEventListener("click", doFilter);
    function showPopup(el) {
        connectedItem = el;
        var val = el.getElementsByTagName("input")[0].value;
        loadImageToPopup(val);
        var focusable = document.getElementsByClassName("admin_images_popup_box")[0];
        focusable.focus();
        popup.style.display = "block";
        doFilter();
    }
    function hidePopup() {
        popup.style.display = "none";
    }
    function loadImages(ids, q, handler) {
        var url = adminPrefix + "/_api/image/list?";
        if (ids.length > 0) {
            url += "ids=" + encodeURIComponent(ids);
        }
        else {
            url += "q=" + encodeURIComponent(q);
        }
        var request = new XMLHttpRequest();
        request.open("GET", url, true);
        request.onload = function () {
            if (this.status == 200) {
                handler(JSON.parse(this.response));
            }
            else {
                console.error("Error while loading images.");
            }
        };
        request.send();
    }
    function bindImage(el) {
        showPreview(el);
        el.getElementsByClassName("admin_images_edit")[0].addEventListener("click", function () {
            showPopup(el);
            return false;
        });
    }
    function showPreview(el) {
        var value = el.getElementsByTagName("input")[0].value;
        var list = el.getElementsByClassName("admin_images_list")[0];
        list.textContent = "";
        el.getElementsByClassName("admin_images_edit")[0].style.display = "none";
        el.getElementsByTagName("progress")[0].style.display = "";
        if (value.length > 0) {
            loadImages(value, "", function (items) {
                doneLoading(el);
                items.forEach(function (item) {
                    var link = document.createElement("a");
                    link.setAttribute("href", adminPrefix + "/file/" + item.ID);
                    link.setAttribute("target", "_blank");
                    var img = document.createElement("img");
                    img.setAttribute("src", item.Thumb);
                    img.classList.add("admin_images_img");
                    link.appendChild(img);
                    list.appendChild(link);
                });
            });
        }
        else {
            doneLoading(el);
        }
    }
    function doneLoading(el) {
        el.getElementsByClassName("admin_images_list")[0].textContent = "";
        el.getElementsByClassName("admin_images_edit")[0].style.display = "";
        el.getElementsByTagName("progress")[0].style.display = "none";
    }
    function showLoadedResult(text) {
        document.getElementsByClassName("admin_images_popup_box_upload_message")[0].textContent = text;
        document.querySelector(".admin_images_popup_box_upload_btn").style.display = "";
        document.querySelector(".admin_images_popup_box_upload input").style.display = "";
    }
    document.getElementsByClassName("admin_images_popup_box_upload_btn")[0].addEventListener("click", function (e) {
        var filesInput = document.querySelector(".admin_images_popup_box_upload input");
        var files = filesInput.files;
        var data = new FormData();
        Array.prototype.forEach.call(files, function (item, i) {
            data.append("file", item);
        });
        data.append("description", document.getElementsByClassName("admin_popup_file_description")[0].value);
        document.getElementsByClassName("admin_popup_file_description")[0].value = "";
        document.getElementsByClassName("admin_images_popup_box_upload_message")[0].textContent = "Uploading...";
        document.getElementsByClassName("admin_images_popup_box_upload_btn")[0].style.display = "none";
        document.querySelector(".admin_images_popup_box_upload input").style.display = "none";
        var request = new XMLHttpRequest();
        request.open("POST", adminPrefix + "/_api/image/upload");
        request.onload = function () {
            if (this.status == 200) {
                var items = JSON.parse(this.response);
                Array.prototype.forEach.call(items, function (item, i) {
                    var img = createDraggableImg(item);
                    selectedContainer.appendChild(img);
                });
                showLoadedResult("Uploaded successfully.");
            }
            else {
                showLoadedResult("Error while uploading files.");
            }
        };
        request.send(data);
    });
    var elements = document.querySelectorAll(".admin_images");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindImage(el);
    });
}
function bindOrder() {
    function orderTable(el) {
        var rows = el.getElementsByClassName("admin_table_row");
        Array.prototype.forEach.call(rows, function (item, i) {
            bindDraggable(item);
        });
        var draggedElement;
        function bindDraggable(row) {
            row.setAttribute("draggable", "true");
            row.addEventListener("dragstart", function (ev) {
                draggedElement = this;
                ev.dataTransfer.setData('text/plain', '');
            });
            row.addEventListener("drop", function (ev) {
                var targetEl = this;
                if (this != draggedElement) {
                    var draggedIndex = -1;
                    var thisIndex = -1;
                    Array.prototype.forEach.call(el.getElementsByClassName("admin_table_row"), function (item, i) {
                        if (item == draggedElement) {
                            draggedIndex = i;
                        }
                        if (item == targetEl) {
                            thisIndex = i;
                        }
                    });
                    if (draggedIndex <= thisIndex) {
                        thisIndex += 1;
                    }
                    DOMinsertChildAtIndex(targetEl.parentElement, draggedElement, thisIndex + 1);
                    saveOrder();
                }
                return false;
            });
            row.addEventListener("dragover", function (ev) {
                ev.preventDefault();
            });
        }
        function saveOrder() {
            var ajaxPath = document.location.pathname + "/order";
            var order = [];
            var rows = el.getElementsByClassName("admin_table_row");
            Array.prototype.forEach.call(rows, function (item, i) {
                order.push(parseInt(item.getAttribute("data-id")));
            });
            var request = new XMLHttpRequest();
            request.open("POST", ajaxPath, true);
            request.onload = function () {
                if (this.status != 200) {
                    console.error("Error while saving order.");
                }
            };
            request.send(JSON.stringify({ "order": order }));
        }
    }
    var elements = document.querySelectorAll(".admin_table-order");
    Array.prototype.forEach.call(elements, function (el, i) {
        orderTable(el);
    });
}
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
function bindDelete() {
    var deleteButtons = document.querySelectorAll(".admin-action-delete");
    for (var i = 0; i < deleteButtons.length; i++) {
        bindDeleteButton(deleteButtons[i]);
    }
}
function bindDeleteButton(btn) {
    btn.addEventListener("click", function () {
        var message = btn.getAttribute("data-confirm-message");
        var url = btn.getAttribute("data-action");
        if (confirm(message)) {
            var request = new XMLHttpRequest();
            request.open("POST", url, true);
            request.onload = function () {
                if (this.status == 200) {
                    document.location.reload();
                }
                else {
                    console.error("Error while deleting item");
                }
            };
            request.send();
        }
    });
}
window.onload = function () {
    bindOrder();
    bindMarkdowns();
    bindTimestamps();
    bindRelations();
    bindImagePicker();
    bindDelete();
};
