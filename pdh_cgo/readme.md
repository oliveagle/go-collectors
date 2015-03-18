why?
====

`pdh` package is fine while developing. but once I deploy it to other windows machines. It raises `segmentation fault` error frequently!
With some digging, I found out that once I comment out `AddEnglishCounter` line from my working code, this problem won't show up again.

This problem is too weird to debug....

So I tried to use static library in c and embed it with cgo. and this approach works!

