class DatePicker {
  constructor(el: HTMLInputElement) {
    var language = "cs";
    language = document.getElementsByTagName("html")[0].lang;
    var i18n = {
      previousMonth: "Previous Month",
      nextMonth: "Next Month",
      months: [
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
      ],
      weekdays: [
        "Sunday",
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
      ],
      weekdaysShort: ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"],
    };

    if (language == "de") {
      i18n = {
        previousMonth: "Vorheriger Monat",
        nextMonth: "Nächsten Monat",
        months: [
          "Januar",
          "Februar",
          "März",
          "April",
          "Kann",
          "Juni",
          "Juli",
          "August",
          "September",
          "Oktober",
          "November",
          "Dezember",
        ],
        weekdays: [
          "Sonntag",
          "Montag",
          "Dienstag",
          "Mittwoch",
          "Donnerstag",
          "Freitag",
          "Samstag",
        ],
        weekdaysShort: ["So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"],
      };
    }

    if (language == "ru") {
      var i18n = {
        previousMonth: "Предыдущий месяц",
        nextMonth: "В следующем месяце",
        months: [
          "Январь",
          "Февраль",
          "Март",
          "Апрель",
          "Май",
          "Июнь",
          "Июль",
          "Август",
          "Сентябрь",
          "Октябрь",
          "Ноябрь",
          "Декабрь",
        ],
        weekdays: [
          "Воскресенье",
          "Понедельник",
          "Вторник",
          "Среда",
          "Четверг",
          "Пятница",
          "Суббота",
        ],
        weekdaysShort: ["Во", "По", "Вт", "Ср", "Че", "Пя", "Су"],
      };
    }

    if (language == "cs") {
      i18n = {
        previousMonth: "Předchozí měsíc",
        nextMonth: "Další měsíc",
        months: [
          "Leden",
          "Únor",
          "Březen",
          "Duben",
          "Květen",
          "Červen",
          "Červenec",
          "Srpen",
          "Září",
          "Říjen",
          "Listopad",
          "Prosinec",
        ],
        weekdays: [
          "Neděle",
          "Pondělí",
          "Úterý",
          "Středa",
          "Čtvrtek",
          "Pátek",
          "Sobota",
        ],
        weekdaysShort: ["Ne", "Po", "Út", "St", "Čt", "Pá", "So"],
      };
    }

    var self = this;
    var firstDay = 0;
    if (language == "cs") {
      firstDay = 1;
    }

    //@ts-ignore
    var pd = new Pikaday({
      field: el,
      setDefaultDate: false,
      i18n: i18n,
      firstDay: firstDay,
      onSelect: (date: any) => {
        el.value = pd.toString();
      },
      toString: (date: any) => {
        const day = date.getDate();
        var dayStr = "" + day;
        if (day < 10) {
          dayStr = "0" + dayStr;
        }
        const month = date.getMonth() + 1;
        var monthStr = "" + month;
        if (month < 10) {
          monthStr = "0" + monthStr;
        }
        const year = date.getFullYear();
        var ret = `${year}-${monthStr}-${dayStr}`;
        return ret;
      },
    });
  }
}

function prettyDate(date: any): string {
  const day = date.getDate();
  const month = date.getMonth() + 1;
  const year = date.getFullYear();
  return `${day}. ${month}. ${year}`;
}
