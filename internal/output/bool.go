package output

// FormatBool renders a boolean for human-readable output as "yes" / "no".
// Machine output (--json/--ndjson) always emits the raw true/false instead.
func FormatBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
