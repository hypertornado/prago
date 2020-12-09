function bindEshopControl() {
    var el = document.querySelector(".eshop_control");
    if (el) {
        new EshopControl(<HTMLDivElement>el);
    }
}

//https://github.com/cozmo/jsQR
class EshopControl {

    canvas: HTMLCanvasElement;
    context: CanvasRenderingContext2D;
    video: HTMLVideoElement;

    constructor(el: HTMLDivElement) {
        this.canvas = el.querySelector(".eshop_control_canvas");
        this.video = el.querySelector(".eshop_control_video");
        this.context = this.canvas.getContext('2d');

        this.initCamera();
        this.captureImage();
    }

    initCamera() {
        navigator.mediaDevices.getUserMedia({ video: true, audio: false })
        .then(
            (stream: MediaStream) => {
                this.video.srcObject = stream;
                this.video.play();
            }
        )
    }

    captureImage() {
        var w = this.video.videoWidth;
        var h = this.video.videoHeight;

        this.canvas.width = w;
        this.canvas.height = h;

        this.context.drawImage(this.video ,0,0,w,h);

        if (w > 0 && h > 0) {
            var imageData = this.context.getImageData(0, 0, w, h);
            try {
                //@ts-ignore
                var code = jsQR(imageData.data, w, h);
                if (code) {
                    console.log(code.data);
                }
            } catch (error) {
                console.log(error);
            }
        }
        requestAnimationFrame(this.captureImage.bind(this));
    }

}