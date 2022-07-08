// The featureflags module exists to control optional functionality
//
// The main purpose is to hide functionality that is still in
// development, that isn't yet fully featured enough to be GA, or that
// isn't relevant to all users.
//
// The format is key-value, both strings, for rare situations where
// new values need to be introduced to the same flag.  You should
// usually treat the value as a boolean - create one flag per
// alternative, don't create complicated sets of values. E.g. set
// FOO_BAR_ENABLED and FOO_BAR_IN_MAIN_MENU as separate flags, not
// FOO_BAR="enabled" or "FOO_BAR="in_main_menu", as the latter format
// is more likely to be confused. At the same time, you should check
// for exact value when checking the flag, to avoid misinterpreting it
// in the future.
//
// All flags that are set here on the backend will also be set in the
// featureflags endpoint on the frontend.
package featureflags

var flags map[string]string = make(map[string]string)

// Set sets one specific featureflag
// Existing flags will be overwritten.
func Set(key, value string) {
	flags[key] = value
}

// Get returns the value of a flag, or "" if the flag wasn't set
// You should use the same behaviour for "flag not set" as you would
// for "flag set to unknown value", so always check for the exact
// value.
func Get(key string) string {
	return flags[key]
}

// GetFlags returns all featureflags
// This is only intended to be used by the API to return the flags to
// the frontend - for all other uses, use `Get`
func GetFlags() map[string]string {
	return flags
}
