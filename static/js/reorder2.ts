/**
 * reorder.ts — moderní drag-and-drop řazení prvků (Prago admin, globální scope).
 *
 *  • Pointer Events (myš / dotyk / pero)
 *  • FLIP animace odsunutých sousedů
 *  • živé přeřazení přímo v DOM
 *  • na taženém prvku je po celou dobu tahu CSS třída "reordering"
 *  • stabilní řazení pro svislý seznam, vodorovný i flex-wrap mřížku
 *  • idempotentní: opakované volání na stejný root starou instanci nejdřív zruší
 *  • zvládá <a>/<img> uvnitř položek (vypne nativní drag + spolkne click po tahu)
 *
 * Pozn.: přidání/odebrání potomků re-inicializaci nepotřebuje — listener je
 * delegovaný na rootu a MutationObserver ošetří touch-action nových položek.
 *
 * Bez `export` — soubor je skript, ne ES modul, aby šel slepit přes `outFile`.
 *
 * Použití:
 *   const handle = makeReorderable(listEl, (d) => uloz(d.order));
 *   const handle = makeReorderable(listEl, { onReorder, handle: ".grip" });
 *   handle.destroy();
 */

interface ReorderDetail {
  item: HTMLElement;
  from: number;
  to: number;
  order: HTMLElement[];
}

interface ReorderOptions {
  onReorder?: (detail: ReorderDetail) => void;
  threshold?: number;
  animationMs?: number;
  handle?: string;
  draggingClass?: string;
}

interface ReorderHandle {
  destroy(): void;
}

type ReorderCallback = (detail: ReorderDetail) => void;

const reorderInstances = new WeakMap<HTMLElement, ReorderHandle>();

function makeReorderable(
  container: HTMLElement,
  optionsOrCallback: ReorderOptions | ReorderCallback,
): ReorderHandle {
  const existing = reorderInstances.get(container);
  if (existing) existing.destroy();

  const opts: ReorderOptions =
    typeof optionsOrCallback === "function"
      ? { onReorder: optionsOrCallback }
      : (optionsOrCallback || {});

  const onReorder = opts.onReorder;
  const threshold = opts.threshold != null ? opts.threshold : 5;
  const animationMs = opts.animationMs != null ? opts.animationMs : 180;
  const handleSel = opts.handle || null;
  const draggingClass = opts.draggingClass || "reordering";

  let candidate: HTMLElement | null = null;
  let dragEl: HTMLElement | null = null;
  let dragging = false;
  let pointerId = -1;

  let startX = 0, startY = 0;
  let pointerX = 0, pointerY = 0;
  let grabX = 0, grabY = 0;
  let baseX = 0, baseY = 0;
  let startIndex = -1;

  let saved: { transition: string; zIndex: string; position: string } | null = null;

  function children(): HTMLElement[] {
    return Array.prototype.slice.call(container.children) as HTMLElement[];
  }

  function indexOf(el: HTMLElement): number {
    return children().indexOf(el);
  }

  function itemFromEvent(e: PointerEvent): HTMLElement | null {
    let node = e.target as HTMLElement | null;
    while (node && node.parentElement !== container) node = node.parentElement;
    return node && node.parentElement === container ? node : null;
  }

  function layoutMode(): "vertical" | "horizontal" | "grid" {
    const kids = children();
    if (kids.length < 2) return "vertical";
    const r0 = kids[0].getBoundingClientRect();
    let sameRow = 0;
    let stacked = false;
    for (let i = 1; i < kids.length; i++) {
      const r = kids[i].getBoundingClientRect();
      if (Math.abs(r.top - r0.top) < Math.max(r.height, r0.height) / 2) sameRow++;
      else stacked = true;
    }
    if (sameRow > 0 && stacked) return "grid";
    if (sameRow > 0) return "horizontal";
    return "vertical";
  }

  function targetIndex(): number {
    const mode = layoutMode();
    let idx = 0;
    for (const s of children()) {
      if (s === dragEl) continue;
      const r = s.getBoundingClientRect();
      let after: boolean;
      if (mode === "vertical") {
        after = pointerY > r.top + r.height / 2;
      } else if (mode === "horizontal") {
        after = pointerX > r.left + r.width / 2;
      } else {
        after = pointerY > r.bottom ||
          (pointerY >= r.top && pointerX > r.left + r.width / 2);
      }
      if (after) idx++;
    }
    return idx;
  }

  // <a>/<img> jsou ve výchozím stavu draggable a jejich nativní drag by zrušil
  // naše pointer eventy — proto ho na rootu vypneme
  function onDragStart(e: Event): void {
    e.preventDefault();
  }

  // po skutečném tahu spolkni následující click, ať odkaz uvnitř nenaviguje
  function suppressNextClick(): void {
    function handler(e: Event): void {
      e.stopPropagation();
      e.preventDefault();
      container.removeEventListener("click", handler, true);
    }
    container.addEventListener("click", handler, true);
    window.setTimeout(function () {
      container.removeEventListener("click", handler, true);
    }, 50);
  }

  function onPointerDown(e: PointerEvent): void {
    if (dragging) return;
    if (e.pointerType === "mouse" && e.button !== 0) return;
    const item = itemFromEvent(e);
    if (!item) return;
    if (handleSel && !(e.target as Element).closest(handleSel)) return;

    candidate = item;
    pointerId = e.pointerId;
    startX = pointerX = e.clientX;
    startY = pointerY = e.clientY;
    const r = item.getBoundingClientRect();
    grabX = e.clientX - r.left;
    grabY = e.clientY - r.top;

    window.addEventListener("pointermove", onPointerMove, { passive: false });
    window.addEventListener("pointerup", onPointerUp);
    window.addEventListener("pointercancel", onPointerUp);
  }

  function onPointerMove(e: PointerEvent): void {
    if (e.pointerId !== pointerId) return;
    pointerX = e.clientX;
    pointerY = e.clientY;

    if (!dragging) {
      if (Math.hypot(pointerX - startX, pointerY - startY) < threshold) return;
      beginDrag();
    }
    e.preventDefault();
    applyTransform();
    maybeSwap();
  }

  function onPointerUp(e: PointerEvent): void {
    if (e.pointerId !== pointerId) return;
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
    window.removeEventListener("pointercancel", onPointerUp);
    if (dragging) {
      suppressNextClick();
      endDrag();
    }
    candidate = null;
    pointerId = -1;
  }

  function beginDrag(): void {
    dragging = true;
    dragEl = candidate;
    if (!dragEl) return;
    startIndex = indexOf(dragEl);

    const r = dragEl.getBoundingClientRect();
    baseX = r.left;
    baseY = r.top;

    saved = {
      transition: dragEl.style.transition,
      zIndex: dragEl.style.zIndex,
      position: dragEl.style.position,
    };
    if (getComputedStyle(dragEl).position === "static") {
      dragEl.style.position = "relative";
    }
    dragEl.style.zIndex = "1000";
    dragEl.style.transition = "none";
    dragEl.classList.add(draggingClass);
    document.body.style.userSelect = "none";
  }

  function applyTransform(): void {
    if (!dragEl) return;
    const tx = pointerX - grabX - baseX;
    const ty = pointerY - grabY - baseY;
    dragEl.style.transform = "translate(" + tx + "px, " + ty + "px)";
  }

  function maybeSwap(): void {
    if (!dragEl) return;
    const sibs = children().filter(function (s) { return s !== dragEl; });
    const idx = targetIndex();
    const ref = idx < sibs.length ? sibs[idx] : null;

    if (ref === null) {
      if (container.lastElementChild === dragEl) return;
    } else {
      if (dragEl.nextElementSibling === ref) return;
    }

    flipReorder(function () {
      if (ref) container.insertBefore(dragEl as HTMLElement, ref);
      else container.appendChild(dragEl as HTMLElement);
    });
  }

  function flipReorder(mutate: () => void): void {
    if (!dragEl) return;
    const moving = children().filter(function (k) { return k !== dragEl; });
    const first = new Map<HTMLElement, DOMRect>();
    for (const k of moving) first.set(k, k.getBoundingClientRect());

    mutate();

    const keep = dragEl.style.transform;
    dragEl.style.transform = "";
    const b = dragEl.getBoundingClientRect();
    baseX = b.left;
    baseY = b.top;
    dragEl.style.transform = keep;
    applyTransform();

    first.forEach(function (f, k) {
      const l = k.getBoundingClientRect();
      const dx = f.left - l.left;
      const dy = f.top - l.top;
      if (!dx && !dy) return;
      k.style.transition = "none";
      k.style.transform = "translate(" + dx + "px, " + dy + "px)";
    });
    requestAnimationFrame(function () {
      first.forEach(function (_f, k) {
        if (k.style.transform === "") return;
        k.style.transition = "transform " + animationMs + "ms ease";
        k.style.transform = "";
      });
    });
  }

  function restore(el: HTMLElement): void {
    el.style.transition = saved ? saved.transition : "";
    el.style.transform = "";
    el.style.zIndex = saved ? saved.zIndex : "";
    el.style.position = saved ? saved.position : "";
    el.classList.remove(draggingClass);
  }

  function endDrag(): void {
    if (!dragEl) return;
    const el = dragEl;
    const finalIndex = indexOf(el);

    let done = false;
    function cleanup(): void {
      if (done) return;
      done = true;
      restore(el);
      el.removeEventListener("transitionend", onEnd);
    }
    function onEnd(ev: TransitionEvent): void {
      if (ev.propertyName === "transform") cleanup();
    }

    el.style.transition = "transform " + animationMs + "ms ease";
    el.style.transform = "translate(0px, 0px)";
    el.addEventListener("transitionend", onEnd);
    window.setTimeout(cleanup, animationMs + 60);

    document.body.style.userSelect = "";
    for (const k of children()) {
      if (k === el) continue;
      k.style.transition = "";
      k.style.transform = "";
    }

    if (onReorder && finalIndex !== startIndex) {
      onReorder({ item: el, from: startIndex, to: finalIndex, order: children() });
    }

    dragging = false;
    dragEl = null;
    saved = null;
  }

  const observer = new MutationObserver(function (muts) {
    for (const m of muts) {
      m.addedNodes.forEach(function (n) {
        if (n instanceof HTMLElement) n.style.touchAction = "none";
      });
    }
  });

  for (const k of children()) k.style.touchAction = "none";
  container.addEventListener("pointerdown", onPointerDown);
  container.addEventListener("dragstart", onDragStart);
  observer.observe(container, { childList: true });

  const api: ReorderHandle = {
    destroy: function () {
      container.removeEventListener("pointerdown", onPointerDown);
      container.removeEventListener("dragstart", onDragStart);
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", onPointerUp);
      window.removeEventListener("pointercancel", onPointerUp);
      observer.disconnect();
      for (const k of children()) k.style.touchAction = "";
      if (dragEl) {
        restore(dragEl);
        document.body.style.userSelect = "";
        dragging = false;
        dragEl = null;
      }
      if (reorderInstances.get(container) === api) reorderInstances.delete(container);
    },
  };

  reorderInstances.set(container, api);
  return api;
}
