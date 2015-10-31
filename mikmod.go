// Package mikmod lets you use MikMod from Go.
package mikmod

/*
#cgo LDFLAGS: -lmikmod
#include <mikmod.h>
*/
import "C"

import (
	"errors"
	"sync"
	"time"
	"unsafe"
)

// Return MikMod library version.
func Version() (major int, minor int, rev int) {
	v := C.MikMod_GetVersion()
	rev = int(v & 0xFF)
	minor = int((v >> 8) & 0xFF)
	major = int((v >> 16) & 0xFF)
	return
}

// Init initializes the MikMod library.  Make sure to call Uninit when
// done.
func Init() error {
	C.MikMod_InitThreads()
	C.MikMod_RegisterAllDrivers()
	C.MikMod_RegisterAllLoaders()
	C.md_mode = C.DMODE_SOFT_MUSIC | C.DMODE_NOISEREDUCTION
	initString := mikmodString("")
	defer C.free(unsafe.Pointer(initString))
	if err := int(C.MikMod_Init(initString)); err != 0 {
		return mikmodError()
	}

	return nil
}

// Uninit uninitializes the MikMod library.
func Uninit() {
	Stop()
	C.MikMod_Exit()
}

// Module represents a MikMod module.  Remember to Close it when done.
type Module struct {
	module *C.MODULE
}

// LoadModuleFromFile attempts to load a MikMod module from the file
// designated by filename.
func LoadModuleFromFile(filename string) (*Module, error) {
	fn := mikmodString(filename)
	defer C.free(unsafe.Pointer(fn))
	module := C.Player_Load(fn, 128, C.BOOL(0))
	if module == nil {
		return nil, mikmodError()
	}
	module.loop = 0
	module.fadeout = 1
	return &Module{module}, nil
}

// LoadModuleFromSlice attempts to load a MikMod module from the
// supplied byte slice.
func LoadModuleFromSlice(b []byte) (*Module, error) {
	module := C.Player_LoadMem((*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)), 128, C.BOOL(0))
	if module == nil {
		return nil, mikmodError()
	}
	module.loop = 0
	module.fadeout = 1
	return &Module{module}, nil
}

// Title returns the module's song name.
func (m *Module) Title() string { return C.GoString((*C.char)(m.module.songname)) }

// NumChannels returns the number of channels used by the module.
func (m *Module) NumChannels() int { return int(m.module.numchn) }

// NumVoices returns the number of voices reserved by the player for
// real and virtual channels.
func (m *Module) NumVoices() int { return int(m.module.numvoices) }

// NumPositions returns the number of song positions.
func (m *Module) NumPositions() int { return int(m.module.numpos) }

// NumPatterns returns the number of song patterns.
func (m *Module) NumPatterns() int { return int(m.module.numpat) }

// NumInstruments returns the number of instruments in the module.
func (m *Module) NumInstruments() int { return int(m.module.numins) }

// NumSamples returns the number of samples in the module.
func (m *Module) NumSamples() int { return int(m.module.numsmp) }

// Tracker returns the name of the tracker used to create the song.
func (m *Module) Tracker() string { return C.GoString((*C.char)(m.module.modtype)) }

// Comment returns the song comment.
func (m *Module) Comment() string { return C.GoString((*C.char)(m.module.comment)) }

// Elapsed returns the time elapsed since the song started playing.
func (m *Module) Elapsed() time.Duration {
	return time.Duration(m.module.sngtime*1000/1024) * time.Millisecond
}

// Speed returns the song speed.
func (m *Module) Speed() int { return int(m.module.sngspd) }

// Tempo returns the song tempo.
func (m *Module) Tempo() int { return int(m.module.bpm) }

// Close frees the module, making it unusable.
func (m *Module) Close() error {
	C.Player_Free(m.module)
	m.module = nil
	return nil
}

var (
	finish chan struct{}
	done   sync.WaitGroup
)

// updateLoop calls MikMod's update routine every 10ms.  It terminates
// when the finish channel is closed, calling Done on the done
// waitgroup.
func updateLoop() {
	for {
		select {
		case <-time.After(10 * time.Millisecond):
			C.MikMod_Update()
		case <-finish:
			done.Done()
			return
		}
	}
}

// Play starts playing a module.
func Play(m *Module) {
	if finish != nil {
		Stop()
	}

	C.Player_Start(m.module)

	finish = make(chan struct{})
	done.Add(1)
	go updateLoop()
}

// Stop stops playing a module.
func Stop() {
	if finish == nil {
		return
	}

	C.Player_Stop()

	close(finish)
	done.Wait()
	finish = nil
}

// IsPlaying returns true if the player is active, and false
// otherwise.
func IsPlaying() bool {
	return int(C.Player_Active()) != 0
}

// mikmodString converts a Go string to a MikMod string; make sure to
// call C.free() on it when done with it.
func mikmodString(s string) *C.CHAR {
	return (*C.CHAR)(C.CString(s))
}

// mikmodErrro returns a Go error corresponding to the current MikMod
// error.
func mikmodError() error {
	return errors.New(C.GoString(C.MikMod_strerror(C.MikMod_errno)))
}
