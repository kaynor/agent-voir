package usage

// NopRecorder discards usage events.
type NopRecorder struct{}

func (NopRecorder) Record(_ Event) {}
