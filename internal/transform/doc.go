// Package transform provides composable secret value transformation functions
// for use in vaultshift pipelines.
//
// A Transformer holds an ordered list of Func steps. Each Func receives the
// current value and returns the transformed value or an error. Steps are
// applied in sequence; the first error halts execution.
//
// Built-in transformation functions:
//
//   - TrimSpace   – removes leading and trailing whitespace
//   - ToUpper     – converts the value to uppercase
//   - ToLower     – converts the value to lowercase
//   - AddPrefix   – prepends a static string to the value
//   - Base64Encode – encodes the value as standard base64
//   - Base64Decode – decodes a standard base64-encoded value
//
// Custom steps can be provided as any function matching the Func signature.
//
// Example:
//
//	tr := transform.New(
//		transform.TrimSpace(),
//		transform.ToUpper(),
//		transform.AddPrefix("APP_"),
//	)
//	result, err := tr.Apply("  my_secret  ")
//	// result == "APP_MY_SECRET"
package transform
