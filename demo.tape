Output demo.gif

Set FontSize 14
Set Width 900
Set Height 600
Set Theme "Dracula"

Type "go run ./cmd/mockserver -port 8090 -latency 2ms -error-rate 0.05 &"
Enter
Sleep 500ms

Type "go run ./cmd/barrage -url http://localhost:8090 -requests 1000 -workers 20 -rate 100"
Enter
Sleep 12s