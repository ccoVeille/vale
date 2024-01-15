package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func toMarkup(comments []Comment) string {
	var markup bytes.Buffer

	for _, comment := range comments {
		markup.WriteString(strings.TrimLeft(comment.Text, " "))
	}

	return markup.String()
}

func TestParse(t *testing.T) {
	source, err := os.ReadFile("../../testdata/comments/in/3.go")
	if err != nil {
		t.Fatal(err)
	}

	lang, err := getLanguageFromExt(".go")
	if err != nil {
		t.Fatal(err)
	}

	comments, err := getComments(source, lang)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(comments)
}

func TestComments(t *testing.T) {
	cases, err := os.ReadDir("../../testdata/comments/in")
	if err != nil {
		t.Fatal(err)
	}

	for i, f := range cases {
		b, err1 := os.ReadFile(fmt.Sprintf("../../testdata/comments/in/%s", f.Name()))
		if err1 != nil {
			t.Fatal(err1)
		}

		lang, err2 := getLanguageFromExt(filepath.Ext(f.Name()))
		if err2 != nil {
			t.Fatal(err2)
		}

		comments, err3 := getComments(b, lang)
		if err3 != nil {
			t.Fatal(err3)
		}

		b2, err4 := os.ReadFile(fmt.Sprintf("../../testdata/comments/out/%d.txt", i))
		if err4 != nil {
			t.Fatal(err4)
		}
		markup := toMarkup(comments)

		if markup != string(b2) {
			err = os.WriteFile(fmt.Sprintf("%d.txt", i), []byte(markup), os.ModePerm)
			if err != nil {
				t.Error(err)
			}
			t.Errorf("%s", markup)
		}
	}
}
