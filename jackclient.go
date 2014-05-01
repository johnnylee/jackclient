package jackclient

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

/*
 #cgo pkg-config: jack
 #include <jack/jack.h>

 extern jack_client_t* wrap_jack_client_open(char*);
 extern void register_callback(jack_client_t*, void*);
*/
import "C"

// ----------------------------------------------------------------------------
func cArrayToSlice32f(cArray *C.float, length int, goSlice *[]float32) {
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(goSlice)))
	sliceHeader.Cap = length
	sliceHeader.Len = length
	sliceHeader.Data = uintptr(unsafe.Pointer(cArray))
}

// ----------------------------------------------------------------------------
type JackClient struct {
	name   string
	client *C.jack_client_t

	// Our jack input and output ports.
	jackIn  []*C.jack_port_t
	jackOut []*C.jack_port_t

	// Our golang buffers for the ports above.
	bufIn  [][]float32
	bufOut [][]float32

	// The callback.
	callback func(bufIn, bufOut [][]float32) error
}

// New: Create a new jack client with the given name and the given number
// of input and output ports.
func New(name string, numInputs, numOutputs int) (*JackClient, error) {
	jc := new(JackClient)
	jc.name = name

	// Create jack client.
	jc.client = C.wrap_jack_client_open(C.CString(name))
	if jc.client == nil {
		return nil, errors.New("Failed to initialize jack client.")
	}

	// Create ports.
	for i := 0; i < numInputs; i++ {
		port := C.jack_port_register(
			jc.client,
			C.CString(fmt.Sprintf("in_%d", i)),
			C.CString(C.JACK_DEFAULT_AUDIO_TYPE),
			C.JackPortIsInput, 0)
		jc.jackIn = append(jc.jackIn, port)
	}

	for i := 0; i < numOutputs; i++ {
		port := C.jack_port_register(
			jc.client,
			C.CString(fmt.Sprintf("out_%d", i)),
			C.CString(C.JACK_DEFAULT_AUDIO_TYPE),
			C.JackPortIsOutput, 0)
		jc.jackOut = append(jc.jackOut, port)
	}

	// Create buffers.
	jc.bufIn = make([][]float32, numInputs)
	jc.bufOut = make([][]float32, numOutputs)

	// Register jack callback.
	C.register_callback(jc.client, unsafe.Pointer(jc))
	C.jack_activate(jc.client)

	// Done.
	return jc, nil
}

// Register the callback function to be called by jackd.
// The function must take inL, inR, outL and outR []float32 slices.
// If the function returns a non-nil error, it will cause the jack client
// to be removed from the call graph.
func (jc *JackClient) RegisterCallback(
	cb func(bufIn, bufOut [][]float32) error) {
	jc.callback = cb
}

func (jc *JackClient) GetSampleRate() int {
	return int(C.jack_get_sample_rate(jc.client))
}

func (jc *JackClient) GetBufferSize() int {
	return int(C.jack_get_buffer_size(jc.client))
}

//export jackClientCallback
func jackClientCallback(nframes C.jack_nframes_t, arg unsafe.Pointer) C.int {
	jc := (*JackClient)(arg)

	// Check for no callback set.
	if jc.callback == nil {
		return 0
	}

	var cArr *C.float
	length := int(nframes)

	for i, port := range jc.jackIn {
		cArr = (*C.float)(C.jack_port_get_buffer(port, nframes))
		cArrayToSlice32f(cArr, length, &jc.bufIn[i])
	}
	for i, port := range jc.jackOut {
		cArr = (*C.float)(C.jack_port_get_buffer(port, nframes))
		cArrayToSlice32f(cArr, length, &jc.bufOut[i])
	}

	if err := jc.callback(jc.bufIn, jc.bufOut); err != nil {
		return 1
	}

	return 0
}
