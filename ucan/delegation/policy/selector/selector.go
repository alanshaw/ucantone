package selector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/alanshaw/ucantone/ipld"
)

// Selector syntax is closely based on jq's "filters". They operate on an
// Invocation's args object.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#selectors
type Selector []Segment

func (s Selector) String() string {
	var b strings.Builder
	for _, seg := range s {
		b.WriteString(seg.String())
	}
	return b.String()
}

func (s Selector) MarshalJSON() ([]byte, error) {
	if len(s) == 0 {
		return json.Marshal(nil)
	}
	return json.Marshal(s.String())
}

func (s *Selector) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return fmt.Errorf("parsing string: %w", err)
	}
	if str == "" {
		return nil
	}
	parsed, err := Parse(str)
	if err != nil {
		return fmt.Errorf("parsing selector: %w", err)
	}
	*s = parsed
	return nil
}

var Identity = Segment{".", true, false, false, nil, "", 0}

var (
	indexRegex = regexp.MustCompile(`^-?\d+$`)
	sliceRegex = regexp.MustCompile(`^((\-?\d+:\-?\d*)|(\-?\d*:\-?\d+))$`)
	fieldRegex = regexp.MustCompile(`^\.[a-zA-Z_]*?$`)
)

type Segment struct {
	str      string
	Identity bool   // Identity flags that this selector is the identity selector.
	Optional bool   // Optional flags that this selector is optional.
	Iterator bool   // Iterator flags that this selector is an iterator segment.
	Slice    []int  // Slice flags that this segment targets a range of a slice.
	Field    string // Field is the name of a field in a struct/map.
	Index    int    // Index is an index of a slice.
}

// String returns the segment's string representation.
func (s Segment) String() string {
	return s.str
}

func Parse(str string) (Selector, error) {
	if string(str[0]) != "." {
		return nil, NewParseError("selector must start with identity segment '.'", str, 0, string(str[0]))
	}

	col := 0
	var sel Selector
	for _, tok := range tokenize(str) {
		seg := tok
		opt := strings.HasSuffix(tok, "?")
		if opt {
			seg = tok[0 : len(tok)-1]
		}
		switch seg {
		case ".":
			if len(sel) > 0 && sel[len(sel)-1].Identity {
				return nil, NewParseError("selector contains unsupported recursive descent segment: '..'", str, col, tok)
			}
			sel = append(sel, Identity)
		case "[]":
			sel = append(sel, Segment{tok, false, opt, true, nil, "", 0})
		default:
			if strings.HasPrefix(seg, "[") && strings.HasSuffix(seg, "]") {
				lookup := seg[1 : len(seg)-1]

				if indexRegex.MatchString(lookup) { // index
					idx, err := strconv.Atoi(lookup)
					if err != nil {
						return nil, NewParseError("invalid index", str, col, tok)
					}
					sel = append(sel, Segment{str: tok, Optional: opt, Index: idx})
				} else if strings.HasPrefix(lookup, "\"") && strings.HasSuffix(lookup, "\"") { // explicit field
					sel = append(sel, Segment{str: tok, Optional: opt, Field: lookup[1 : len(lookup)-1]})
				} else if sliceRegex.MatchString(lookup) { // slice [3:5] or [:5] or [3:]
					var rng []int
					splt := strings.Split(lookup, ":")
					if splt[0] == "" {
						rng = append(rng, 0)
					} else {
						i, err := strconv.Atoi(splt[0])
						if err != nil {
							return nil, NewParseError("invalid slice index", str, col, tok)
						}
						rng = append(rng, i)
					}
					if splt[1] != "" {
						i, err := strconv.Atoi(splt[1])
						if err != nil {
							return nil, NewParseError("invalid slice index", str, col, tok)
						}
						rng = append(rng, i)
					}
					sel = append(sel, Segment{str: tok, Optional: opt, Slice: rng})
				} else {
					return nil, NewParseError(fmt.Sprintf("invalid segment: %s", seg), str, col, tok)
				}
			} else if fieldRegex.MatchString(seg) {
				sel = append(sel, Segment{str: tok, Optional: opt, Field: seg[1:]})
			} else {
				return nil, NewParseError(fmt.Sprintf("invalid segment: %s", seg), str, col, tok)
			}
		}
		col += len(tok)
	}
	return sel, nil
}

func tokenize(str string) []string {
	var toks []string
	col := 0
	ofs := 0
	ctx := ""

	for col < len(str) {
		char := string(str[col])

		if char == "\"" && string(str[col-1]) != "\\" {
			col++
			if ctx == "\"" {
				ctx = ""
			} else {
				ctx = "\""
			}
			continue
		}

		if ctx == "\"" {
			col++
			continue
		}

		if char == "." || char == "[" {
			if ofs < col {
				toks = append(toks, str[ofs:col])
			}
			ofs = col
		}
		col++
	}

	if ofs < col && ctx != "\"" {
		toks = append(toks, str[ofs:col])
	}

	return toks
}

// Select uses a selector to extract a value from the passed subject.
func Select(sel Selector, subject any) (any, error) {
	return resolve(sel, subject, nil)
}

func resolve(sel Selector, subject any, at []string) (any, error) {
	cur := subject
	for i, seg := range sel {
		if seg.Identity {
			continue
		} else if seg.Iterator {
			curType := reflect.TypeOf(cur)
			if curType != nil && curType.Kind() == reflect.Slice {
				var many []any
				v := reflect.ValueOf(cur)
				for k := range v.Len() {
					key := fmt.Sprintf("%d", k)
					r, err := resolve(sel[i+1:], v.Index(k).Interface(), append(at[:], key))
					if err != nil {
						return nil, err
					}
					if r == nil {
						continue
					}
					rTyp := reflect.TypeOf(r)
					if rTyp != nil {
						if rTyp.Kind() == reflect.Slice {
							rVal := reflect.ValueOf(r)
							for j := range rVal.Len() {
								many = append(many, rVal.Index(j).Interface())
							}
						} else {
							many = append(many, r)
						}
					}
				}
				cur = many
			} else if m, ok := cur.(ipld.Map); ok {
				var many []any
				for k, v := range m {
					r, err := resolve(sel[i+1:], v, append(at[:], k))
					if err != nil {
						return nil, err
					}
					if r == nil {
						continue
					}
					rTyp := reflect.TypeOf(r)
					if rTyp != nil {
						if rTyp.Kind() == reflect.Slice {
							rVal := reflect.ValueOf(r)
							for j := range rVal.Len() {
								many = append(many, rVal.Index(j).Interface())
							}
						} else {
							many = append(many, r)
						}
					}
				}
				cur = many
			} else if seg.Optional {
				cur = nil
			} else {
				return nil, NewResolutionError(fmt.Sprintf("can not iterate over type: %s", reflect.TypeOf(cur)), at)
			}
		} else if seg.Field != "" {
			at = append(at, seg.Field)
			if m, ok := cur.(ipld.Map); ok {
				v, ok := m[seg.Field]
				if !ok && !seg.Optional {
					return nil, NewResolutionError(fmt.Sprintf("object has no field named: %s", seg.Field), at)
				}
				cur = v
			} else if seg.Optional {
				cur = nil
			} else {
				return nil, NewResolutionError(fmt.Sprintf("can not access field: %s on type: %s", seg.Field, reflect.TypeOf(cur)), at)
			}
		} else if seg.Slice != nil {
			curType := reflect.TypeOf(cur)
			if curType != nil && curType.Kind() == reflect.Slice {
				v := reflect.ValueOf(cur)
				idxs := seg.Slice
				if len(idxs) == 1 {
					idxs = []int{idxs[0], v.Len()}
				}
				if idxs[1] > v.Len() {
					return nil, NewResolutionError(fmt.Sprintf("index out of bounds: %d", idxs[1]), at)
				}
				cur = v.Slice(idxs[0], idxs[1]).Interface()
			} else if curStr, ok := cur.(string); ok {
				idxs := seg.Slice
				if len(idxs) == 1 {
					idxs = []int{idxs[0], len(curStr)}
				}
				if idxs[1] > len(curStr) {
					return nil, NewResolutionError(fmt.Sprintf("index out of bounds: %d", idxs[1]), at)
				}
				cur = curStr[idxs[0]:idxs[1]]
			} else if seg.Optional {
				cur = nil
			} else {
				return nil, NewResolutionError(fmt.Sprintf("can not index: %s on type: %s", seg.Field, reflect.TypeOf(cur)), at)
			}
		} else {
			at = append(at, fmt.Sprintf("%d", seg.Index))
			curType := reflect.TypeOf(cur)
			if curType != nil && curType.Kind() == reflect.Slice {
				v := reflect.ValueOf(cur)
				idx := seg.Index
				if idx < 0 {
					idx = v.Len() + idx
				}
				if idx < 0 || idx >= v.Len() {
					if seg.Optional {
						cur = nil
					} else {
						return nil, NewResolutionError(fmt.Sprintf("index out of bounds: %d", seg.Index), at)
					}
				} else {
					cur = v.Index(idx).Interface()
				}
			} else if curStr, ok := cur.(string); ok {
				idx := seg.Index
				if idx < 0 {
					idx = len(curStr) + idx
				}
				if idx < 0 || idx >= len(curStr) {
					if seg.Optional {
						cur = nil
					} else {
						return nil, NewResolutionError(fmt.Sprintf("index out of bounds: %d", seg.Index), at)
					}
				} else {
					cur = string(curStr[idx])
				}
			} else if seg.Optional {
				cur = nil
			} else {
				return nil, NewResolutionError(fmt.Sprintf("can not access field: %s on type: %s", seg.Field, reflect.TypeOf(cur)), at)
			}
		}
	}

	curType := reflect.TypeOf(cur)
	// if cur is a slice, we need to return it as a typed slice if all items
	// have the same type.
	if curType != nil && curType.Kind() == reflect.Slice {
		v := reflect.ValueOf(cur)
		var itemsType reflect.Type
		hasCommonType := true
		for i := range v.Len() {
			typ := reflect.TypeOf(v.Index(i).Interface())
			if hasCommonType {
				// first iteration (or all nil)
				if itemsType == nil {
					itemsType = typ
				} else if itemsType != typ {
					hasCommonType = false
					itemsType = nil
					break
				}
			}
		}

		// if all items have the same type and the type is not nil, create and
		// return a typed slice
		if hasCommonType && itemsType != nil {
			sliceType := reflect.SliceOf(itemsType)
			sliceValue := reflect.MakeSlice(sliceType, v.Len(), v.Len())
			for i := range v.Len() {
				sliceValue.Index(i).Set(reflect.ValueOf(v.Index(i).Interface()))
			}
			return sliceValue.Interface(), nil
		}
	}

	return cur, nil
}
