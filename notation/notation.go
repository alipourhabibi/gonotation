package notation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alipourhabibi/gonotation/glob"
	"github.com/dlclark/regexp2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var ERRInvalidRegex = fmt.Errorf("Invalid regex")

type Notation struct {
	source      string
	isArray     bool
	restrictive bool
}

func New(source string) Notation {
	return Notation{
		source: source,
	}
}

func (n Notation) Filter(globList []glob.Glob, restrictive bool) (string, error) {
	notation, err := n.filter(globList, restrictive)
	return notation.source, err
}

func (n Notation) filter(globList []glob.Glob, restictive bool) (Notation, error) {
	n.restrictive = restictive
	globs := glob.Normalize(globList, restictive)
	length := len(globs)
	empty := "{}"
	t := gjson.Parse(n.source)
	if t.Type.String() == "array" {
		n.isArray = true
		empty = "[]"
	}
	m, err := regexp.MatchString(glob.NEGATE_ALL, globs[0])
	if err != nil {
		return Notation{}, ERRInvalidRegex
	}
	if length == 0 || (length == 1 && (globs[0] == "" || m)) {
		n.source = empty
	}
	cloned := n.source
	firstWildCard, err := regexp.MatchString(glob.WILDCARD, globs[0])
	if length == 1 && firstWildCard {
		n.source = cloned
	}
	filtered := Notation{}
	if firstWildCard {
		filtered = New(cloned)
		if len(globs) >= 1 {

		}
	} else {
		filtered = New(empty)
	}
	for i := 0; i < len(globs); i++ {
		var normalized string
		var emptyValue string
		var eType string
		globNotation := globs[i]
		g := glob.NewInspect(globNotation)
		if len(g.AbsGlob.GetGlob()) >= 2 && g.AbsGlob.GetGlob()[len(g.AbsGlob.GetGlob())-2:] == ".*" {
			normalized = g.AbsGlob.GetGlob()[:len(g.AbsGlob.GetGlob())-2]
			if g.IsNegated {
				emptyValue = "{}"
				eType = "object"
			}
		} else if len(g.AbsGlob.GetGlob()) >= 3 && g.AbsGlob.GetGlob()[len(g.AbsGlob.GetGlob())-3:] == "[*]" {
			normalized = g.AbsGlob.GetGlob()[:len(g.AbsGlob.GetGlob())-3]
			if g.IsNegated {
				emptyValue = "[]"
				eType = "array"
			}
		} else {
			normalized = g.AbsGlob.GetGlob()
		}

		rg := regexp2.MustCompile(glob.WILDCARDS, 0)
		m, err := rg.MatchString(normalized)
		if err != nil {
			return Notation{}, ERRInvalidRegex
		}
		if !m {
			if g.IsNegated {
				insRemove, err := sjson.Delete(filtered.source, normalized)
				if err != nil {
					return Notation{}, ERRInvalidRegex
				}
				filtered.source = insRemove
				par := parent(insRemove)
				if emptyValue != "" {
					if insRemove == "undefined" || insRemove == "Null" {
						// isValSet = false
					}
					not := gjson.Parse(par)
					// setMode := "overwrite"
					if not.Type.String() == "array" {
						// setMode = "insert"
					}
					//filtered.source, err = sjson.Set(filtered.source, normalized, emptyValue)
					filtered.source, err = sjson.Delete(filtered.source, normalized)
					if err != nil {
						return Notation{}, ERRInvalidRegex
					}
				}
			} else {
				insGet := gjson.Get(n.source, normalized)
				filtered.source, err = sjson.Set(filtered.source, normalized, insGet.Value())
				if err != nil {
					return Notation{}, ERRInvalidRegex
				}
			}
		}
		gsource := gjson.Parse(n.source)
		gsource.ForEach(func(key, value gjson.Result) bool {
			orig := glob.NewInspect(normalized)
			originalIsCovered := glob.Covers(orig, glob.NewInspect(key.String()), false)
			if !originalIsCovered {
				return true
			}
			if n.restrictive && emptyValue != "" {
				vType := gjson.Parse(value.String()).Type.String()
				spl, err := split(key.String())
				if err != nil {
					return false
				}
				if vType != eType && len(spl) == len(g.Glob.Notes)-1 {
					return false
				}
			}
			res := gjson.Parse(n.source)
			res.ForEach(func(k2, v2 gjson.Result) bool {
				// m, err := regexp.MatchString(v2.String(), g.Glob.GetGlob())
				m := glob.Covers(g, glob.NewInspect(k2.String()), false)
				if err != nil {
					return false
				}
				if m {
					l, _ := split(v2.String())
					levelLen := len(l)
					if g.IsNegated && len(g.Glob.Notes) <= levelLen {
						filtered.source, err = sjson.Delete(filtered.source, g.AbsGlob.GetGlob())
						return false
					}
				}
				return true
			})
			return true
		})
	}
	return filtered, nil
}

func split(notation string) ([]string, error) {
	reMatcher := regexp.MustCompile(glob.REMATCHER)
	return reMatcher.FindAllString(notation, -1), nil
}
func last(notation string) string {
	list, _ := split(notation)
	return list[len(list)-1]
}

func parent(notation string) string {
	last := last(notation)
	parent := notation[:len(notation)-len(last)]
	parent = strings.TrimSuffix(parent, ".")
	return parent
}
