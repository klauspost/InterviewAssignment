# apache log importer/analyzer
[![Build Status](https://travis-ci.org/klauspost/InterviewAssignment.svg?branch=master)](https://travis-ci.org/klauspost/InterviewAssignment)

Log parser for apache/nginx style logs.

This package will import logs to Elasticsearch to enable data visualization.

See an [example visualization](http://tinyurl.com/pduvdxl).

# installation
This package uses [Go 1.5's vendor experiment](https://medium.com/@freeformz/go-1-5-s-vendor-experiment-fd3e830f52c3#.kl6i4y54k). 

In the same directory as this file, execute:

```bash
export GO15VENDOREXPERIMENT=1
go install ./cmd/importlogs
```

`importlogs` requires an [Elasticsearch server](https://www.elastic.co/) to store the data. 
The library assumes Elasticsearch v2.1 in a default configuration installed on the local machine.


# usage

To import one or more files, execute:

```bash
importlogs [flags] file1.gz [file2.gz...]
```

This will import all specified files. These are assumed to be gzipped apache/nginx style logs, though you can specify custom formats.

The custom flags are: 
        
| Flag                | Explanation                                                                                                                                             |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-clean`            | clean the index before adding content                                                                                                                   |
| `-e`                | continue to next file if an error occurs                                                                                                                |
| `-elastic=URL`      | url to elasticseach server (http) (default `"http://127.0.0.1:9200"`). Overriden if environment variable "ELASTICSEARCH_PORT_9200_TCP" is set           |
| `-format="..."`     | Log format (default `"$remote_addr - - [$time_local] \"$method $uri $protocol\" $status $size"`). See Custom log formatting below.                      |
| `-geodb="path"`     | Path to MaxMind GeoLite2 or GeoIP2 mmdb database to translate IP to location.                                                                           |
| `-timeformat="..."` | time format in Go time.Parse format. (default `"02/Jan/2006:15:04:05 -0700"`). See [time.Parse](https://golang.org/pkg/time/#Parse) for more information on the format. |
| `-test`             | write json representation of requests to stdout. This can be used to test a filter, and observe enrichment data. Note that the JSON representation is unordered. |

For dockerized deployment, `importlogs` will read the `ELASTICSEARCH_PORT_9200_TCP` environment variable, which can be used for linking the [official docker image](https://hub.docker.com/_/elasticsearch/) automatically.

## Custom log formatting

You can specify a custom log parsing format. The default parse format is matching content at http://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html

For more information on the format, see https://github.com/satyrius/gonx#format

The following fields are recognized and extracted:

 * `remote_addr`: Remote address of the requestor. This can be either a hostname, or an IP address.
 * `uri`: The requested URI without hostname.
 * `method`: The request method. `GET`, `PUT`, etc.
 * `protocol`: The request protocol. `HTTP/1.1`, etc.
 * `time_local`:  The local server time. The time must be parseable with the `-timeformat`.
 * `status`: The server status reply code.
 * `size`: Size of the reply in bytes. Can be '-' on bodyless replies.

## elasticsearch model

Data is stored in `requests-yyyy.mm.dd` indexes, with one index per day, similar to Logstash/Heka and similar tools.

When possible, the data is enriched with geolocation, country, local time.


# postmortem

## design considerations

I settled pretty early on Elasticsearch/Kibana for data visualization. This combination is commonly used along with Logstash or a similar program to visualize logs and metrics.

Given my self-imposed, time limit of 3 working days, my impression was that I would get the best result by focusing on import parsing and data enrichment, and avoid writing a server with a complete HTML/javascript setup.
If this was for a client, I found it more valuable to be able to have a dynamic presentation, where actual data mining could be seen, compared to a few static graphs, which a "homebrewed" solution would likely have ended up with.   

**Elasticsearch/Kibana Pros**
 * Quick prototyping.
 * Proven stack combination.
 * Large existing userbase.
 * Extremely easy and flexible visulization (Kibana), can adjust to business parameters.
 * Aggregations can be adjusted on the fly (not used in this project).

**Elasticsearch/Kibana Cons**
 * ES is hard to manage in production (in short: developers love it, ops hate it).
 * Extending Kibana can be challenging.
 * Scalability can become an issue if daily request are an order or two of magnitude larger.
 * Metrics are better handled by a time-series DB like InfluxDB.

Given our talks about Cassandra it would have been interesting to implement it with that backend. It could easily be added as a storage backend, but from a business perspective I could not justify that.

## what went right?

### early stack choice

The early stack choice meant that it was fairly quick to get up an easy testing/prototyping setup. 
This allowed very quick iteration times, and you could visualize the end result seconds after implementation.

In this particular case, I had little doubt that the combination used could fulfill the demands of the assignment.
However that is of course not always a given, and a lot of time was "saved" in having an early stack set up, that I had confidence in for the task.

### third party components

For log parsing, I chose to base it on an existing third party component. It saved a lot of time, and made the implementation fast and very adaptable, since the configuration can be fully exposed to the user.
This mainly ensures that future changes that may come up, are a lot easier to handle. 
It is tempting to "roll your own" implementation of a simple task like splitting a log line into components, but you often underestimate the value of a specialized, well-tested package. 

For Elasticsearch, I tried a new client library, and it was a huge upgrade from the previous library I had tried, which always left a lot of adjustments. 
Elasticsearch is very complex due to its flexible nature, and this library leverages a lot of the headaches of this. The library is without a doubt the best Go client at the moment. 

### ease of use and speed

The importer performs rather well (imports 3-7k records/second). The main bottleneck is Elasticsearch writes. Since we are dealing with log files that aren't realtime, 

The current implementation is easy to use, and can be customized to a large extent. Interrupted imports can be restarted without any consistency issues. 

Aggregating more values requires code changes. The alternative is to have aggregation in the elasticsearch database. 
However, it was a design choise that the ES server should require no configuration to work for easy deployment. 

It would be possible to create a separate tool that maintains in-DB aggregations (as ES aggregation scripts), or calculates aggregation separately to separate indexes.    

## what went wrong?

### vendoring

For this project I decided to try the Go 1.5 /vendor experiment to gain experience with that, and my initial impression isn't entirely positive.

First of all, keeping dependencies synchronized requires manual work. `godep` allows "automatic" vendoring management, but there are problems which keeps it from being reliable. Most obvious is that build tags are not considered when saving dependencies, leaving out platform specific dependencies.

Secondly, various Go tools does not work with this. `go fmt`, `go test`, etc also processes the vendored packages, which creates problems for CI tests, that may depend on a different environment.  

For vendoring, I would still prefer a mono-repo, with import paths being rewritten.

### promises of 3rd party packages

For this project, we rely on third party packages. While this has the clear advantages described above, it does create some testing dilemmas. 

When considering test, we should not be writing tests for "external" packages, meaning, we should not test that external packages by themselves work. 
We must however provide some leverage to automatically test integration with the `func TestImport` in importlogs_test.go. This assures that we will be notified if formatting changes due to our own or third party updates.
We also provide "-test" parameter, which allows for integration tests with custom formatting, etc.

So dealing with third party packages does provide some scenarios where tests may feel insufficient, mainly due to the fact that we are providers or functionality not written by us. 
In that case it is a matter of evaluating the included package, or as in this case create tests that indicate changed behaviour.     

### model versioning

We do not provide any model versioning indication. For a long term project, this should be considered and implemented.

There is currently no way of "upgrading" data from one model to another, other than reimporting the data. 
For a limited data set, that is a viable strategy, but for a scalable system that may not be an option, and the data should have version indication to facilitate upgrades. 

## future considerations
 * Feature: Look up IP addresses for hosts. Probably requires caching. 
 * Feature: Add more storage backends.
 * Deployment: Docker image shouldn't include build environment.
 * Model: JSON representation should probably be `{"id" = {...}, ....}` to make order indedependent lookups easier. 

## facts

* Hours used: 24 hours, including documentation, sample deployment, excluding postmortem.
* 2 packages (1 model, 1 executable), 9 vendored packages.
* Vendoring: GO15VENDOREXPERIMENT.
* Deployment: Docker.
* IDE: Visual Studio Code.
* VCS: Git (TortoiseGIT).


----

Recruitment conversation-starter
================================
This brief assignment has the purpose of providing something tangible to talk about during recruitment interviews for a position as a systems developer.

When it comes to bringing a new developer on board, it is important to ensure that it is a good fit in terms of both technical competencies and development practice.

Expectations
------------
The first aspect is the _practice_ of developing software: How you handle yourself in a terminal, using software versioning systems, IDEs of preference, working in agile teams, meeting deadlines, keeping updated on current developments within the field etc. 
This is the stuff that keeps the wheels turning, and it’s just about as important to us as programming skills.

Which brings us to the second aspect, namely _technical competence_: Experience, knowledge and awareness of best practices on relevant development stacks. Languages, algorithms, paradigms and patterns. System architecture, frameworks, entity-relation-diagramming, database design. Security. Community participation and contributions to Open Source projects.

The Assignment
------------------------
In order to assess your competences in the two aspects above, we have formulated a brief assignment that can form the basis of conversation. We are aware that you have your own things going on, and this shouldn’t take more than a couple of hours to do. Remember that it is a basis for conversation, not a billable client project.

**_Focus on_**: Showcasing interesting use of technology, using standard components and patterns, following code standards, writing tests and documentation, and using your code versioning software well. Remember that we value both practice and technical competency.

**_Think about_**: What you want to talk about when we do the interview. It doesn’t matter if your implementation is not very fleshed out, if we feel that you have thought different solutions through and can argue for/against them.

What stack did you choose? Why? What issues of scaling did you think about? Performance? Monitoring? What third-party components did you use/avoid? Why?

Assignment Definition
---------------------
**_"Develop something that can periodically read from very large Apache log files, parse the log entries and store them in a structured fashion, so that they can be used for statistical analysis"._**

   * There are some decent [sample Apache data available from NASA](http://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html "NASA HTTP log file example").
   * We are interested in aggregating traffic per client IP/host, in bucket intervals of one hour. 
   * Output could include a diagram showing traffic fluctuations for time of day. Are NASA servers more busy mornings or evenings? 
   * Stack should include PHP/Symfony, Ruby, Python or Go  - your choice.

If you wish, you can focus on making a fast, parallelized log parser, on statistical analysis, or something entirely else that you find interesting. Feel free to impress.

Delivery
--------

 *   Deliver by sending us a link to a publicly available fork of this repository showing both code and commit history.
 *   Please include, in a brief readme, particular points or areas you wish us to focus on.
 
That’s it! We look forward to seeing what you can do!
