echo "package admin
const adminTemplates = \`" > admin_templates.go
cat templates/*.tmpl >> admin_templates.go
echo "\`
" >> admin_templates.go

echo "
const adminCSS = \`" >> admin_templates.go
cat static/normalize.css >> admin_templates.go
cat static/admin.css >> admin_templates.go
echo "\`
" >> admin_templates.go

echo "
const adminJS = \`" >> admin_templates.go
cat static/*.js >> admin_templates.go
echo "\`
" >> admin_templates.go


