
#include "_cgo_export.h"
#include <jack/jack.h>


extern int jackClientCallback(jack_nframes_t nframes, void* arg);

int jackProcess(jack_nframes_t nframes, void *arg) {
  jackClientCallback(nframes, arg);
  return 0;
}

jack_client_t* wrap_jack_client_open(char* client_name) {
  jack_client_t* client = jack_client_open(client_name, JackNullOption, NULL);
  return client;
}

void register_callback(jack_client_t *client, void* arg) {
  jack_set_process_callback(client, jackProcess, arg);
}

