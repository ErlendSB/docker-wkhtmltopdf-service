# wkhtmltopdf as a web service

A dockerized webservice written in [Go](https://golang.org/) that uses [wkhtmltopdf](http://wkhtmltopdf.org/) to convert HTML into documents (images or pdf files).

# Usage  
Do a HTTP POST to ip address (port 9090) with the following json:  
```
{
    "Url" : "http://sau.no",
    "Html" : "<html><head></head><body><p style='color:red;'>example</p></body></html>",
    "Output" :"png",
    "Options" : {"height" : "800", "width":"600"}
}

```
The Html field will be used if supplied. Omit it if you want to use the Url instead.


# License

This code is released under the [MIT License](http://opensource.org/licenses/MIT).
