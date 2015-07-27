jackclient
==========

Jack audio client library for go. 

## Note

This isn't working with the current Go release. The new garbage collector moves things around in memory so the callback pointer passed into C can (and will) become invalid. 

The fix for this problem is to pass an integer into the C code that indexes into an array or map that contains a pointer to the Go object. 
