package prago

const defaultNewsletterTemplate = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"> 
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>{{.title}}</title>

<style type="text/css">
  body {
    background-color: #edfaff;
    font-style: normal;
    font-size: 15px;
    line-height: 1.5em;
    font-weight: 400;
    color: #01354a;
    font-family: Arial, sans-serif !important;
  }

  .middle {
    background-color: #fff;
    padding: 10px;
  }

  img {
    max-width: 100%;
  }

  a {
    color: #009ee0;
  }

  a:hover {
    text-decoration: none;
  }

  .unsubscribe {
    color: #999;
    display: block;
    text-align: center;
    font-size: 11px;
  }

  h1 {
    text-align: center;
    line-height: 1.2em;
  }

  hr {
    border-top: 1px solid #009ee0;
    border-bottom: none;
  }

  table {
    margin-top: 5px;
  }

  td {
    padding: 0px 5px;
    vertical-align: top;
  }

</style>

</head>
<body>

<table width="100%" border="0" cellspacing="0" cellpadding="0"><tr><td width="100%" align="center">
  <table width="450" border="0" align="center" cellpadding="0" cellspacing="0">
    <tr><td width="450" align="left" class="middle">
          <div class="middle_header">
            <a href="{{.baseUrl}}/?utm_source=newsletter&utm_medium=prago&utm_campaign={{.id}}">{{.site}}</a>
            <h1>{{.title}}</h1>
          </div>
          {{.content}}

          {{$baseUrl := .baseUrl}}
          {{if .sections}}
            {{range $section := .sections}}
              <!-- start layout-2 section -->
              <table width="572" border="0" cellspacing="0" cellpadding="0" align="center" style="border:1px solid; border-color:#eeeeee; background-color: #ffffff;" class="container">
                <tr>      
                  <td align="center" valign="top">
                    <table width="100%" border="0" cellspacing="0" cellpadding="0" align="center"  >          
                      <!-- start space -->
                      <tr>
                        <td valign="top" height="9" >
                        </td>
                      </tr>
                      <!-- end space -->
                      <tr>
                        <td valign="top" align="center" >
                          
                          <!-- start space width  -->
                          <table width="1" border="0" cellspacing="0" cellpadding="0" align="left" class="remove">
                            <tr>
                              <td align="center" valign="top">
                                <p style="padding-left:5px;mso-table-lspace:0;mso-table-rspace:0;"><img src="{{$baseUrl}}/newsletter/spacer.gif" alt="" style="border:none; display:block !important;" width="4" /></p>
                              </td>
                            </tr>
                          </table>
                          <!-- end space width  -->

                          <table width="222" border="0" cellspacing="0" cellpadding="0" align="left" class="container">
                            <tr>
                              <td align="center" valign="top">
                                <img src="{{$section.Image}}" alt="image-1" style="vertical-align: top;" width="222" />
                              </td>
                            </tr>
                          </table>

                          <!-- start space width  -->
                          <table width="1" border="0" cellspacing="0" cellpadding="0" align="left" class="remove">
                            <tr>
                              <td align="center" valign="top">
                                <p style="padding-left:5px;mso-table-lspace:0;mso-table-rspace:0;"><img src="{{$baseUrl}}/newsletter/spacer.gif" alt="" style="border:none; display:block !important;" width="12" /></p>
                              </td>
                            </tr>
                          </table>
                          <!-- end space width  -->

                          <table width="308" border="0" cellspacing="0" cellpadding="0" align="left" class="container">
                            <!-- start space -->
                            <tr>
                              <td valign="top" height="30" >
                              </td>
                            </tr>
                            <!-- end space -->
                            <tr>
                              <td align="left" valign="top" class="p-10-align-c">
                                <h2 style=" font-size: 24px; line-height: 30px; font-family:Trebuchet MS, sans-serif; color:#787778; font-weight: bold; padding: 0; margin: 0;">{{$section.Name}}</h2>
                              </td>
                            </tr>
                            <!-- start space -->
                            <tr>
                              <td valign="top" height="6" >
                              </td>
                            </tr>
                            <!-- end space -->
                            <tr>
                              <td align="left" valign="top" style="font-size: 13px; line-height: 20px; font-family:Arial, Helvetica, sans-serif; color:#858485; font-weight: normal;" class="p-10-align-c">
                                {{$section.Text}}
                              </td>
                            </tr>
                            <!-- start space -->
                            <tr>
                              <td valign="top" height="7" >
                              </td>
                            </tr>
                            <!-- end space -->
                            <tr>
                              <td align="left" valign="top" class="p-10-align-c">
                                <a href="{{$section.URL}}" style="font-size: 14px; color: #d44e48; line-height: 20px; font-family:Trebuchet MS, sans-serif; font-weight: bold; font-style: italic; text-decoration: none;" >{{$section.Button}} >></a>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                      <!-- start space -->
                      <tr>
                        <td valign="top" height="10" >
                        </td>
                      </tr>
                      <!-- end space -->
                    </table>
                  </td>
                </tr>
              </table>
              <!-- end layout-2 section -->

              <!-- start space -->
              <table width="100%" border="0" cellspacing="0" cellpadding="0" align="center">
                <tr>
                  <td valign="top" height="29" >
                  </td>
                </tr>
              </table>
              <!-- end space -->
              
            {{end}}
          {{end}}

          <a href="{{.unsubscribe}}" class="unsubscribe">Odhlásit odběr novinek</a>
    </td></tr>
  </table>
</td></tr></table>

</body>
</html>
`
