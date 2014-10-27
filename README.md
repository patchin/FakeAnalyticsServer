debugAnalyticsServer
====================

Go webserver that prints out the Google Analytics requests from clients altered to send traffic to it.

This directory should be placed in the following directory hierarchy:
<go proj dir>/src/
and the GOPATH env be setup so that the Go tools behave properly. 
Go tools are designed with convention over configuration in mind.
(http://golang.org/doc/code.html#tmp_2)

This server is designed to run in the Heroku environment with the Go buildpack.
It outputs everything to stdout, and that can be viewed with 'heroku logs' or 'heroku logs -t'
