# x

This project is powering a site called 8hrs.xyz which keeps all frontpage articles from hacker news for 8 hours.

Use latest go with modules to build this project.

build.sh builds with version as the latest git commit hash.

sqlitedb is needed. create a database file before running the app.

configuration that is required to run the executable are passed as env vars and listed in  x.service.

modify the config.json as required.

there is one http endpoint.

. `/` serves the index html or json based on reqquested content-type 
. `/feed/{rss|atom|json}` responds with rss or atom or json feed

(order of list items is randomized between requests)

http rate limiting middleware is used to limit requests to 5 per minute.

this whole flow of the app depends on the hacker news apis.

it also tweets the news items to twitter.

Previous code was at github.com/mseshachalam/x8h which did not use any database. 
