# Go-MikMod

Use [MikMod](http://mikmod.sourceforge.net/) from Go.  Incompletely!

# Example

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/death/go-mikmod"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Supply a module filename")
	}

	if err := mikmod.Init(); err != nil {
		log.Fatal(err)
	}
	defer mikmod.Uninit()

	m, err := mikmod.LoadModuleFromFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	log.Printf("Playing '%s'...\n", m.Title())

	mikmod.Play(m)
	defer mikmod.Stop()
	for mikmod.IsPlaying() {
		time.Sleep(time.Second)
	}
}
```

# License

MIT
