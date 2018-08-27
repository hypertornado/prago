function bindStats() {
    var elements = document.querySelectorAll(".admin_stats_pie");
    Array.prototype.forEach.call(elements, function (el, i) {
        new PieChart(el);
    });
    var elements = document.querySelectorAll(".admin_stats_timeline");
    Array.prototype.forEach.call(elements, function (el, i) {
        new Timeline(el);
    });
}
class PieChart {
    constructor(el) {
        var canvas = el.querySelector("canvas");
        var ctx = canvas.getContext('2d');
        var labelA = el.getAttribute("data-label-a");
        var labelB = el.getAttribute("data-label-b");
        var valueA = parseInt(el.getAttribute("data-value-a"));
        var valueB = parseInt(el.getAttribute("data-value-b"));
        var data = {
            datasets: [{
                    data: [valueA, valueB],
                    backgroundColor: ["#4078c0", "#eee"]
                }],
            labels: [
                labelA,
                labelB
            ]
        };
        var myChart = new Chart(ctx, {
            type: "pie",
            data: data,
            options: {
                "responsive": false
            }
        });
    }
}
class Timeline {
    constructor(el) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        var resource = el.getAttribute("data-resource");
        var field = el.getAttribute("data-field");
        var canvas = el.querySelector("canvas");
        this.ctx = canvas.getContext('2d');
        this.loadData(resource, field);
    }
    loadData(resource, field) {
        var data = {
            resource: resource,
            field: field
        };
        var request = new XMLHttpRequest();
        request.open("POST", this.adminPrefix + "/_api/stats", true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                var parsed = JSON.parse(request.responseText);
                this.createChart(parsed.labels, parsed.values);
            }
            else {
                console.error("error while loading list");
            }
        });
        request.send(JSON.stringify(data));
    }
    createChart(labels, values) {
        var data = {
            labels: labels,
            datasets: [{
                    backgroundColor: '#4078c0',
                    data: values
                }]
        };
        var myChart = new Chart(this.ctx, {
            type: "bar",
            data: data,
            options: {
                legend: {
                    display: false
                },
                scales: {
                    yAxes: [{
                            ticks: {
                                beginAtZero: true
                            }
                        }]
                }
            }
        });
    }
}
class Autoresize {
    constructor(el) {
        this.el = el;
        this.el.addEventListener('input', this.resizeIt.bind(this));
        this.resizeIt();
    }
    resizeIt() {
        var height = this.el.scrollHeight + 2;
        this.el.style.height = height + 'px';
    }
}
function DOMinsertChildAtIndex(parent, child, index) {
    if (index >= parent.children.length) {
        parent.appendChild(child);
    }
    else {
        parent.insertBefore(child, parent.children[index]);
    }
}
function bindImageViews() {
    var els = document.querySelectorAll(".admin_item_view_image_content");
    for (var i = 0; i < els.length; i++) {
        new ImageView(els[i]);
    }
}
class ImageView {
    constructor(el) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.el = el;
        var ids = el.getAttribute("data-images").split(",");
        this.addImages(ids);
    }
    addImages(ids) {
        this.el.innerHTML = "";
        for (var i = 0; i < ids.length; i++) {
            if (ids[i] != "") {
                this.addImage(ids[i]);
            }
        }
    }
    addImage(id) {
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
        var img = document.createElement("img");
        img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
        img.setAttribute("draggable", "false");
        container.appendChild(img);
        this.el.appendChild(container);
    }
}
function bindImagePickers() {
    var els = document.querySelectorAll(".admin_images");
    for (var i = 0; i < els.length; i++) {
        new ImagePicker(els[i]);
    }
}
class ImagePicker {
    constructor(el) {
        this.el = el;
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.hiddenInput = el.querySelector(".admin_images_hidden");
        this.preview = el.querySelector(".admin_images_preview");
        this.fileInput = this.el.querySelector(".admin_images_fileinput");
        this.progress = this.el.querySelector("progress");
        this.el.querySelector(".admin_images_loaded").classList.remove("hidden");
        this.hideProgress();
        var ids = this.hiddenInput.value.split(",");
        this.el.addEventListener("click", (e) => {
            if (e.altKey) {
                var ids = window.prompt("IDs of images", this.hiddenInput.value);
                this.hiddenInput.value = ids;
                e.preventDefault();
                return false;
            }
        });
        this.fileInput.addEventListener("dragenter", (ev) => {
            this.fileInput.classList.add("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragleave", (ev) => {
            this.fileInput.classList.remove("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragover", (ev) => {
            ev.preventDefault();
        });
        this.fileInput.addEventListener("drop", (ev) => {
            var text = ev.dataTransfer.getData('Text');
            return;
        });
        for (var i = 0; i < ids.length; i++) {
            var id = ids[i];
            if (id) {
                this.addImage(id);
            }
        }
        this.fileInput.addEventListener("change", () => {
            var files = this.fileInput.files;
            var formData = new FormData();
            if (files.length == 0) {
                return;
            }
            for (var i = 0; i < files.length; i++) {
                formData.append("file", files[i]);
            }
            var request = new XMLHttpRequest();
            request.open("POST", this.adminPrefix + "/_api/image/upload");
            request.addEventListener("load", (e) => {
                this.hideProgress();
                if (request.status == 200) {
                    var data = JSON.parse(request.response);
                    for (var i = 0; i < data.length; i++) {
                        this.addImage(data[i].UID);
                    }
                }
                else {
                    alert("Error while uploading image.");
                    console.error("Error while loading item.");
                }
            });
            this.fileInput.type = "";
            this.fileInput.type = "file";
            this.showProgress();
            request.send(formData);
        });
    }
    updateHiddenData() {
        var ids = [];
        for (var i = 0; i < this.preview.children.length; i++) {
            var item = this.preview.children[i];
            var uuid = item.getAttribute("data-uuid");
            ids.push(uuid);
        }
        this.hiddenInput.value = ids.join(",");
    }
    addImage(id) {
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("data-uuid", id);
        container.setAttribute("draggable", "true");
        container.setAttribute("target", "_blank");
        container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
        container.addEventListener("dragstart", (e) => {
            this.draggedElement = e.target;
        });
        container.addEventListener("drop", (e) => {
            var droppedElement = e.toElement;
            if (!droppedElement) {
                droppedElement = e.originalTarget;
            }
            for (var i = 0; i < 3; i++) {
                if (droppedElement.nodeName == "A") {
                    break;
                }
                else {
                    droppedElement = droppedElement.parentElement;
                }
            }
            var draggedIndex = -1;
            var droppedIndex = -1;
            var parent = this.draggedElement.parentElement;
            for (var i = 0; i < parent.children.length; i++) {
                var child = parent.children[i];
                if (child == this.draggedElement) {
                    draggedIndex = i;
                }
                if (child == droppedElement) {
                    droppedIndex = i;
                }
            }
            if (draggedIndex == -1 || droppedIndex == -1) {
                return;
            }
            if (draggedIndex <= droppedIndex) {
                droppedIndex += 1;
            }
            DOMinsertChildAtIndex(parent, this.draggedElement, droppedIndex);
            this.updateHiddenData();
            e.preventDefault();
            return false;
        });
        container.addEventListener("dragover", (e) => {
            e.preventDefault();
        });
        container.addEventListener("click", (e) => {
            var target = e.target;
            if (target.classList.contains("admin_images_image_delete")) {
                var parent = e.currentTarget.parentNode;
                parent.removeChild(e.currentTarget);
                this.updateHiddenData();
                e.preventDefault();
                return false;
            }
        });
        var img = document.createElement("img");
        img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
        img.setAttribute("draggable", "false");
        container.appendChild(img);
        var del = document.createElement("div");
        del.textContent = "×";
        del.classList.add("admin_images_image_delete");
        container.appendChild(del);
        this.preview.appendChild(container);
        this.updateHiddenData();
    }
    hideProgress() {
        this.progress.classList.add("hidden");
    }
    showProgress() {
        this.progress.classList.remove("hidden");
    }
}
function bindLists() {
    var els = document.getElementsByClassName("admin_table-list");
    for (var i = 0; i < els.length; i++) {
        new List(els[i]);
    }
}
class List {
    constructor(el) {
        this.el = el;
        this.page = 1;
        this.typeName = el.getAttribute("data-type");
        if (!this.typeName) {
            return;
        }
        this.progress = el.querySelector(".admin_table_progress");
        this.tbody = el.querySelector("tbody");
        this.tbody.textContent = "";
        this.bindFilter();
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.prefilterField = el.getAttribute("data-prefilter-field");
        this.prefilterValue = el.getAttribute("data-prefilter-value");
        this.orderColumn = el.getAttribute("data-order-column");
        if (el.getAttribute("data-order-desc") == "true") {
            this.orderDesc = true;
        }
        else {
            this.orderDesc = false;
        }
        this.bindOrder();
        this.load();
    }
    load() {
        this.progress.classList.remove("hidden");
        var request = new XMLHttpRequest();
        request.open("POST", this.adminPrefix + "/_api/list/" + this.typeName + document.location.search, true);
        request.addEventListener("load", () => {
            this.tbody.innerHTML = "";
            if (request.status == 200) {
                this.tbody.innerHTML = request.response;
                var count = request.getResponseHeader("X-Count");
                var totalCount = request.getResponseHeader("X-Total-Count");
                var countStr = count + " / " + totalCount;
                this.el.querySelector(".admin_table_count").textContent = countStr;
                bindOrder();
                this.bindPagination();
                this.bindClick();
                this.tbody.classList.remove("admin_table_loading");
            }
            else {
                console.error("error while loading list");
            }
            this.progress.classList.add("hidden");
        });
        var requestData = this.getListRequest();
        request.send(JSON.stringify(requestData));
    }
    bindPagination() {
        var pages = this.el.querySelectorAll(".pagination_page");
        for (var i = 0; i < pages.length; i++) {
            var pageEl = pages[i];
            pageEl.addEventListener("click", (e) => {
                var el = e.target;
                var page = parseInt(el.getAttribute("data-page"));
                this.page = page;
                this.load();
                e.preventDefault();
                return false;
            });
        }
    }
    bindClick() {
        var rows = this.el.querySelectorAll(".admin_table_row");
        for (var i = 0; i < rows.length; i++) {
            var row = rows[i];
            var id = row.getAttribute("data-id");
            row.addEventListener("click", (e) => {
                console.log("ROOOOW");
                var target = e.target;
                if (target.classList.contains("preventredirect")) {
                    return;
                }
                var el = e.currentTarget;
                var url = el.getAttribute("data-url");
                window.location.href = url;
            });
            var buttons = row.querySelector(".admin_list_buttons");
            buttons.addEventListener("click", (e) => {
                var url = e.target.getAttribute("href");
                if (url != "") {
                    window.location.href = url;
                    e.preventDefault();
                    e.stopPropagation();
                    return false;
                }
            });
        }
    }
    bindOrder() {
        this.renderOrder();
        var headers = this.el.querySelectorAll(".admin_table_orderheader");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.addEventListener("click", (e) => {
                var el = e.target;
                var name = el.getAttribute("data-name");
                if (name == this.orderColumn) {
                    if (this.orderDesc) {
                        this.orderDesc = false;
                    }
                    else {
                        this.orderDesc = true;
                    }
                }
                else {
                    this.orderColumn = name;
                    this.orderDesc = false;
                }
                this.renderOrder();
                this.load();
                e.preventDefault();
                return false;
            });
        }
    }
    renderOrder() {
        var headers = this.el.querySelectorAll(".admin_table_orderheader");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.classList.remove("ordered");
            header.classList.remove("ordered-desc");
            var name = header.getAttribute("data-name");
            if (name == this.orderColumn) {
                header.classList.add("ordered");
                if (this.orderDesc) {
                    header.classList.add("ordered-desc");
                }
            }
        }
    }
    getListRequest() {
        var ret = {};
        ret.Page = this.page;
        ret.OrderBy = this.orderColumn;
        ret.OrderDesc = this.orderDesc;
        ret.Filter = this.getFilterData();
        ret.PrefilterField = this.prefilterField;
        ret.PrefilterValue = this.prefilterValue;
        return ret;
    }
    getFilterData() {
        var ret = {};
        var items = this.el.querySelectorAll(".admin_table_filter_item");
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            var typ = item.getAttribute("data-typ");
            var val = item.value.trim();
            if (val) {
                ret[typ] = val;
            }
        }
        return ret;
    }
    bindFilter() {
        this.bindFilterRelations();
        this.filterInputs = this.el.querySelectorAll(".admin_table_filter_item");
        for (var i = 0; i < this.filterInputs.length; i++) {
            var input = this.filterInputs[i];
            input.addEventListener("input", this.inputListener.bind(this));
        }
        this.inputPeriodicListener();
    }
    inputListener(e) {
        if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
            return;
        }
        this.tbody.classList.add("admin_table_loading");
        this.page = 1;
        this.changed = true;
        this.changedTimestamp = Date.now();
        this.progress.classList.remove("hidden");
    }
    bindFilterRelations() {
        var els = this.el.querySelectorAll(".admin_table_filter_item-relations");
        for (var i = 0; i < els.length; i++) {
            this.bindFilterRelation(els[i]);
        }
    }
    bindFilterRelation(select) {
        var typ = select.getAttribute("data-typ");
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/resource/" + typ, true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                var resp = JSON.parse(request.response);
                for (var item of resp) {
                    var option = document.createElement("option");
                    option.setAttribute("value", item.id);
                    option.innerText = item.name;
                    select.appendChild(option);
                }
            }
            else {
                console.error("Error wile loading relation " + typ + ".");
            }
        });
        request.send();
    }
    inputPeriodicListener() {
        setInterval(() => {
            if (this.changed == true && Date.now() - this.changedTimestamp > 500) {
                this.changed = false;
                this.load();
            }
        }, 200);
    }
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
                row.classList.add("admin_table_row-selected");
                draggedElement = this;
                ev.dataTransfer.setData('text/plain', '');
                ev.dataTransfer.effectAllowed = "move";
                var d = document.createElement("div");
                d.style.display = "none";
                ev.dataTransfer.setDragImage(d, 0, 0);
            });
            row.addEventListener("dragenter", function (ev) {
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
                    if (draggedIndex < thisIndex) {
                        thisIndex += 1;
                    }
                    DOMinsertChildAtIndex(targetEl.parentElement, draggedElement, thisIndex);
                }
                return false;
            });
            row.addEventListener("drop", function (ev) {
                saveOrder();
                row.classList.remove("admin_table_row-selected");
                return false;
            });
            row.addEventListener("dragover", function (ev) {
                ev.preventDefault();
            });
        }
        function saveOrder() {
            var adminPrefix = document.body.getAttribute("data-admin-prefix");
            var typ = document.querySelector(".admin_table-order").getAttribute("data-type");
            var ajaxPath = adminPrefix + "/_api/order/" + typ;
            var order = [];
            var rows = el.getElementsByClassName("admin_table_row");
            Array.prototype.forEach.call(rows, function (item, i) {
                order.push(parseInt(item.getAttribute("data-id")));
            });
            var request = new XMLHttpRequest();
            request.open("POST", ajaxPath, true);
            request.addEventListener("load", () => {
                if (request.status != 200) {
                    console.error("Error while saving order.");
                }
            });
            request.send(JSON.stringify({ "order": order }));
        }
    }
    var elements = document.querySelectorAll(".admin_table-order");
    Array.prototype.forEach.call(elements, function (el, i) {
        orderTable(el);
    });
}
function bindMarkdowns() {
    var elements = document.querySelectorAll(".admin_markdown");
    Array.prototype.forEach.call(elements, function (el, i) {
        new MarkdownEditor(el);
    });
}
class MarkdownEditor {
    constructor(el) {
        this.el = el;
        this.textarea = el.querySelector(".textarea");
        this.preview = el.querySelector(".admin_markdown_preview");
        new Autoresize(this.textarea);
        var prefix = document.body.getAttribute("data-admin-prefix");
        var helpLink = el.querySelector(".admin_markdown_show_help");
        helpLink.setAttribute("href", prefix + "/_help/markdown");
        this.lastChanged = Date.now();
        this.changed = false;
        let showChange = el.querySelector(".admin_markdown_preview_show");
        showChange.addEventListener("change", () => {
            this.preview.classList.toggle("hidden");
        });
        setInterval(() => {
            if (this.changed && (Date.now() - this.lastChanged > 500)) {
                this.loadPreview();
            }
        }, 100);
        this.textarea.addEventListener("change", this.textareaChanged.bind(this));
        this.textarea.addEventListener("keyup", this.textareaChanged.bind(this));
        this.loadPreview();
        this.bindCommands();
        this.bindShortcuts();
    }
    bindCommands() {
        var btns = this.el.querySelectorAll(".admin_markdown_command");
        for (var i = 0; i < btns.length; i++) {
            btns[i].addEventListener("mousedown", (e) => {
                var cmd = e.target.getAttribute("data-cmd");
                this.executeCommand(cmd);
                e.preventDefault();
                return false;
            });
        }
    }
    bindShortcuts() {
        this.textarea.addEventListener("keydown", (e) => {
            if (e.metaKey == false && e.ctrlKey == false) {
                return;
            }
            switch (e.keyCode) {
                case 66:
                    this.executeCommand("b");
                    break;
                case 73:
                    this.executeCommand("i");
                    break;
                case 75:
                    this.executeCommand("h2");
                    break;
                case 85:
                    this.executeCommand("a");
                    break;
            }
        });
    }
    executeCommand(commandName) {
        switch (commandName) {
            case "b":
                this.setAroundMarkdown("**", "**");
                break;
            case "i":
                this.setAroundMarkdown("*", "*");
                break;
            case "a":
                this.setAroundMarkdown("[", "]()");
                var newEnd = this.textarea.selectionEnd + 2;
                this.textarea.selectionStart = newEnd;
                this.textarea.selectionEnd = newEnd;
                break;
            case "h2":
                var start = "## ";
                var end = "";
                var text = this.textarea.value;
                if (text[this.textarea.selectionStart - 1] !== "\n") {
                    start = "\n" + start;
                }
                if (text[this.textarea.selectionEnd] !== "\n") {
                    end = "\n";
                }
                this.setAroundMarkdown(start, end);
                break;
        }
        this.textareaChanged();
    }
    setAroundMarkdown(before, after) {
        var text = this.textarea.value;
        var selected = text.substr(this.textarea.selectionStart, this.textarea.selectionEnd - this.textarea.selectionStart);
        var newText = text.substr(0, this.textarea.selectionStart);
        newText += before;
        var newStart = newText.length;
        newText += selected;
        var newEnd = newText.length;
        newText += after;
        newText += text.substr(this.textarea.selectionEnd, text.length);
        this.textarea.value = newText;
        this.textarea.selectionStart = newStart;
        this.textarea.selectionEnd = newEnd;
        this.textarea.focus();
    }
    textareaChanged() {
        this.changed = true;
        this.lastChanged = Date.now();
    }
    loadPreview() {
        this.changed = false;
        var request = new XMLHttpRequest();
        request.open("POST", document.body.getAttribute("data-admin-prefix") + "/_api/markdown", true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                this.preview.innerHTML = JSON.parse(request.response);
            }
            else {
                console.error("Error while loading markdown preview.");
            }
        });
        request.send(this.textarea.value);
    }
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
    var elements = document.querySelectorAll(".admin_item_relation");
    Array.prototype.forEach.call(elements, function (el, i) {
        new RelationPicker(el);
    });
}
class RelationPicker {
    constructor(el) {
        this.selectedClass = "admin_item_relation_picker_suggestion-selected";
        this.input = el.getElementsByTagName("input")[0];
        this.previewContainer = el.querySelector(".admin_item_relation_preview");
        this.relationName = el.getAttribute("data-relation");
        this.progress = el.querySelector("progress");
        this.changeSection = el.querySelector(".admin_item_relation_change");
        this.changeButton = el.querySelector(".admin_item_relation_change_btn");
        this.changeButton.addEventListener("click", () => {
            this.showSearch();
            this.pickerInput.focus();
        });
        this.suggestionsEl = el.querySelector(".admin_item_relation_picker_suggestions_content");
        this.suggestions = [];
        this.picker = el.querySelector(".admin_item_relation_picker");
        this.pickerInput = this.picker.querySelector("input");
        this.pickerInput.addEventListener("input", () => {
            this.getSuggestions(this.pickerInput.value);
        });
        this.pickerInput.addEventListener("blur", () => {
            this.suggestionsEl.classList.add("hidden");
        });
        this.pickerInput.addEventListener("focus", () => {
            this.suggestionsEl.classList.remove("hidden");
            this.getSuggestions(this.pickerInput.value);
        });
        this.pickerInput.addEventListener("keydown", this.suggestionInput.bind(this));
        this.getData();
    }
    getData() {
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/preview/" + this.relationName + "/" + this.input.value, true);
        request.addEventListener("load", () => {
            this.progress.classList.add("hidden");
            if (request.status == 200) {
                this.showPreview(JSON.parse(request.response));
            }
            else {
                this.showSearch();
            }
        });
        request.send();
    }
    showPreview(data) {
        this.previewContainer.textContent = "";
        this.input.value = data.ID;
        var el = this.createPreview(data, true);
        this.previewContainer.appendChild(el);
        this.previewContainer.classList.remove("hidden");
        this.changeSection.classList.remove("hidden");
        this.picker.classList.add("hidden");
    }
    showSearch() {
        this.previewContainer.classList.add("hidden");
        this.changeSection.classList.add("hidden");
        this.picker.classList.remove("hidden");
        this.suggestions = [];
        this.suggestionsEl.innerText = "";
        this.pickerInput.value = "";
    }
    getSuggestions(q) {
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/search/" + this.relationName + "?q=" + encodeURIComponent(q), true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                if (q != this.pickerInput.value) {
                    return;
                }
                var data = JSON.parse(request.response);
                this.suggestions = data;
                this.suggestionsEl.innerText = "";
                for (var i = 0; i < data.length; i++) {
                    var item = data[i];
                    var el = this.createPreview(item, false);
                    el.classList.add("admin_item_relation_picker_suggestion");
                    el.setAttribute("data-position", i + "");
                    el.addEventListener("mousedown", this.suggestionClick.bind(this));
                    el.addEventListener("mouseenter", this.suggestionSelect.bind(this));
                    this.suggestionsEl.appendChild(el);
                }
            }
            else {
                console.log("Error while searching");
            }
        });
        request.send();
    }
    suggestionClick() {
        var selected = this.getSelected();
        if (selected >= 0) {
            this.showPreview(this.suggestions[selected]);
        }
    }
    suggestionSelect(e) {
        var target = e.currentTarget;
        var position = parseInt(target.getAttribute("data-position"));
        this.select(position);
    }
    getSelected() {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        return parseInt(selected.getAttribute("data-position"));
    }
    unselect() {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        selected.classList.remove(this.selectedClass);
        return parseInt(selected.getAttribute("data-position"));
    }
    select(i) {
        this.unselect();
        this.suggestionsEl.querySelectorAll(".admin_preview")[i].classList.add(this.selectedClass);
    }
    suggestionInput(e) {
        switch (e.keyCode) {
            case 13:
                this.suggestionClick();
                e.preventDefault();
                return true;
            case 38:
                var i = this.getSelected();
                if (i < 1) {
                    i = this.suggestions.length - 1;
                }
                else {
                    i = i - 1;
                }
                this.select(i);
                e.preventDefault();
                return false;
            case 40:
                var i = this.getSelected();
                if (i >= 0) {
                    i += 1;
                    i = i % this.suggestions.length;
                }
                else {
                    i = 0;
                }
                this.select(i);
                e.preventDefault();
                return false;
        }
    }
    createPreview(data, anchor) {
        var ret = document.createElement("div");
        if (anchor) {
            ret = document.createElement("a");
        }
        ret.classList.add("admin_preview");
        ret.setAttribute("href", data.URL);
        var image = document.createElement("div");
        image.classList.add("admin_preview_image");
        image.setAttribute("style", "background-image: url('" + data.Image + "');");
        var right = document.createElement("div");
        right.classList.add("admin_preview_right");
        var name = document.createElement("div");
        name.classList.add("admin_preview_name");
        name.textContent = data.Name;
        var description = document.createElement("description");
        description.classList.add("admin_preview_description");
        description.textContent = data.Description;
        ret.appendChild(image);
        right.appendChild(name);
        right.appendChild(description);
        ret.appendChild(right);
        return ret;
    }
}
function bindRelationsOLD() {
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
        request.addEventListener("load", () => {
            if (request.status >= 200 && request.status < 400) {
                var resp = JSON.parse(request.response);
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
        });
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
function bindPlacesView() {
    var els = document.querySelectorAll(".admin_item_view_place");
    for (var i = 0; i < els.length; i++) {
        new PlacesView(els[i]);
    }
}
class PlacesView {
    constructor(el) {
        var val = el.getAttribute("data-value");
        el.innerText = "";
        var coords = val.split(",");
        if (coords.length != 2) {
            el.innerText = "-";
            el.classList.remove("admin_item_view_place");
            return;
        }
        var position = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) };
        var zoom = 18;
        var map = new google.maps.Map(el, {
            center: position,
            zoom: zoom
        });
        var marker = new google.maps.Marker({
            position: position,
            map: map
        });
    }
}
function bindPlaces() {
    bindPlacesView();
    function bindPlace(el) {
        var mapEl = document.createElement("div");
        mapEl.classList.add("admin_place_map");
        el.appendChild(mapEl);
        var position = { lat: 0, lng: 0 };
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
        var searchInput = document.createElement("input");
        searchInput.classList.add("input", "input-placesearch");
        var searchBox = new google.maps.places.SearchBox(searchInput);
        map.controls[google.maps.ControlPosition.LEFT_TOP].push(searchInput);
        searchBox.addListener('places_changed', () => {
            var places = searchBox.getPlaces();
            if (places.length > 0) {
                map.fitBounds(places[0].geometry.viewport);
                marker.setPosition({ lat: places[0].geometry.location.lat(), lng: places[0].geometry.location.lng() });
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
function bindForm() {
    var els = document.querySelectorAll(".form_leavealert");
    for (var i = 0; i < els.length; i++) {
        new Form(els[i]);
    }
}
class Form {
    constructor(el) {
        this.dirty = false;
        el.addEventListener("submit", () => {
            this.dirty = false;
        });
        let els = el.querySelectorAll(".form_watcher");
        for (var i = 0; i < els.length; i++) {
            var input = els[i];
            input.addEventListener("input", () => {
                this.dirty = true;
            });
            input.addEventListener("change", () => {
                this.dirty = true;
            });
        }
        window.addEventListener("beforeunload", (e) => {
            if (this.dirty) {
                var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
                e.returnValue = confirmationMessage;
                return confirmationMessage;
            }
        });
    }
}
function bindFilter() {
    var els = document.querySelectorAll(".admin_filter_layout_date");
    for (var i = 0; i < els.length; i++) {
        new FilterDate(els[i]);
    }
}
class FilterDate {
    constructor(el) {
        this.hidden = el.querySelector(".admin_table_filter_item");
        this.from = el.querySelector(".admin_filter_layout_date_from");
        this.to = el.querySelector(".admin_filter_layout_date_to");
        this.from.addEventListener("input", this.changed.bind(this));
        this.to.addEventListener("input", this.changed.bind(this));
    }
    changed() {
        var val = "";
        if (this.from.value && this.to.value) {
            val = this.from.value + " - " + this.to.value;
        }
        this.hidden.value = val;
        var event = new Event('change');
        this.hidden.dispatchEvent(event);
    }
}
document.addEventListener("DOMContentLoaded", () => {
    bindStats();
    bindMarkdowns();
    bindTimestamps();
    bindRelations();
    bindImagePickers();
    bindLists();
    bindForm();
    bindImageViews();
    bindFlashMessages();
    bindFilter();
    bindScrolled();
});
function bindFlashMessages() {
    var messages = document.querySelectorAll(".flash_message");
    for (var i = 0; i < messages.length; i++) {
        var message = messages[i];
        message.addEventListener("click", (e) => {
            var target = e.target;
            if (target.classList.contains("flash_message_close")) {
                var current = e.currentTarget;
                current.classList.add("hidden");
            }
        });
    }
}
function bindScrolled() {
    var lastScrollPosition = 0;
    var header = document.querySelector(".admin_header");
    document.addEventListener("scroll", (event) => {
        if (document.body.clientWidth < 1100) {
            return;
        }
        var scrollPosition = window.scrollY;
        if (scrollPosition > 0) {
            header.classList.add("admin_header-scrolled");
        }
        else {
            header.classList.remove("admin_header-scrolled");
        }
        lastScrollPosition = scrollPosition;
    });
}
