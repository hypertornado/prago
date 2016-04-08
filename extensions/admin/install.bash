echo "package admin
const TEMPLATES = \`" > admin_templates.go
cat templates/*.tmpl >> admin_templates.go
echo "\`
" >> admin_templates.go



echo "package admin
const CSS = \`" > admin_css.go
cat css/normalize.css >> admin_css.go
cat css/admin.css >> admin_css.go
echo "\`
" >> admin_css.go