package main

import (
	"reflect"
	"testing"
)

type MapData struct {
	data   map[string][]string
	expRes bool
	// True if error expected (err != nil), false otherwise
	expErr bool
}

func TestCheckOutputJobs(t *testing.T) {
	sample0 := map[string][]string{}
	sample1 := map[string][]string{"Title": {"t1", "t2", "t3"}, "Company": {"c1", "c2", "c3"}, "Job Desc": {"j1", "j2", "j3"}}
	sample2 := map[string][]string{"Title": {"t1", "t2"}, "Company": {"c1", "c2", "c3"}, "Job Desc": {"j1", "j2", "j3"}}
	sample3 := map[string][]string{"Title": {"t1", "t2", "t3"}, "Company": {"c1", "c2"}, "Job Desc": {"j1", "j2", "j3"}}
	sample4 := map[string][]string{"Title": {"t1", "t2", "t3"}, "Company": {"c1", "c2", "c3"}, "Job Desc": {"j1", "j2"}}
	testData := []MapData{
		{data: sample0, expRes: false, expErr: true},
		{data: sample1, expRes: true, expErr: false},
		{data: sample2, expRes: false, expErr: true},
		{data: sample3, expRes: false, expErr: true},
		{data: sample4, expRes: false, expErr: true},
	}
	for _, elem := range testData {
		get, getErr := checkScrapedJobs(elem.data)
		if get != elem.expRes {
			t.Errorf("CheckOutputJobs() FAILED: Expected %t, Actual: %t", get, elem.expRes)
		}
		if getErr == nil && elem.expErr {
			t.Errorf("CheckOutputJobs() FAILED: Expected: error to trigger, Actual: no error")
		}
		if getErr != nil && !elem.expErr {
			t.Errorf("CheckOutputJobs() FAILED: Expected: no error, Actual: error triggered")
		}
	}
}

type FilterData struct {
	data     []map[string]string
	filter   []string
	expected []map[string]string
	expErr   bool
}

func TestAndFilter(t *testing.T) {
	sample := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l b l"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l B l"},
	}

	filter0 := []string{}
	filter1 := []string{"a"}
	filter2 := []string{"a", "b"}
	filter3 := []string{"b", "a"}
	filter4 := []string{"b", "i"}
	filter5 := []string{"g", "i"}

	expected0 := []map[string]string{}
	expected1 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
	}

	expected2 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
	}

	expected3 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
	}
	expected4 := []map[string]string{
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
	}

	expected5 := []map[string]string{}

	testData := []FilterData{
		{sample, filter0, expected0, true},
		{sample, filter1, expected1, false},
		{sample, filter2, expected2, false},
		{sample, filter3, expected3, false},
		{sample, filter4, expected4, false},
		{sample, filter5, expected5, false},
	}
	i := 0
	for _, elem := range testData {
		get, getErr := andFilter(elem.filter, elem.data)
		if len(elem.expected) != len(get) {
			t.Errorf("AndFilter() FAILED, Test No: %d, Expected: %v Actual: %v", i, elem.expected, get)
		} else if !(reflect.DeepEqual(get, elem.expected)) && len(elem.expected) != 0 {
			t.Errorf("AndFilter() FAILED, Test No: %d, Expected: %v Actual: %v", i, elem.expected, get)
		}
		if elem.expErr && getErr == nil {
			t.Errorf("AndFilter() FAILED: Expected: error to trigger, Actual: no error")
		}
		if !elem.expErr && getErr != nil {
			t.Errorf("AndFilter() FAILED: Expected: no error, Actual: error triggered")
		}
		i++
	}
}

func TestOrFilter(t *testing.T) {
	sample := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l b l"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l B l"},
	}

	filter0 := []string{}
	filter1 := []string{"a"}
	filter2 := []string{"a", "b"}
	filter3 := []string{"b", "a"}
	filter4 := []string{"b", "i"}
	filter5 := []string{"g", "i"}

	expected0 := []map[string]string{}
	expected1 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
	}

	expected2 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l b l"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l B l"},
	}

	expected3 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l b l"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l B l"},
	}

	expected4 := []map[string]string{
		{"a": "a a a", "b": "b b b", "JobDescription": "a b c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "a B c"},
		{"a": "a a a", "b": "b b b", "JobDescription": "A b c"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l b l"},
		{"k": "k k k", "l": "l l l", "JobDescription": "l B l"},
	}

	expected5 := []map[string]string{
		{"e": "e e e", "f": "f f f", "JobDescription": "a g f"},
		{"e": "e e e", "f": "f f f", "JobDescription": "A g f"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i b"},
		{"h": "h h h", "i": "i i i", "JobDescription": "a i B"},
		{"h": "h h h", "i": "i i i", "JobDescription": "b i a"},
	}

	testData := []FilterData{
		{sample, filter0, expected0, true},
		{sample, filter1, expected1, false},
		{sample, filter2, expected2, false},
		{sample, filter3, expected3, false},
		{sample, filter4, expected4, false},
		{sample, filter5, expected5, false},
	}
	i := 0
	for _, elem := range testData {
		get, getErr := orFilter(elem.filter, elem.data)
		if len(elem.expected) != len(get) {
			t.Errorf("OrFilter() FAILED, Test No: %d, Expected: %v Actual: %v", i, elem.expected, get)
		} else if !(reflect.DeepEqual(get, elem.expected)) && len(elem.expected) != 0 {
			t.Errorf("OrFilter() FAILED, Test No: %d, Expected: %v Actual: %v", i, elem.expected, get)
		}
		if elem.expErr && getErr == nil {
			t.Errorf("OrFilter() FAILED: Expected: error to trigger, Actual: no error")
		}
		if !elem.expErr && getErr != nil {
			t.Errorf("OrFilter() FAILED: Expected: no error, Actual: error triggered")
		}
		i++
	}
}

type DataConv struct {
	data     map[string][]string
	expected []map[string]string
	expErr   bool
}

func TestDataConversion(t *testing.T) {
	data0 := map[string][]string{}
	data1 := map[string][]string{"example": {"a", "b", "c"}}
	data2 := map[string][]string{
		"example": {"a", "b", "c"},
		"hello":   {"v", "d", "t"},
		"yum":     {"e", "r", "p"},
	}
	data3 := map[string][]string{
		"pie":   {"123", "34", ""},
		"hack":  {"9ie", "", "3"},
		"pizza": {"", "2", "23"},
	}

	expect0 := []map[string]string{}
	expect1 := []map[string]string{
		{"example": "a"},
		{"example": "b"},
		{"example": "c"},
	}
	expect2 := []map[string]string{
		{"example": "a", "hello": "v", "yum": "e"},
		{"example": "b", "hello": "d", "yum": "r"},
		{"example": "c", "hello": "r", "yum": "p"},
	}
	expect3 := []map[string]string{
		{"pie": "123", "hack": "9ie", "pizza": ""},
		{"pie": "34", "hack": "", "pizza": "2"},
		{"pie": "", "hack": "3", "pizza": "23"},
	}

	testData := []DataConv{
		{data0, expect0, true},
		{data1, expect1, false},
		{data2, expect2, false},
		{data3, expect3, false},
	}

	i := 0
	for _, elem := range testData {
		var get []map[string]string
		get, getErr := dataConversion(elem.data, elem.expected)
		if len(elem.expected) != len(get) {
			t.Errorf("DataConversion() FAILED, Test No: %d, Expected: %v Actual: %v", i, elem.expected, get)
		} else if !(reflect.DeepEqual(get, elem.expected)) && len(elem.expected) != 0 {
			t.Errorf("DataConversion() FAILED, Test No: %d, Expected: %v Actual: %v", i, elem.expected, get)
		}
		if elem.expErr && getErr == nil {
			t.Errorf("DataConversion() FAILED: Expected: error to trigger, Actual: no error")
		}
		if !elem.expErr && getErr != nil {
			t.Errorf("DataConversion() FAILED: Expected: no error, Actual: error triggered")
		}
		i++
	}
}
