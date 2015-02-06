# rsstweetbot
A simple Bot that will fetch an RSS feed and Tweet each headline and link

I hacked this together in a day as a proof of concept. There are no tests and likely bugs. I'll update this on the go and will improve the code with at least proper logging and maybe a text config instead of one in code.

# usage
Copy [the example config](config.go.example) to config.go (or any other .go file name) and enter:
* your RSS feed URL
* Your Twitter authentication data
* A location for the sqlite cache database

compile and run!
