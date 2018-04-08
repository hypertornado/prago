package extensions

import (
	"fmt"
	"github.com/hypertornado/prago"
)

//Sitemap renders sites sitemap xml file
func Sitemap(request prago.Request, urls []string) {
	request.Response().Header().Set("Content-Type", "text/xml")
	request.Response().WriteHeader(200)
	prev := `<?xml version="1.0" encoding="UTF-8"?><urlset
      xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
      xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
      xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9
      http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">`
	request.Response().Write([]byte(prev))

	for _, v := range urls {
		url := fmt.Sprintf("<url><loc>%s</loc></url>", v)
		request.Response().Write([]byte(url))
	}

	after := `</urlset>`
	request.Response().Write([]byte(after))
}
