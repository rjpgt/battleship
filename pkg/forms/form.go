package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Regexp patterns for the various ships
var (
	BtlshipRX    = regexp.MustCompile(`^(\s*\d{2}\s*,){4}\s*\d{2}\s*$`)
	CruiserRX    = regexp.MustCompile(`^(\s*\d{2}\s*,){3}\s*\d{2}\s*$`)
	FrigateRX    = regexp.MustCompile(`^(\s*\d{2}\s*,){2}\s*\d{2}\s*$`)
	DestroyerRX  = regexp.MustCompile(`^(\s*\d{2}\s*,){2}\s*\d{2}\s*$`)
	PatrolboatRX = regexp.MustCompile(`^(\s*\d{2}\s*,){1}\s*\d{2}\s*$`)
	FirePosRX    = regexp.MustCompile(`^\s*\d{2}\s*$`)
)

// Form embeds an anonymous url.Values object
// to hold the form data and an Errors field
// to hold any validation errors.
type Form struct {
	url.Values
	Errors errors
}

// New initializes a custom Form struct. This
// takes the form data as a parameter.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required validates form fields that should
// not be blank.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MaxLength validates a field that has a maximum character limit.
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

// PermittedValues checks that a field in a form matches a set of permitted values.
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

// MinLength checks that a field has a given minimum length
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))

	}
}

// MatchesPattern checks a ship positions field has the right pattern matching the ship
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// HorizOrVert checks that a ship is placed horizontally or vertically
func (f *Form) HorizOrVert(field string) {
	posns := strings.Split(f.Get(field), ",")
	nums := make([]int, len(posns))
	rowNums := make([]int, len(posns))
	colNums := make([]int, len(posns))
	for i := range posns {
		nums[i], _ = strconv.Atoi(strings.TrimSpace(posns[i]))
	}
	for i := range rowNums {
		rowNums[i] = nums[0] + i
	}
	for i := range colNums {
		colNums[i] = nums[0] + i*10
	}
	if !sliceEq(nums, rowNums) && !sliceEq(nums, colNums) {
		f.Errors.Add(field, "Ship must be placed horizontally or vertically")
	}
}

// NonOverlapping checks that the ship positions do not overlap
func (f *Form) NonOverlapping(fields ...string) {
	squareCount := map[string]int{}
	for _, field := range fields {
		posns := strings.Split(f.Get(field), ",")
		for _, posn := range posns {
			squareCount[posn]++
			if squareCount[posn] == 2 {
				f.Errors.Add(field, fmt.Sprintf("%s is overlapping", posn))
			}
		}
	}
}

// ValidateNewGameForm validates the entire form for a new game
func (f *Form) ValidateNewGameForm() {
	f.Required("username", "btlship", "cruiser", "frigate", "destroyer", "patrolboat")
	f.MinLength("username", 4)
	f.MaxLength("username", 10)
	f.MatchesPattern("btlship", BtlshipRX)
	f.HorizOrVert("btlship")
	f.MatchesPattern("cruiser", CruiserRX)
	f.HorizOrVert("cruiser")
	f.MatchesPattern("frigate", FrigateRX)
	f.HorizOrVert("frigate")
	f.MatchesPattern("destroyer", DestroyerRX)
	f.HorizOrVert("destroyer")
	f.MatchesPattern("patrolboat", PatrolboatRX)
	f.HorizOrVert("patrolboat")
	f.NonOverlapping("btlship", "cruiser", "frigate", "destroyer", "patrolboat")
}

// ValidateFireForm validates the fire position coordinates
func (f *Form) ValidateFireForm() {
	f.Required("target_pos")
	f.MatchesPattern("target_pos", FirePosRX)
}

// Valid validates the form
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

func sliceEq(s1, s2 []int) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
