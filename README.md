# Authdoor

Authdoor is reverse proxy supporting easy, accessible configuration and arbitrary authentication protocols.

Okay, authdoor allows to you create a list of auth functions.
Authfunctions can be seperate modules? submodules within gomodules? these will be the providers.
Authdoor creates a slice of functions, which may have data structures that are shared among multiple goroutines. The slice itself is obviously read by multiple goroutines- read.

TODO:
we need to have namespaced authlists (cookie name)
You can also add and remove authfunc instances.
This needs to be fleshed out, especially with waitgroups and states:

The writers use the mutex to know when they can go... but for the webserver it will need to be ordered so that writers stay in transaction order (we can support transactions entirely in this model)
We need to let the users finish their flight before we write to the stage buffer (in case they are still in it)- there should be some timeout in case someone hangs
Write into the stage buffer and switch it up
Remember that we can use a cookie (read: should use a cookie) to tell what auth we shoudl choose (that auth may use the cookie), if it fails, we should scroll through them one by one.

If auth fails completely, if no one knows who the user is, a login page obviously needs to be presented.... we'll have to figure this.


There also needs to be a server that will dynamically create authdoor, associate it with file servers/routers (read config and also grpc). Probably using fast http.




The slice is ordered by a priority number. Each auth function has a structure that gives it a priority, a name, and a function. There is a map that points to the index, and probably a counter to indicate how many ppeople are reading (and a lock if currently writing),

When you insert or remove an auth function, it should probably write a whole new datastructure.
There should be a lock to demonstrate that we are currently transitioning the datastructure, and not start an op.
Then you should write a value that indicates which auth structure to use. When the previous auth structure moves to 0, then you can start a new operation.

But these two slices might need to share state, like with a dynamic function of changing variables.



// This (below) is better for the server which will dynamically change handlers and use the auth functions its loaded with
// We also need to attach actual handlers, but maybe that's better for the server
// We also need to turn this into a javascript endpoint, probably with GRPC. Not sure if that will be in the server or w/e.
// We also need to integrate it such that we can retaint some user info, and that it can be used as the beginning of an auth system
// We should have a module so that it can store custom auth
// We should have a module so that we can verify sessions
// But how are we going to namespace it...
// And we're going to rewrite it in C
// Also we should have some way to store user info, and an arbitrary way to process data
type AuthFuncRegstry struct {
	value map[string]AuthFunc
}

// Add will add a function to an authFunc registry
func (r *AuthFuncRegistry) Add(authFunc AuthFunc, name string) error {
	if _, ok := r.value[name]; ok {
		return ErrNameTaken
	}
	r.value[name] = authFunc
	return nil
}

# TODO:

Rename from router to authdoor- file and repo
