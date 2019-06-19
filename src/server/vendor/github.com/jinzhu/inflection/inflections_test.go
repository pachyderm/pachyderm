package inflection

import (
	"strings"
	"testing"
)

var inflections = map[string]string{
	"star":        "stars",
	"STAR":        "STARS",
	"Star":        "Stars",
	"bus":         "buses",
	"fish":        "fish",
	"mouse":       "mice",
	"query":       "queries",
	"ability":     "abilities",
	"agency":      "agencies",
	"movie":       "movies",
	"archive":     "archives",
	"index":       "indices",
	"wife":        "wives",
	"safe":        "saves",
	"half":        "halves",
	"move":        "moves",
	"salesperson": "salespeople",
	"person":      "people",
	"spokesman":   "spokesmen",
	"man":         "men",
	"woman":       "women",
	"basis":       "bases",
	"diagnosis":   "diagnoses",
	"diagnosis_a": "diagnosis_as",
	"datum":       "data",
	"medium":      "media",
	"stadium":     "stadia",
	"analysis":    "analyses",
	"node_child":  "node_children",
	"child":       "children",
	"experience":  "experiences",
	"day":         "days",
	"comment":     "comments",
	"foobar":      "foobars",
	"newsletter":  "newsletters",
	"old_news":    "old_news",
	"news":        "news",
	"series":      "series",
	"species":     "species",
	"quiz":        "quizzes",
	"perspective": "perspectives",
	"ox":          "oxen",
	"photo":       "photos",
	"buffalo":     "buffaloes",
	"tomato":      "tomatoes",
	"dwarf":       "dwarves",
	"elf":         "elves",
	"information": "information",
	"equipment":   "equipment",
	"criterion":   "criteria",
}

// storage is used to restore the state of the global variables
// on each test execution, to ensure no global state pollution
type storage struct {
	singulars    RegularSlice
	plurals      RegularSlice
	irregulars   IrregularSlice
	uncountables []string
}

var backup = storage{}

func init() {
	AddIrregular("criterion", "criteria")
	copy(backup.singulars, singularInflections)
	copy(backup.plurals, pluralInflections)
	copy(backup.irregulars, irregularInflections)
	copy(backup.uncountables, uncountableInflections)
}

func restore() {
	copy(singularInflections, backup.singulars)
	copy(pluralInflections, backup.plurals)
	copy(irregularInflections, backup.irregulars)
	copy(uncountableInflections, backup.uncountables)
}

func TestPlural(t *testing.T) {
	for key, value := range inflections {
		if v := Plural(strings.ToUpper(key)); v != strings.ToUpper(value) {
			t.Errorf("%v's plural should be %v, but got %v", strings.ToUpper(key), strings.ToUpper(value), v)
		}

		if v := Plural(strings.Title(key)); v != strings.Title(value) {
			t.Errorf("%v's plural should be %v, but got %v", strings.Title(key), strings.Title(value), v)
		}

		if v := Plural(key); v != value {
			t.Errorf("%v's plural should be %v, but got %v", key, value, v)
		}
	}
}

func TestSingular(t *testing.T) {
	for key, value := range inflections {
		if v := Singular(strings.ToUpper(value)); v != strings.ToUpper(key) {
			t.Errorf("%v's singular should be %v, but got %v", strings.ToUpper(value), strings.ToUpper(key), v)
		}

		if v := Singular(strings.Title(value)); v != strings.Title(key) {
			t.Errorf("%v's singular should be %v, but got %v", strings.Title(value), strings.Title(key), v)
		}

		if v := Singular(value); v != key {
			t.Errorf("%v's singular should be %v, but got %v", value, key, v)
		}
	}
}

func TestAddPlural(t *testing.T) {
	defer restore()
	ln := len(pluralInflections)
	AddPlural("", "")
	if ln+1 != len(pluralInflections) {
		t.Errorf("Expected len %d, got %d", ln+1, len(pluralInflections))
	}
}

func TestAddSingular(t *testing.T) {
	defer restore()
	ln := len(singularInflections)
	AddSingular("", "")
	if ln+1 != len(singularInflections) {
		t.Errorf("Expected len %d, got %d", ln+1, len(singularInflections))
	}
}

func TestAddIrregular(t *testing.T) {
	defer restore()
	ln := len(irregularInflections)
	AddIrregular("", "")
	if ln+1 != len(irregularInflections) {
		t.Errorf("Expected len %d, got %d", ln+1, len(irregularInflections))
	}
}

func TestAddUncountable(t *testing.T) {
	defer restore()
	ln := len(uncountableInflections)
	AddUncountable("", "")
	if ln+2 != len(uncountableInflections) {
		t.Errorf("Expected len %d, got %d", ln+2, len(uncountableInflections))
	}
}

func TestGetPlural(t *testing.T) {
	plurals := GetPlural()
	if len(plurals) != len(pluralInflections) {
		t.Errorf("Expected len %d, got %d", len(plurals), len(pluralInflections))
	}
}

func TestGetSingular(t *testing.T) {
	singular := GetSingular()
	if len(singular) != len(singularInflections) {
		t.Errorf("Expected len %d, got %d", len(singular), len(singularInflections))
	}
}

func TestGetIrregular(t *testing.T) {
	irregular := GetIrregular()
	if len(irregular) != len(irregularInflections) {
		t.Errorf("Expected len %d, got %d", len(irregular), len(irregularInflections))
	}
}

func TestGetUncountable(t *testing.T) {
	uncountables := GetUncountable()
	if len(uncountables) != len(uncountableInflections) {
		t.Errorf("Expected len %d, got %d", len(uncountables), len(uncountableInflections))
	}
}

func TestSetPlural(t *testing.T) {
	defer restore()
	SetPlural(RegularSlice{{}, {}})
	if len(pluralInflections) != 2 {
		t.Errorf("Expected len 2, got %d", len(pluralInflections))
	}
}

func TestSetSingular(t *testing.T) {
	defer restore()
	SetSingular(RegularSlice{{}, {}})
	if len(singularInflections) != 2 {
		t.Errorf("Expected len 2, got %d", len(singularInflections))
	}
}

func TestSetIrregular(t *testing.T) {
	defer restore()
	SetIrregular(IrregularSlice{{}, {}})
	if len(irregularInflections) != 2 {
		t.Errorf("Expected len 2, got %d", len(irregularInflections))
	}
}

func TestSetUncountable(t *testing.T) {
	defer restore()
	SetUncountable([]string{"", ""})
	if len(uncountableInflections) != 2 {
		t.Errorf("Expected len 2, got %d", len(uncountableInflections))
	}
}
