
function DOMinsertChildAtIndex(parent: HTMLElement, child: HTMLElement, index: number) {
  if (index >= parent.children.length) {
    parent.appendChild(child);
  } else {
    parent.insertBefore(child, parent.children[index]);
  }
}