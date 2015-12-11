/*
importlogs will import apache/nginx style logs into an elasticsearch database.

  usage: importlogs [flags] file1.gz [file2.gz...]
        Imports gzipped log files.

  flags:

  -clean
        clean the index before adding content
  -e
        continnue to next file if an error occurs
  -elastic string
        url to elasticseach server (http) (default "http://127.0.0.1:9200")
        Overriden if environment variable "ELASTICSEARCH_PORT_9200_TCP" is set.
  -format string
        Log format (default "$remote_addr - - [$time_local] \"$method $uri $protocol\" $status $size")
  -geodb string
        MaxMind GeoLite2 or GeoIP2 mmdb database to translate IP to location
  -timeformat string
        Time format in Go time.Parse format. (default "02/Jan/2006:15:04:05 -0700")
        See https://golang.org/pkg/time/#Parse for more information on the format.

Specifying custom log formatting

You can specify a custom log parsing format. The default parse format is matching content at http://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html

For more information on the format, see https://github.com/satyrius/gonx#format

The following fields are recognized and extracted:
      - "remote_addr"
      Remote address of the requestor. This can be either a hostname, or an IP address.

	- "uri"
      The requested URI without hostname.

	- "method"
      The request method.

	- "protocol"
      The request protocol.

	- "time_local"
      The local server time.
      The time must be parseable with the "-timeformat".

	- "status"
      The server status reply code.

	- "size"
      Size of the reply in bytes. Can be '-' on bodyless replies.
*/
package main
