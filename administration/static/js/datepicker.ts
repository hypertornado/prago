function bindDatePicker() {
  var dates = document.querySelectorAll(".form_input-date");
  for (var i = 0; i < dates.length; i++) {
    var dateEl = <HTMLInputElement>dates[i];
    new DatePicker(dateEl);
  }
}

class DatePicker {

  constructor(el: HTMLInputElement) { 
    var language = "cs" //el.getAttribute("data-language");
    var i18n = {
      previousMonth : 'Previous Month',
      nextMonth     : 'Next Month',
      months        : ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
      weekdays      : ['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'],
      weekdaysShort : ['Su','Mo','Tu','We','Th','Fr','Sa']
    };

    if (language == "de") {
      i18n = {
        previousMonth : 'Vorheriger Monat',
        nextMonth     : 'Nächsten Monat',
        months        : ["Januar", "Februar", "März", "April", "Kann", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"],
        weekdays      : ['Sonntag','Montag','Dienstag','Mittwoch','Donnerstag','Freitag','Samstag'],
        weekdaysShort : ['So','Mo','Di','Mi','Do','Fr','Sa']
      };
    }

    if (language == "ru") {
      var i18n = {
        previousMonth : 'Предыдущий месяц',
        nextMonth     : 'В следующем месяце',
        months        : ["Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"],
        weekdays      : ["Воскресенье", "Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота"],
        weekdaysShort : ['Во','По','Вт','Ср','Че','Пя','Су']
      };
    }

    if (language == "cs") {
      i18n = {
        previousMonth : 'Předchozí měsíc',
        nextMonth     : 'Další měsíc',
        months        : ["Leden", "Únor", "Březen", "Duben", "Květen", "Červen", "Červenec", "Srpen", "Září", "Říjen", "Listopad", "Prosinec"],
        weekdays      : ['Neděle','Pondělí','Úterý','Středa','Čtvrtek','Pátek','Sobota'],
        weekdaysShort : ['Ne','Po','Út','St','Čt','Pá','So']
      };
    }

    var self = this;

    //@ts-ignore
    var pd = new Pikaday({
      field: el,
      format: 'DD MM YYYY',
      //firstDay: 1,
      defaultDate: new Date(),
      i18n: i18n,
      onSelect: (date) => {
        el.value = pd.toString();
        //el.value = "2019-11-10";
        //console.log("SELECT");
      },
      toString: (date) => {
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
      }
    });
  }
}

/*
onSelect: (date) => {
  //$(el).dispatchEvent(new Event("changed_value"));
  dispatchEventHack(el, "changed_value");
  if (this.nights > 0 && this.related && this.related.value == "") {
    var newDate = addDays(date, this.nights);
    self.related.value = prettyDate(newDate);
  }
},
//@ts-ignore
toString(date: any, format: any): string {
    return prettyDate(date);
}

*/


function prettyDate(date: any): string {
  const day = date.getDate();
  const month = date.getMonth() + 1;
  const year = date.getFullYear();
  return `${day}. ${month}. ${year}`;
}

/*function addDays(theDate: any, days: number) {
  return new Date(theDate.getTime() + days*24*60*60*1000);
}*/