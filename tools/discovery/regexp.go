package discovery

import (
	"regexp"
)

var (
	// TODO: This doesnot conform to https://tools.ietf.org/html/rfc3986 yet.
	host     = match(`[a-zA-Z0-9][a-zA-Z0-9\:\_\-\.]{1,128}`)
	path     = match(`[a-z0-9-_]+(?:/[a-z0-9-_.]+){0,7}`)
	fragment = match(`[a-z0-9-_.]{1,128}`)

	RegHostBasedImage = anchored(capture(host), literal(`/`), capture(path), optional(literal(`#`), capture(fragment)))
)

// match compiles the string to a regular expression.
var match = regexp.MustCompile

// literal compiles s into a literal regular expression, escaping any regexp
// reserved characters.
func literal(s string) *regexp.Regexp {
	re := match(regexp.QuoteMeta(s))

	if _, complete := re.LiteralPrefix(); !complete {
		panic("must be a literal")
	}

	return re
}

// expression defines a full expression, where each regular expression must
// follow the previous.
func expression(res ...*regexp.Regexp) *regexp.Regexp {
	var s string
	for _, re := range res {
		s += re.String()
	}

	return match(s)
}

// optional wraps the expression in a non-capturing group and makes the
// production optional.
func optional(res ...*regexp.Regexp) *regexp.Regexp {
	return match(group(expression(res...)).String() + `?`)
}

// repeated wraps the regexp in a non-capturing group to get one or more
// matches.
func repeated(res ...*regexp.Regexp) *regexp.Regexp {
	return match(group(expression(res...)).String() + `+`)
}

// group wraps the regexp in a non-capturing group.
func group(res ...*regexp.Regexp) *regexp.Regexp {
	return match(`(?:` + expression(res...).String() + `)`)
}

// capture wraps the expression in a capturing group.
func capture(res ...*regexp.Regexp) *regexp.Regexp {
	return match(`(` + expression(res...).String() + `)`)
}

// anchored anchors the regular expression by adding start and end delimiters.
func anchored(res ...*regexp.Regexp) *regexp.Regexp {
	return match(`^` + expression(res...).String() + `$`)
}
