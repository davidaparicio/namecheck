## Examples

With Docker

```
david@mac:~$ docker run davidaparicio/namecheck:v0.0.5 dadideo
Unable to find image 'davidaparicio/namecheck:v0.0.5' locally
v0.0.5: Pulling from davidaparicio/namecheck
78c398760b13: Pull complete
ec85b45ae3bd: Pull complete
Digest: sha256:8155d0b049b6080100f32d47892c514fbb6fa7f790e03af47b42adc7b7ee0b39
Status: Downloaded newer image for davidaparicio/namecheck:v0.0.5
{dadideo GitHub true true}
{dadideo GitHub true true}
{dadideo GitHub true true}
{dadideo Twitter true false}
{dadideo Twitter true false}
{dadideo Twitter true false}
```

With Go (Build or Run)

```
david@mac:~$ go run cmd/cli/main.go dadieo
{dadieo GitHub true true}
{dadieo Twitter true false}
{dadieo Twitter true false}
{dadieo Twitter true false}
{dadieo GitHub true true}
{dadieo GitHub true true}
```
The format JSON (CLI) is: ```{pseudo provider isThePseudoValid isThePseudoAvailable}```

With Server+Curl
```
curl 'http://localhost:8080/check?username=dadideo'
{"username":"dadideo","results":[{"platform":"Twitter","valid":"true","available":"false"},{"platform":"Twitter","valid":"true","available":"false"},{"platform":"Twitter","valid":"true","available":"false"},{"platform":"GitHub","valid":"true","available":"true"},{"platform":"GitHub","valid":"true","available":"true"},{"platform":"GitHub","valid":"true","available":"true"}]}
```

The format JSON (SERVER) is: ```{"username":"XXX","results":[{"platform":"XXX","valid":"true/false","available":"true/false"}```

With Server+Browser (like Mozilla Firefox)

![Firefox Namecheck Example with Syntax coloration](https://i58.servimg.com/u/f58/11/58/68/69/namech12.jpg)