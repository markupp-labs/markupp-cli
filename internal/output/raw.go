package output

// Raw writes text exactly as given, without sanitizing it and without adding a final
// newline, for note content the caller asked for byte for byte (mu notes get
// --content), so it can be saved and edited back.
func (p *Printer) Raw(text string) error { return p.write(text) }
