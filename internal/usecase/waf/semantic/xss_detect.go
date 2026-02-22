package semantic

import "strings"

// detectXSS checks if the input contains XSS attack patterns.
// Uses HTML context analysis rather than simple regex matching.
func detectXSS(input string) bool {
	if len(input) == 0 {
		return false
	}

	lower := strings.ToLower(input)

	// Check for script tags
	if hasScriptTag(lower) {
		return true
	}

	// Check for event handlers
	if hasEventHandler(lower, input) {
		return true
	}

	// Check for dangerous protocols
	if hasDangerousProtocol(lower) {
		return true
	}

	// Check for dangerous HTML tags with executable context
	if hasDangerousTag(lower) {
		return true
	}

	return false
}

// hasScriptTag checks for <script> tag patterns.
func hasScriptTag(lower string) bool {
	// Look for <script with optional attributes
	idx := 0
	for {
		pos := strings.Index(lower[idx:], "<script")
		if pos == -1 {
			break
		}
		pos += idx
		afterTag := pos + 7
		if afterTag < len(lower) {
			ch := lower[afterTag]
			// <script> or <script followed by space/tab/newline (attributes)
			if ch == '>' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '/' {
				return true
			}
		} else if afterTag == len(lower) {
			// <script at end of string — suspicious
			return true
		}
		idx = afterTag
	}

	// Look for </script>
	if strings.Contains(lower, "</script") {
		return true
	}

	return false
}

// eventHandlers is the list of HTML event handler attribute names.
var eventHandlers = []string{
	"onerror", "onload", "onclick", "onmouseover", "onfocus", "onblur",
	"onsubmit", "onchange", "onkeyup", "onkeydown", "onkeypress",
	"onmouseout", "onmouseenter", "onmouseleave", "onmousedown", "onmouseup",
	"onmousemove", "ondblclick", "oncontextmenu", "onwheel", "onscroll",
	"ondrag", "ondragstart", "ondragend", "ondragover", "ondragenter",
	"ondragleave", "ondrop", "oncopy", "oncut", "onpaste", "onselect",
	"oninput", "onreset", "onsearch", "ontoggle", "onformdata",
	"oninvalid", "onbeforeinput", "onanimationstart", "onanimationend",
	"onanimationiteration", "ontransitionend", "ontransitionstart",
	"onpointerdown", "onpointerup", "onpointermove", "onpointerover",
	"onpointerout", "onpointerenter", "onpointerleave",
	"ongotpointercapture", "onlostpointercapture",
	"ontouchstart", "ontouchend", "ontouchmove", "ontouchcancel",
	"onafterprint", "onbeforeprint", "onbeforeunload", "onhashchange",
	"onmessage", "onoffline", "ononline", "onpagehide", "onpageshow",
	"onpopstate", "onstorage", "onunload", "onresize",
	"onabort", "oncanplay", "oncanplaythrough", "ondurationchange",
	"onemptied", "onended", "onloadeddata", "onloadedmetadata",
	"onloadstart", "onpause", "onplay", "onplaying", "onprogress",
	"onratechange", "onseeked", "onseeking", "onstalled", "onsuspend",
	"ontimeupdate", "onvolumechange", "onwaiting",
}

// hasEventHandler checks for on*= event handler patterns in HTML context.
func hasEventHandler(lower, original string) bool {
	// We need to find on*= patterns that are within an HTML tag context
	for _, handler := range eventHandlers {
		idx := 0
		for {
			pos := strings.Index(lower[idx:], handler)
			if pos == -1 {
				break
			}
			pos += idx

			// Check if followed by = (possibly with whitespace)
			end := pos + len(handler)
			for end < len(lower) && (lower[end] == ' ' || lower[end] == '\t') {
				end++
			}
			if end < len(lower) && lower[end] == '=' {
				// Check if we're in an HTML tag context (look back for '<')
				if isInTagContext(lower, pos) {
					return true
				}
			}
			idx = pos + 1
		}
	}
	return false
}

// isInTagContext checks if position is inside an HTML tag by looking backward for '<'.
func isInTagContext(lower string, pos int) bool {
	// Walk backward from pos, looking for '<' before any '>'
	for i := pos - 1; i >= 0; i-- {
		if lower[i] == '<' {
			return true
		}
		if lower[i] == '>' {
			return false
		}
	}
	return false
}

// hasDangerousProtocol checks for javascript: or vbscript: protocol usage.
func hasDangerousProtocol(lower string) bool {
	// javascript: (with optional whitespace between javascript and colon)
	protocols := []string{"javascript", "vbscript"}

	for _, proto := range protocols {
		idx := 0
		for {
			pos := strings.Index(lower[idx:], proto)
			if pos == -1 {
				break
			}
			pos += idx
			// Check for colon after the protocol name (with optional whitespace)
			end := pos + len(proto)
			for end < len(lower) && (lower[end] == ' ' || lower[end] == '\t' || lower[end] == '\n' || lower[end] == '\r') {
				end++
			}
			if end < len(lower) && lower[end] == ':' {
				return true
			}
			idx = pos + 1
		}
	}

	return false
}

// dangerousTags are HTML tags that can execute code.
var dangerousTags = []string{
	"iframe", "object", "embed", "applet", "form",
	"base", "link", "meta",
}

// hasDangerousTag checks for dangerous HTML tags with event handlers or src attributes.
func hasDangerousTag(lower string) bool {
	for _, tag := range dangerousTags {
		pattern := "<" + tag
		idx := 0
		for {
			pos := strings.Index(lower[idx:], pattern)
			if pos == -1 {
				break
			}
			pos += idx
			afterTag := pos + len(pattern)
			if afterTag < len(lower) {
				ch := lower[afterTag]
				if ch == '>' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '/' {
					return true
				}
			}
			idx = pos + 1
		}
	}

	// SVG with event handlers
	if svgIdx := strings.Index(lower, "<svg"); svgIdx != -1 {
		afterSvg := svgIdx + 4
		if afterSvg < len(lower) {
			ch := lower[afterSvg]
			if ch == '>' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '/' {
				// Check for onload or other event handlers
				rest := lower[afterSvg:]
				closingIdx := strings.Index(rest, ">")
				if closingIdx == -1 {
					closingIdx = len(rest)
				}
				tagContent := rest[:closingIdx]
				if strings.Contains(tagContent, "onload") || strings.Contains(tagContent, "onerror") {
					return true
				}
			}
		}
	}

	// <math> with event handlers
	if mathIdx := strings.Index(lower, "<math"); mathIdx != -1 {
		afterMath := mathIdx + 5
		if afterMath < len(lower) {
			ch := lower[afterMath]
			if ch == '>' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '/' {
				rest := lower[afterMath:]
				closingIdx := strings.Index(rest, ">")
				if closingIdx == -1 {
					closingIdx = len(rest)
				}
				tagContent := rest[:closingIdx]
				if strings.Contains(tagContent, "onload") || strings.Contains(tagContent, "onerror") {
					return true
				}
			}
		}
	}

	return false
}
