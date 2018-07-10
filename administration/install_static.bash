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
const chartJS = \`" >> admin_templates.go
cat static/public/admin/_static/Chart.min.js >> admin_templates.go
echo "\`
" >> admin_templates.go


