Output demo.gif

Set FontSize 13
Set Width 900
Set Height 550
Set Theme "Dracula"

Type "go run ./cmd/mockserver -port 8090 -latency 3ms -error-rate 0.03 &"
Enter
Sleep 800ms

Type "go run ./cmd/barrage -url http://localhost:8090 -requests 800 -workers 20 -rate 80"
Enter
Sleep 12s