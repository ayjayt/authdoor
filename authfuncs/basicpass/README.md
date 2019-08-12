# basicpass

basicpass simple implements an interface over any path that forces the user to enter a password. it curretly has no way to logout except by deleting the cookie, although a logout path can be set.

send textbox with submit button
submit button will post to same domain as location
auth func needs to look for a cookie and map it to its session map
authfunc needs to look for a particular type of body, and if it finds it, check it against the password
if authed, auth it, but also set a cookie on the user

the submit button could
b) post back with a body. if the authhandler detects any POST, it's fine, it tries to process it.
c) it could create a seession token based on that and write a cookie.
