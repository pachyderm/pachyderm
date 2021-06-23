// package clientsdk implements functions for using the gRPC APIs for the various Pachyderm services.
//
// The approach of this package is "maybe better, never worse", meaning no functionality is obfuscated, changed, or removed
// from the underlying generated gRPC clients.
// The functions in this package can be used to remove boilerplate, but the caller is never in a position where they need to manage 2 parallel abstractions.
//
// Another `Client` struct probably does not belong in here.

package clientsdk
