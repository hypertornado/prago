class FormDateRange {

    constructor(el: HTMLDivElement) {
        let fromInput = <HTMLInputElement>el.querySelector(".form_input-from");
        let toInput = <HTMLInputElement>el.querySelector(".form_input-to");

        let fromCalendar = <HTMLButtonElement>el.querySelector(".form_daterange_calendar-from");
        let toCalendar = <HTMLButtonElement>el.querySelector(".form_daterange_calendar-to");

        this.bindCalendarAndInput(fromInput, fromCalendar);
        this.bindCalendarAndInput(toInput, toCalendar);

        el.querySelector(".form_daterange_more").addEventListener("mousedown", (e: Event) => {
            new PopupForm("/admin/_dateranges", (data: any) => {
                let dates = data.Data.split("_");
                fromInput.value = dates[0];
                toInput.value = dates[1];
            });
        });
    }

    bindCalendarAndInput(input: HTMLInputElement, btn: HTMLButtonElement) {
        btn.addEventListener("mousedown", (e: Event) => {
            input.showPicker();
            input.focus();
            e.preventDefault();
        });
    }

}