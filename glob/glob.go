package glob

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var restrictive = false

// const reMATCHER = "/(\\[(\\d+|\\*|\".*\"|'.*')\\]|[a-z$_][a-z$_\\d]*|\\*)/gi;"
const reMATCHER = `(?mi)(!?)(\[(\d+|\*|".*"|'.*')\]|[a-z$_][a-z$_\d]*|\*)`
const REMATCHER = reMATCHER

// const VAR = "/^[a-z$_][a-z$_\\d]*$/i"
const VAR = `(?i)^[a-z$_][a-z$_\\d]*$`
const ARRAY_NOTE = "/^\\[(\\d+)\\]$/"

// const ARRAY_GLOB_NOTE = "/^\\[(\\d+|\\*)\\]$/"
const ARRAY_GLOB_NOTE = `^\[(\d+|\*)\]$`
const OBJECT_BRACKETS = "/^\\[(?:'(.*)'|\"(.*)\"|`(.*)`)\\]$/"

// const WILDCARD = "/^(\\[\\*\\]|\\*)$/"
const WILDCARD = `^(\[\*\]|\*)$`

// matches `*` and `[*]` if outside of quotes.
// const WILDCARDS = "/(\\*|\\[\\*\\])(?=(?:[^\"]|\"[^\"]*\")*$)(?=(?:[^']|'[^']*')*$)/"
const WILDCARDS = `(\*|\[\*\])(?=(?:[^"]|"[^"]*")*$)(?=(?:[^']|'[^']*')*$)`

// matches trailing wildcards at the end of a non-negated glob.
// e.g. `x.y.*[*].*` » $1 = `x.y`, $2 = `.*[*].*`
const NON_NEG_WILDCARD_TRAIL = "/^(?!!)(.+?)(\\.\\*|\\[\\*\\])+$/"
const NEGATE_ALL = "/^!(\\*|\\[\\*\\])$/"

// ending with '.*' or '[*]'

const _reFlags = "/\\w*$/"

var re []string = []string{
	VAR,
	ARRAY_NOTE,
	ARRAY_GLOB_NOTE,
	OBJECT_BRACKETS,
	WILDCARD,
	WILDCARDS,
	NON_NEG_WILDCARD_TRAIL,
	NEGATE_ALL,
}

func Covers(a, b GlobInspect, match bool) bool {
	return covers(a, b, match)
}

func covers(a, b GlobInspect, match bool) bool {

	notesA := a.Glob.Notes
	if notesA == nil {
		notesA = split(a.AbsGlob, restrictive)
	}

	notesB := b.Glob.Notes
	if notesB == nil {
		notesB = split(b.AbsGlob, restrictive)
	}

	if !match && a.Glob.isNegated && len(notesA) > len(notesB) {
		return false
	}

	covers := true
	fn := coversNote
	if match {
		fn = matchesNote
	}

	for i := 0; i < len(notesA); i++ {
		if len(notesB) <= i {
			break
		}
		if !fn(notesA[i], notesB[i]) {
			covers = false
			break
		}
	}

	return covers
}
func matchesNote(a, b string) bool {
	if a == "" || b == "" {
		return true // glob e.g.: [2][1] matches [2] and vice-versa.
	}
	return coversNote(a, b) || coversNote(b, a)
}
func coversNote(a, b string) bool {
	if a == "" || b == "" {
		return false // glob e.g.: [2] does not cover [2][1]
	}
	bIsArr, _ := regexp.MatchString(ARRAY_GLOB_NOTE, b)
	if a == "*" {
		return !bIsArr // obj-wildcard a will cover b if not array
	}
	if a == "[*]" {
		return bIsArr // arr-wildcard a will cover b if array
	}
	reg, _ := regexp.MatchString(WILDCARD, b)
	if reg {
		return false // if b is wildcard (obj or arr) won't be covered
	}
	g := normalizeNote(a) == normalizeNote(b) // normalize both and check for equality
	return g
}
func NormalizeNote(note string) string {
	r, _ := regexp.MatchString(VAR, note)
	if r {
		return note
	}

	m := regexp.MustCompile(ARRAY_NOTE).FindStringSubmatch(note)
	if len(m) > 1 {
		d, err := strconv.Atoi(m[1])
		if err != nil {
			return ""
		}
		return string(d)
	}

	m = regexp.MustCompile(OBJECT_BRACKETS).FindStringSubmatch(note)
	if len(m) > 0 {
		return m[1] + m[2] + m[3]
	}

	return ""
}

func normalizeNote(note string) bool {
	r, _ := regexp.MatchString(VAR, note)
	if r {
		return true
	}

	m := regexp.MustCompile(ARRAY_NOTE).FindStringSubmatch(note)
	if len(m) > 1 {
		_, err := strconv.Atoi(m[1])
		if err != nil {
			return false
		}
		return true
	}

	m = regexp.MustCompile(OBJECT_BRACKETS).FindStringSubmatch(note)
	if len(m) > 0 {
		return m[1] != "" || m[2] != "" || m[3] != ""
	}

	return false
}

func joinNotes(notes []string) string {
	lastIndex := len(notes) - 1
	var sb strings.Builder

	for i, current := range notes {
		if current == "" {
			continue
		}

		var next string
		if lastIndex >= i+1 {
			next = notes[i+1]
		} else {
			next = ""
		}

		var dot string
		if next != "" {
			if next[0] == '[' {
				dot = ""
			} else {
				dot = "."
			}
		} else {
			dot = ""
		}

		sb.WriteString(current + dot)
	}

	return sb.String()
}

func NewGlob(glob string) Glob {
	ins := inspect(Glob{glob: glob})
	notes := split(ins.AbsGlob, false)
	return Glob{
		glob:  glob,
		Notes: notes,
	}
}

func NewInspect(glob string) GlobInspect {
	ins := inspect(Glob{glob: glob})
	// notes := split(ins.AbsGlob, false)
	return ins
}

type Glob struct {
	Notes     []string
	isNegated bool
	glob      string
}

func (g Glob) GetGlob() string {
	return g.glob
}

func removeTrailingWildcards(glob Glob) Glob {
	// return glob.replace(/(.+?)(\.\*|\[\*\])*$/, '$1');

	glob.glob = strings.ReplaceAll(glob.glob, NON_NEG_WILDCARD_TRAIL, "$1")
	return glob
}

func split(glob Glob, normalize bool) []string {
	/*
	   if (!Glob.isValid(glob)) {
	       throw new NotationError(`${ERR_INVALID} '${glob}'`);
	   }
	*/
	neg := false
	if glob.glob[0] == '!' {
		neg = true
	}
	// trailing wildcards are redundant only when not negated
	var g = glob
	if !neg && normalize {
		g = removeTrailingWildcards(glob)
	}
	g.glob = strings.TrimPrefix(g.glob, "!")
	re := regexp.MustCompile(reMATCHER)
	ss := re.FindAllString(g.glob, -1)
	return ss
}

func intersect(globA, globB Glob, restrictive bool) string {
	var bang string
	notesA := split(globA, true)
	notesB := split(globB, true)

	if restrictive {
		if globA.glob[0] == '!' || globB.glob[0] == '!' {
			bang = "!"
		} else {
			bang = ""
		}
	} else {
		if globA.glob[0] == '!' && globB.glob[0] == '!' {
			bang = "!"
		} else if (len(notesA) > len(notesB) && globA.glob[0] == '!') || (len(notesB) > len(notesA) && globB.glob[0] == '!') {
			bang = "!"
		} else {
			bang = ""
		}
	}

	length := int(math.Max(float64(len(notesA)), float64(len(notesB))))
	notesI := make([]string, 0)
	var a, b string

	for i := 0; i < length; i++ {
		if len(notesA) > i {
			a = notesA[i]
		} else {
			a = ""
		}

		if len(notesB) > i {
			b = notesB[i]
		} else {
			b = ""
		}

		matcheda, _ := regexp.MatchString(WILDCARD, a)
		matchedb, _ := regexp.MatchString(WILDCARD, b)
		if a == b {
			notesI = append(notesI, a)
		} else if a != "" && matcheda {
			if b == "" {
				notesI = append(notesI, a)
			} else {
				notesI = append(notesI, b)
			}
		} else if b != "" && matchedb {
			if a == "" {
				notesI = append(notesI, b)
			} else {
				notesI = append(notesI, a)
			}
		} else if a != "" && b == "" {
			notesI = append(notesI, a)
		} else if a == "" && b != "" {
			notesI = append(notesI, b)
		} else {
			notesI = nil
			break
		}
	}

	if len(notesI) > 0 {
		return bang + joinNotes(notesI)
	}

	return ""
}

func invert(glob string) string {
	if glob[0] == '!' {
		return glob[1:]
	} else {
		return "!" + glob
	}
}
func checkAddIntersection(gA, gB Glob, restrictive bool, original, list []GlobInspect, intersections map[string]string) {
	inter := intersect(gA, gB, restrictive)
	if inter == "" {
		return
	}

	hasInverted := false
	if !restrictive {
		hasInverted = indexOf(original, invert(inter)) >= 0
	}

	if indexOf(list, inter) >= 0 || hasInverted {
		return
	}

	intersections[inter] = inter
}

func indexOf(slice []GlobInspect, value string) int {
	for i, v := range slice {
		if v.Glob.glob == value {
			return i
		}
	}
	return -1
}

/*
func checkAddIntersection(gA, gB Glob, restrictive bool) {
	inter := intersect(gA, gB, restrictive)
	if inter == "" {
		return
	}
	hasInverted := false
	if !restrictive {
		inverted := invert(inter)
		for _, orig := range globs {
			if orig.glob == inverted {
				hasInverted = true
				break
			}
		}
	}
	if _, exists := intersections[inter]; exists || hasInverted {
		return
	}
	intersections[inter] = inter
}
*/

var rx = regexp.MustCompile(`^\s*!`)

func negFirstSort(a, b string) bool {
	negA := rx.MatchString(a)
	negB := rx.MatchString(b)

	if negA && negB {
		if len(a) >= len(b) {
			return true
		} else {
			return false
		}
	} else if negA {
		return false
	} else if negB {
		return true
	} else {
		return true
	}
}

type GlobInspect struct {
	parentt     Glob
	Glob        Glob
	AbsGlob     Glob
	IsNegated   bool
	IsArrayGlob bool
}

func (g GlobInspect) String() string {
	return g.Glob.glob
}

func inspect(glob Glob) GlobInspect {
	g := strings.TrimSpace(glob.glob)
	glob.glob = g
	/*
		if !isValid(g) {
			return GlobInspect{}, fmt.Errorf("%s '%s'", ERR_INVALID, glob)
		}
	*/

	isNegated := g[0] == '!'
	if !isNegated {
		glob = removeTrailingWildcards(glob)
	}
	absGlob := glob
	if isNegated {
		absGlob.glob = g[1:]
	}

	return GlobInspect{
		Glob:      glob,
		AbsGlob:   absGlob,
		IsNegated: isNegated,
		// IsArrayGlob: isArrayGlob(absGlob),
	}
}

func negLastSort(a, b string) bool {
	negA := rx.MatchString(a)
	negB := rx.MatchString(b)

	if negA && negB {
		if len(a) >= len(b) {
			return true
		} else {
			return false
		}
	} else if negA {
		return true
	} else if negB {
		return false
	} else {
		return true
	}
}

/*
func (a Glob) String() string {
	return a.glob
}
func (a GlobInspect) String() string {
	return a.Glob.glob
}
*/

func Normalize(globList []Glob, restrictive bool) []string {
	var normalized = []string{}
	var ignored = make(map[string]bool)
	var intersections = make(map[string]string)

	// flags

	var negateAll bool

	// var duplicate = false
	var hasExactNeg = false
	var negCoversPos = false
	var negCoveredByPos = false
	var negCoveredByNeg = false
	var posCoversPos = false
	var posCoveredByNeg = false
	var posCoveredByPos = false
	var list []GlobInspect
	for _, glob := range globList {
		inspect := inspect(glob)
		list = append(list, inspect)
	}
	if restrictive {
		sort.Slice(list, func(i, j int) bool {
			return negFirstSort(list[i].Glob.glob, list[j].Glob.glob)
		})
	} else {
		sort.Slice(list, func(i, j int) bool {
			return negLastSort(list[i].Glob.glob, list[j].Glob.glob)
		})
	}

	if len(list) == 1 {
		g := list[0]
		if g.Glob.isNegated {
			return []string{}
		}
		return []string{g.Glob.glob}
	}

	normalized = []string{}
	ignored = make(map[string]bool)
	intersections = make(map[string]string)
	negateAll = false

	copyList := []GlobInspect{}
	for _, v := range list {
		copyList = append(copyList, v)
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	processed := map[string]struct{}{}
	w := 0
	for _, s := range list {
		if _, exists := processed[s.Glob.glob]; !exists {
			// If this city has not been seen yet, add it to the list
			processed[s.Glob.glob] = struct{}{}
			list[w] = s
			w++
		}
	}
	list = list[:w]
	for indexA := len(list) - 1; indexA >= 0; indexA-- {
		var duplicate bool
		a := list[indexA]

		// if `strict` is enabled, return empty if a negate-all is found
		// (which itself is also redundant if single): '!*' or '![*]'
		// Assuming re.NEGATE_ALL is a regular expression for matching negate-all patterns.
		// You need to define it accordingly in your Go code.
		// If restrictive is a variable, you need to define it as well.
		if matched, _ := regexp.MatchString(NEGATE_ALL, a.Glob.glob); matched {
			negateAll = true
			if restrictive {
				// return true
				break
			}
		}

		// flags
		for indexB := len(list) - 1; indexB >= 0; indexB-- {
			b := list[indexB]

			if indexA == indexB {
				// return false
				continue // move to next
			}

			// e.g. ['x.y.z', '[1].x', 'c'] » impossible! the tested source
			// object cannot be both an array and an object.
			if a.IsArrayGlob != b.IsArrayGlob {
				// Assuming NotationError is a custom error type you have defined.
				// You need to define it accordingly in your Go code.
				panic(fmt.Sprintf("Integrity failed. Cannot have both object and array notations for root level"))
			}

			// remove if duplicate
			if a.Glob.glob == b.Glob.glob {
				list = append(list[:indexA], list[indexA+1:]...)
				list = append(list[:indexB], list[indexB+1:]...)
				duplicate = true
				// return true
				break
			}

			// remove if positive has an exact negated (negated wins when
			// normalized) e.g. ['*', 'a', '!a'] => ['*', '!a']
			if !a.Glob.isNegated && isReverseOf(a, b) {
				ignored[a.Glob.glob] = true
				hasExactNeg = true
				// return true
				break
			}

			// if already excluded b, go on to next
			if ignored[b.Glob.glob] {
				continue
			}

			coversB := covers(a, b, restrictive)
			coveredByB := false
			if !coversB {
				coveredByB = covers(b, a, restrictive)
			}

			if a.IsNegated {
				if b.IsNegated {
					// if negated (a) covered by any other negated (b); remove (a)!
					if coveredByB {
						negCoveredByNeg = true
						ignored[a.Glob.glob] = true
						// return true
						break
					}
				} else {
					if coversB {
						negCoversPos = true
					}
					if coveredByB {
						negCoveredByPos = true
					}
					// try intersection if none covers the other and only
					// one of them is negated.
					if !coversB && !coveredByB {
						checkAddIntersection(a.Glob, b.Glob, restrictive, copyList, list, intersections)
					}
				}
			} else {
				if b.IsNegated {
					// if positive (a) covered by any negated (b); remove (a)!
					if coveredByB {
						posCoveredByNeg = true
						if restrictive {

							ignored[a.Glob.glob] = true
							// return true
							break
						}
						// return false
						continue // next
					}
					// try intersection if none covers the other and only
					// one of them is negated.
					if !coversB && !coveredByB {
						checkAddIntersection(a.Glob, b.Glob, restrictive, copyList, list, intersections)
					}
				} else {
					if coversB {
						posCoversPos = true
					}
					// if positive (a) covered by any other positive (b); remove (a)!
					if coveredByB {
						posCoveredByPos = true
						if restrictive {
							// return true
							break
						}
					}
				}
			}
		}
		var keepNeg bool
		var keepPos bool
		if restrictive {
			keepNeg = (negCoversPos || negCoveredByPos) && negCoveredByNeg == false
			keepPos = (posCoversPos || posCoveredByPos == false) && posCoveredByNeg == false
		} else {
			keepNeg = negCoveredByPos && negCoveredByNeg == false
			keepPos = posCoveredByNeg || posCoveredByPos == false
		}
		var t bool
		if a.Glob.isNegated {
			t = keepNeg
		} else {
			t = keepPos
		}
		var keep = duplicate == false && hasExactNeg == false && (t)
		if keep {
			normalized = append(normalized, a.Glob.glob)
		} else {
			ignored[a.Glob.glob] = true
		}
	}
	if restrictive && negateAll {
		return []string{}
	}
	keys := []string{}
	for k := range intersections {
		keys = append(keys, k)
	}
	if len(keys) > 0 {
		normalized = append(normalized, keys...)

		newNorms := []Glob{}
		for _, v := range normalized {
			newNorms = append(newNorms, Glob{glob: v})
		}
		return Normalize(newNorms, restrictive)
	}
	return sortNorm(normalized)
	// return normalized
}

func sortNorm(norms []string) []string {
	sort.Slice(norms, func(i, j int) bool {
		a := compare(norms[i], norms[j])
		if a == 1 {
			return false
		}
		return true
	})
	return norms
}

type callbackfunc func(GlobInspect, int) bool

func eachRight(array []GlobInspect, callback callbackfunc) {
	index := len(array)
	for index > 0 {
		index--
		if callback(array[index], index) == false {
			return
		}
	}
}

func isReverseOf(a, b GlobInspect) bool {
	return a.Glob.isNegated != b.Glob.isNegated &&
		a.Glob.glob == b.Glob.glob
}

func compare(globA, globB string) int {
	// Trivial case, both are exactly the same!
	// or both are wildcard e.g. `*` or `[*]`
	t1, _ := regexp.MatchString(WILDCARD, globB)
	t2, _ := regexp.MatchString(WILDCARD, globA)
	if globA == globB || t1 && t2 {
		return 0
	}

	ag := NewGlob(globA)
	bg := NewGlob(globB)
	a := inspect(ag)
	b := inspect(bg)

	// Check depth (number of levels)
	if len(a.Glob.Notes) == len(b.Glob.Notes) {
		// Check and compare if these are globs that represent items in the
		// "same" array. If not, this will return 0.
		aIdxCompare := compareArrayItemGlobs(a, b)
		// We'll only continue comparing if 0 is returned
		if aIdxCompare != 0 {
			return aIdxCompare
		}

		// Count wildcards
		reg1 := regexp.MustCompile(WILDCARD)
		wildCountA := len(reg1.FindAllString(a.AbsGlob.glob, -1))
		wildCountB := len(reg1.FindAllString(b.AbsGlob.glob, -1))
		if wildCountA == wildCountB {
			// Check for negation
			if !a.Glob.isNegated && b.Glob.isNegated {
				return -1
			}
			if a.Glob.isNegated && !b.Glob.isNegated {
				return 1
			}
			// Both are negated or neither are, return alphabetical
			return strings.Compare(a.AbsGlob.glob, b.AbsGlob.glob)
		}
		if wildCountA > wildCountB {
			return -1
		}
		return 1
	}

	if len(a.Glob.Notes) < len(b.Glob.Notes) {
		return -1
	}
	return 1
}

func (g GlobInspect) last() string {
	return g.Glob.Notes[len(g.Glob.Notes)-1]
}

func compareArrayItemGlobs(a, b GlobInspect) int {
	// Both should be negated
	reg1, _ := regexp.MatchString(ARRAY_GLOB_NOTE, a.last())
	reg2, _ := regexp.MatchString(ARRAY_GLOB_NOTE, b.last())
	if !a.Glob.isNegated || !b.Glob.isNegated ||
		// Should be the same length (since we're comparing for items in the same array)
		len(a.Glob.Notes) != len(b.Glob.Notes) ||
		// Last notes should be array brackets

		reg1 || reg2 ||
		// Last notes should be different to compare
		a.last() == b.last() {
		return 0
	}

	// Negated !..[*] should come last
	if a.last() == "[*]" {
		return 1 // b is first
	}
	if b.last() == "[*]" {
		return -1 // a is first
	}

	if a.parent().Glob.glob != "" && b.parent().Glob.glob != "" {
		if covers(a.parent(), b.parent(), true) {
			return compArrIdx(a.last(), b.last())
		}
		return 0
	}
	return compArrIdx(a.last(), b.last())
}
func (g GlobInspect) parent() GlobInspect {
	tem := Glob{}
	retGlob := GlobInspect{}
	// Setting on first call instead of in the constructor, for performance optimization.
	if g.parentt.glob == "" {
		if len(g.Glob.Notes) > 1 {
			tem.glob = strings.TrimSuffix(g.AbsGlob.glob[:len(g.AbsGlob.glob)-len(g.last())], ".")
			retGlob = inspect(tem)
		} else {
			tem.glob = ""
			g.parentt = tem
			retGlob = inspect(tem)
		}
	}
	return retGlob
}

func idxVal(note string) int64 {
	// we return -1 for wildcard bec. we need it to come last

	// below will never execute when called from _compareArrayItemGlobs
	/* istanbul ignore next */
	// if (note === '[*]') return -1;

	// e.g. '[2]' » 2
	str := strings.ReplaceAll(note, `[[\]]`, "")
	a, _ := strconv.ParseInt(str, 10, 64)
	return a
}
func compArrIdx(lastA, lastB string) int {
	var iA = idxVal(lastA)
	var iB = idxVal(lastB)

	// below will never execute when called from _compareArrayItemGlobs
	/* istanbul ignore next */
	// if (iA === iB) return 0;
	if iA > iB {
		return -1
	}
	return 1
}
