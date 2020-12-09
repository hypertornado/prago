echo "package administration
const adminTemplates = \`" > admin_templates.go
cat templates/*.tmpl >> admin_templates.go
echo "\`
" >> admin_templates.go

echo "
const adminCSS = \`" >> admin_templates.go
cat static/public/admin/_static/admin.css >> admin_templates.go
echo "\`
" >> admin_templates.go

echo "
const adminJS = \`" >> admin_templates.go
cat static/public/admin/_static/admin.js >> admin_templates.go
echo "\`
" >> admin_templates.go

echo "
const pikadayJS = \`" >> admin_templates.go
cat static/public/admin/_static/pikaday.js >> admin_templates.go
echo "\`
" >> admin_templates.go


