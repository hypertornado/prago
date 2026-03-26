declare var Hls: any;

class VideoView {
  el: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    const url = el.getAttribute("data-videos");
    const videoEl = el.querySelector<HTMLVideoElement>(".videoview_video");
    const progressEl = el.querySelector<HTMLProgressElement>(".progress");

    if (!url || !videoEl) {
      return;
    }

    videoEl.controls = true;

    if (typeof Hls !== "undefined" && Hls.isSupported()) {
      const hls = new Hls();
      hls.loadSource(url);
      hls.attachMedia(videoEl);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        videoEl.classList.remove("hidden");
        if (progressEl) {
          progressEl.classList.add("hidden");
        }
      });
    } else if (videoEl.canPlayType("application/vnd.apple.mpegurl")) {
      videoEl.src = url;
      videoEl.classList.remove("hidden");
      if (progressEl) {
        progressEl.classList.add("hidden");
      }
    }
  }
}
