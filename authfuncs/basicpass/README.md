# basicpass

basicpass simple implements an interface over any path that forces the user to enter a password. it curretly has no way to logout except by deleting the cookie, although a logout path can be set.

todo: figure out how to declare an authfunc, the authfunc shoudl fail and respond with an input text box and a submit button.

the submit button could
a) write a cookie (nah)
b) post back with a body. if the authhandler detects any POST, it's fine, it tries to process it.
c) it could create a seession token based on that and write a cookie.
