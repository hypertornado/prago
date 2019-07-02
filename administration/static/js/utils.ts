function DOMinsertChildAtIndex(parent: HTMLElement, child: HTMLElement, index: number) {
  if (index >= parent.children.length) {
    parent.appendChild(child);
  } else {
    parent.insertBefore(child, parent.children[index]);
  }
}

function encodeParams(data: any) {
  var ret = "";
  for (var k in data) {
    if (!data[k]) {
      continue;
    }
    if (ret != "") {
      ret += "&";
    }
    ret += encodeURIComponent(k) + "=" + encodeURIComponent(data[k]);
  }
  if (ret != "") {
    ret = "?" + ret;
  }
  return ret;
}