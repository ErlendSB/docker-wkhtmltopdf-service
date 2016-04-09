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

# License

This code is released under the [MIT License](http://opensource.org/licenses/MIT).
