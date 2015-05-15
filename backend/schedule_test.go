package main

import (
	"reflect"
	"testing"
	"time"
)

func TestSubslice(t *testing.T) {
	table := []struct{ in, items, out []string }{
		{[]string{"a", "b", "c"}, []string{"a", "c"}, []string{"b"}},
		{[]string{"a", "b", "c"}, []string{"a", "c", "b"}, []string{}},
		{[]string{"a", "b", "c"}, []string{"d"}, []string{"a", "b", "c"}},
		{[]string{"b", "c"}, []string{"a", "c"}, []string{"b"}},
		{[]string{"a", "b", "c"}, []string{}, []string{"a", "b", "c"}},
		{[]string{"abc", "def"}, []string{"ab"}, []string{"abc", "def"}},
	}
	for i, test := range table {
		out := subslice(test.in, test.items...)
		if !reflect.DeepEqual(out, test.out) {
			t.Errorf("%d: subslice(%v, %v) = %v; want %v", i, test.in, test.items, out, test.out)
		}
	}
}

func TestUnique(t *testing.T) {
	table := []struct{ in, out []string }{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"a", "b", "b", "a", "d"}, []string{"a", "b", "d"}},
		{[]string{"a", "a", "a"}, []string{"a"}},
		{[]string{}, []string{}},
	}
	for i, test := range table {
		out := unique(test.in)
		if !reflect.DeepEqual(out, test.out) {
			t.Errorf("%d: unique(%v) = %v; want %v", i, test.in, out, test.out)
		}
	}
}

func TestDiffEventData(t *testing.T) {
	a := &eventSession{
		Title:     "Keynote",
		StartTime: time.Date(2015, 5, 28, 9, 30, 0, 0, time.UTC),
		Tags:      []string{"FLAG_KEYNOTE"},
		Filters:   map[string]bool{"Live streamed": true},
	}
	b := &eventSession{
		Title:     "Keynote",
		StartTime: time.Date(2015, 5, 28, 9, 30, 0, 0, time.UTC),
		Tags:      []string{"FLAG_KEYNOTE"},
		Filters:   map[string]bool{"Live streamed": true},
		Speakers:  []string{},
	}
	dc := diffEventData(
		&eventData{Sessions: map[string]*eventSession{"__keynote__": a}},
		&eventData{Sessions: map[string]*eventSession{"__keynote__": b}},
	)
	if l := len(dc.Sessions); l != 0 {
		t.Errorf("len(dc.Sessions) = %d; want 0", l)
	}
}

func TestDiffEventDataVideo(t *testing.T) {
	date := time.Now().Round(time.Second)
	past := date.Add(-time.Hour)
	future := date.Add(time.Hour)

	table := []struct {
		end1, end2   time.Time
		live1, live2 bool
		yt1, yt2     string
		diff         string
	}{
		// past sessions
		{past, past, true, false, "live", "recored", updateVideo},
		{past, past, true, false, "", "recored", updateVideo},
		{past.Add(-time.Hour), past, true, false, "live", "recored", updateVideo},
		{past.Add(-time.Hour), past, true, false, "", "recored", updateVideo},
		{past, past, false, false, "", "recored", updateVideo},
		{past, past, false, false, "recorded1", "recored2", updateVideo},
		{past, past, false, false, "recorded1", "", ""},
		{past, past, false, true, "", "live", ""},
		{past, past, false, true, "recorded", "live", ""},
		{past, past, false, true, "recorded", "", ""},
		{past, past, true, false, "live", "", ""},
		{past, past, true, true, "", "live", ""},
		{past, past, true, true, "live1", "live2", ""},
		{past, past, true, true, "live1", "", ""},
		// future sessions; i = 14
		{future, future, true, false, "", "", ""},
		{future, future, true, false, "live", "", ""},
		{future, future, true, false, "", "recorded", ""},
		{future, future, true, false, "live", "recorded", ""},
		{future, future, false, true, "", "", ""},
		{future, future, false, true, "live", "", ""},
		{future, future, false, true, "", "recorded", ""},
		{future, future, false, true, "live", "recorded", ""},
		{future, future, true, true, "live1", "live2", ""},
		{future, future, true, true, "live", "", ""},
		{future, future, true, true, "", "live", ""},
		{future, future, false, false, "live1", "live2", ""},
		{future, future, false, false, "live", "", ""},
		{future, future, false, false, "", "live", ""},
	}
	for i, test := range table {
		a := &eventSession{
			EndTime: test.end1,
			IsLive:  test.live1,
			YouTube: test.yt1,
		}
		b := &eventSession{
			EndTime: test.end2,
			IsLive:  test.live2,
			YouTube: test.yt2,
		}
		dc := diffEventData(
			&eventData{Sessions: map[string]*eventSession{"id": a}},
			&eventData{Sessions: map[string]*eventSession{"id": b}},
		)
		switch {
		case test.diff == "" && len(dc.Sessions) != 0:
			t.Errorf("%d: diff(%v, %q, %v, %q) = %q; want 0 sessions",
				i, test.live1, test.yt1, test.live2, test.yt2, dc.Sessions["id"].Update)
		case test.diff != "" && len(dc.Sessions) == 0:
			t.Errorf("%d: 0 sessions; want b.Update = %q", i, test.diff)
		case test.diff != "" && len(dc.Sessions) != 0:
			if up := dc.Sessions["id"].Update; up != test.diff {
				t.Errorf("%d: diff(%v, %q, %v, %q) = %q; want %q",
					i, test.live1, test.yt1, test.live2, test.yt2, up, test.diff)
			}
		}
		if b.EndTime != test.end2 {
			t.Errorf("%d: b.EndTime = %v; want %v", b.EndTime, test.end2)
		}
		if b.IsLive != test.live2 {
			t.Errorf("%d: b.IsLive = %v; want %v", b.IsLive, test.live2)
		}
		if b.YouTube != test.yt2 {
			t.Errorf("%d: b.YouTube = %v; want %v", b.YouTube, test.yt2)
		}
	}
}

func TestThumbURL(t *testing.T) {
	table := []struct{ in, out string }{
		{"http://example.org/image.jpg", "http://example.org/image.jpg"},
		{"http://example.org/images/__w/img.jpg", "http://example.org/images/__w/img.jpg"},
		{"http://example.org/images/__w-400-600/img.jpg", "http://example.org/images/w400/img.jpg"},
		{"http://example.org/__w-200-400-600-800-1000/img.jpg", "http://example.org/w200/img.jpg"},
	}
	for _, test := range table {
		out := thumbURL(test.in)
		if out != test.out {
			t.Errorf("thumbURL(%q) = %q; want %q", test.in, out, test.out)
		}
	}
}
