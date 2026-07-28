# demo.tape
Output demo.gif

Set FontSize 14
Set Width 800
Set Height 500
Set Theme "Dracula"

Type "go run ./cmd/mockserver -port 8090 -latency 5ms -error-rate 0.02 &"
Enter
Sleep 500ms

Type "go run ./cmd/barrage -url http://localhost:8090 -requests 500 -workers 20 -rate 80"
Enter
Sleep 8s