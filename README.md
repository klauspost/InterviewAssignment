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
