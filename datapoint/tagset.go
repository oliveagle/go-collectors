package datapoint

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	// "unicode/utf8"
	"bytes"
)

// TagSet is a helper class for tags.
type TagSet map[string]string

// Copy creates a new TagSet from t.
func (t TagSet) Copy() TagSet {
	n := make(TagSet)
	for k, v := range t {
		n[k] = v
	}
	return n
}

// Merge adds or overwrites everything from o into t and returns t.
func (t TagSet) Merge(o TagSet) TagSet {
	for k, v := range o {
		t[k] = v
	}
	return t
}

// Equal returns true if t and o contain only the same k=v pairs.
func (t TagSet) Equal(o TagSet) bool {
	if len(t) != len(o) {
		return false
	}
	for k, v := range t {
		if ov, ok := o[k]; !ok || ov != v {
			return false
		}
	}
	return true
}

// Subset returns true if all k=v pairs in o are in t.
func (t TagSet) Subset(o TagSet) bool {
	for k, v := range o {
		if tv, ok := t[k]; !ok || tv != v {
			return false
		}
	}
	return true
}

// Intersection returns the intersection of t and o.
func (t TagSet) Intersection(o TagSet) TagSet {
	r := make(TagSet)
	for k, v := range t {
		if o[k] == v {
			r[k] = v
		}
	}
	return r
}

// String converts t to an OpenTSDB-style {a=b,c=b} string, alphabetized by key.
func (t TagSet) String() string {
	return fmt.Sprintf("{%s}", t.Tags())
}

// Tags is identical to String() but without { and }.
func (t TagSet) Tags() string {
	var keys []string
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := &bytes.Buffer{}
	for i, k := range keys {
		if i > 0 {
			fmt.Fprint(b, ",")
		}
		fmt.Fprintf(b, "%s=%s", k, t[k])
	}
	return b.String()
}

func (d *DataPoint) clean() error {
	if err := d.Tags.Clean(); err != nil {
		return err
	}
	m, err := Clean(d.Metric)
	if err != nil {
		return fmt.Errorf("cleaning metric %s: %s", d.Metric, err)
	}
	if d.Metric != m {
		d.Metric = m
	}
	switch v := d.Value.(type) {
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			d.Value = i
		} else if f, err := strconv.ParseFloat(v, 64); err == nil {
			d.Value = f
		} else {
			return fmt.Errorf("Unparseable number %v", v)
		}
	case uint64:
		if v > math.MaxInt64 {
			d.Value = float64(v)
		}
	case *big.Int:
		if bigMaxInt64.Cmp(v) < 0 {
			if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
				d.Value = f
			}
		}
	}
	return nil
}

var bigMaxInt64 = big.NewInt(math.MaxInt64)

// Clean removes characters from t that are invalid for OpenTSDB metric and tag
// values. An error is returned if a resulting tag is empty.
func (t TagSet) Clean() error {
	for k, v := range t {
		kc, err := Clean(k)
		if err != nil {
			return fmt.Errorf("cleaning tag %s: %s", k, err)
		}
		vc, err := Clean(v)
		if err != nil {
			return fmt.Errorf("cleaning key %s: %s", v, err)
		}
		if kc != k || vc != v {
			delete(t, k)
			t[kc] = vc
		}
	}
	return nil
}

// ValidTag returns true if s is a valid metric or tag.
func ValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case strings.ContainsAny(string(c), `-_./`):
		case unicode.IsLetter(c):
		default:
			return false
		}
	}
	return true
}

// ParseTags parses OpenTSDB tagk=tagv pairs of the form: k=v,m=o. Validation
// errors do not stop processing, and will return a non-nil TagSet.
func ParseTags(t string) (TagSet, error) {
	ts := make(TagSet)
	var err error
	for _, v := range strings.Split(t, ",") {
		sp := strings.SplitN(v, "=", 2)
		if len(sp) != 2 {
			return nil, fmt.Errorf("opentsdb: bad tag: %s", v)
		}
		for i, s := range sp {
			sp[i] = strings.TrimSpace(s)
			if i > 0 {
				continue
			}
			if !ValidTag(sp[i]) {
				err = fmt.Errorf("invalid character in %s", sp[i])
			}
		}
		for _, s := range strings.Split(sp[1], "|") {
			if s == "*" {
				continue
			}
			if !ValidTag(s) {
				err = fmt.Errorf("invalid character in %s", sp[1])
			}
		}
		if _, present := ts[sp[0]]; present {
			return nil, fmt.Errorf("opentsdb: duplicated tag: %s", v)
		}
		ts[sp[0]] = sp[1]
	}
	return ts, err
}

var groupRE = regexp.MustCompile("{[^}]+}")

// ReplaceTags replaces all tag-like strings with tags from the given
// group. For example, given the string "test.metric{host=*}" and a TagSet
// with host=test.com, this returns "test.metric{host=test.com}".
func ReplaceTags(text string, group TagSet) string {
	return groupRE.ReplaceAllStringFunc(text, func(s string) string {
		tags, err := ParseTags(s[1 : len(s)-1])
		if err != nil {
			return s
		}
		for k := range tags {
			if group[k] != "" {
				tags[k] = group[k]
			}
		}
		return fmt.Sprintf("{%s}", tags.Tags())
	})
}
