package admin
const CSS = `
/*! normalize.css v3.0.2 | MIT License | git.io/normalize */img,legend{border:0}legend,td,th{padding:0}html{font-family:sans-serif;-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}body{margin:0}article,aside,details,figcaption,figure,footer,header,hgroup,main,menu,nav,section,summary{display:block}audio,canvas,progress,video{display:inline-block;vertical-align:baseline}audio:not([controls]){display:none;height:0}[hidden],template{display:none}a{background-color:transparent}a:active,a:hover{outline:0}abbr[title]{border-bottom:1px dotted}b,optgroup,strong{font-weight:700}dfn{font-style:italic}h1{font-size:2em;margin:.67em 0}mark{background:#ff0;color:#000}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative;vertical-align:baseline}sup{top:-.5em}sub{bottom:-.25em}svg:not(:root){overflow:hidden}figure{margin:1em 40px}hr{-moz-box-sizing:content-box;box-sizing:content-box;height:0}pre,textarea{overflow:auto}code,kbd,pre,samp{font-family:monospace,monospace;font-size:1em}button,input,optgroup,select,textarea{color:inherit;font:inherit;margin:0}button{overflow:visible}button,select{text-transform:none}button,html input[type=button],input[type=reset],input[type=submit]{-webkit-appearance:button;cursor:pointer}button[disabled],html input[disabled]{cursor:default}button::-moz-focus-inner,input::-moz-focus-inner{border:0;padding:0}input{line-height:normal}input[type=checkbox],input[type=radio]{box-sizing:border-box;padding:0}input[type=number]::-webkit-inner-spin-button,input[type=number]::-webkit-outer-spin-button{height:auto}input[type=search]{-webkit-appearance:textfield;-moz-box-sizing:content-box;-webkit-box-sizing:content-box;box-sizing:content-box}input[type=search]::-webkit-search-cancel-button,input[type=search]::-webkit-search-decoration{-webkit-appearance:none}fieldset{border:1px solid silver;margin:0 2px;padding:.35em .625em .75em}table{border-collapse:collapse;border-spacing:0}html {
  box-sizing: border-box;
}
*, *:before, *:after {
  box-sizing: inherit;
}

html, body{
  height: 100%;
  font-family: 'Arial', sans-serif;
  font-size: 16px;
  line-height: 1.4em;
  color: #01354a;
}

p {
  margin: 0px;
  margin-bottom: 0.5em;
}


a {
  color: #01354a;
}

a:hover {
  text-decoration: none;
}

.admin_header {
  background: #fafafa;
  padding: 10px 10px;
  font-size: 1.0em;
  border-bottom: 1px solid #ddd;
  color: black;
}

.admin_header h1 {
  display: inline-block;
  font-size: 1.1em;
  margin: 0px 5px;
}

.admin_header_list {
  margin: 0px;
  padding: 0px;
}
.admin_header_list {
  display: inline-block;
}
.admin_header_list li {
  display: inline-block;
  padding: 0px 2px;
}
.admin_header_list-right {
  float: right;
}
.admin_content {
  max-width: 600px;
  padding: 10px 10px;
  margin: 0 auto;
}
.admin_footer {
  background: #01354a;
  color: white;
  text-align: center;
  margin: 0 auto;
  padding: 20px 10px;
}
.admin_table {
  width: 100%;
}
.admin_table thead {
  font-weight: bold;
}
.admin_table td {
  padding: 3px 5px;
  border: 1px solid #ccc;
}



.btn {
  background: #dd2e4f;
  display: inline-block;
  padding: 5px 10px;
  text-decoration: none;
  color: white;
  font-weight: bold;
  border-radius: 5px;
  font-family: 'Catamaran', sans-serif;
  border: none;
  outline: none;
  border: 1px solid rgba(0,0,0,0);
  line-height: 1em;
}

.btn:focus {
  border: 1px solid #009ee0;
}


.form {
  background-color: #fafafa;
  padding: 5px 20px 20px 20px;
  margin: 10px auto;
  border-radius: 3px;
  box-shadow: 0px 0px 2px rgba(0,0,0,0.1);
}

.form_errors_error {
  border: 1px solid #dd2e4f;
  color: #dd2e4f;
  padding: 5px;
  text-align: center;
  border-radius: 3px;
}

.form_label {
  display: block;
  margin: 20px 0px;
}

.form_label-required input {
  border-width: 2px;
}

.form_label-required .form_label_text {
  font-weight: bold;
}

.form_label-errors {
  color: #dd2e4f;
}

.form_label-errors input, .form_label-errors textarea {
  border-color: #dd2e4f !important;
}

.form_label_errors {
  font-size: 0.8em;
}

.form_label_text-checkbox {
  padding: 0px 5px;
}

.input {
  display: inline-block;
  padding: 6px 5px;
  font-size: 13px;
  line-height: 20px;
  color: #333;
  width: 100%;
  vertical-align: middle;
  border: 1px solid #d5d5d5;
  border-radius: 3px;
  outline: none;
}

.input[readonly], .textarea[readonly] {
  border: 1px solid #d5d5d5 !important;
  background: #fafafa !important;
}

.input:focus {
  border-color: #009ee0;
}

.textarea {
  min-height: 150px;
}

.admin_table {
  width: 100%;
  margin-bottom: 30px;
}
.admin_table thead {
  font-weight: bold;
}
.admin_table td {
  padding: 3px 5px;
  border: 1px solid #ccc;
}

.admin_header_item-active {
  //background-color: white;
  //border: 2px solid #009ee0;
}

.admin_header_item-active a {
  font-weight: bold;
}

.flash {
  text-align: center;
  padding: 5px;
  background: #009ee0;
}

.flash_message {
  color: white;
  font-weight: bold;
  display: inline-block;
}
`

