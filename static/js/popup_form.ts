
class PopupForm extends Popup {

    private dataHandler: Function;

    constructor(path: string, dataHandler: Function) {
        super("⌛️");
        this.dataHandler = dataHandler;
        this.setCancelable();
        this.present();

        this.loadForm(path);
    }

    loadForm(path: string) {
        fetch(path)
        .then((response) => {
            if (response.ok) {
                return response.text();
            } else {
                this.unpresent();
                new Alert("Formulář nelze nahrát.");
            }
        })
        .then((textVal) => {
            this.wide();
            const parser = new DOMParser();
            const document = parser.parseFromString(textVal, "text/html");
            let formContainerEl = <HTMLDivElement>document.querySelector(".form_container");

            this.setContent(formContainerEl);
            new FormContainer(formContainerEl, this.okHandler.bind(this));
            this.setTitle(formContainerEl.getAttribute("data-form-name"));
        });
    }

    okHandler(data: any) {
        this.unpresent();
        this.dataHandler(data)
    }
}