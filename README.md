# go-collectors
metrics collectors written in go

* [Linux][url_travis]: ![ci travis][ci_travis]
* [Windows][url_ci_win]: ![windows test][ci_windows]  

[ci_travis]: https://travis-ci.org/oliveagle/go-collectors.svg "CI Travis"
[ci_windows]: https://ci.appveyor.com/api/projects/status/github/oliveagle/go-collectors?branch=master&svg=true "Windows Build"

[url_travis]: https://travis-ci.org/oliveagle/go-collectors "url travis"
[url_ci_win]: https://ci.appveyor.com/project/oliveagle/go-collectors/build/1.0.20 "url windows ci"


** play at your own risk **

`go-collectors` is ported from `bosun.org` project, and is focusing on functions to collect metrics only. no saving, no reporting at all. 


##### WINDOWS CI:

coz windows ci is reletively slow compare to travis, many commits just line up and wait to be tested. so windows ci will only test `ci_win` branch. 