echo "package admin
const adminTemplates = \`" > admin_templates.go
cat templates/*.tmpl >> admin_templates.go
echo "\`
" >> admin_templates.go


