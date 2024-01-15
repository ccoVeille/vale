package lint

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Comment represents an in-code comment (line or block).
type Comment struct {
	Text   string
	Line   int
	Offset int
	Scope  string
}

func getLanguageFromExt(ext string) (*sitter.Language, error) {
	switch ext {
	case ".go":
		return golang.GetLanguage(), nil
	case ".c", ".h":
		return c.GetLanguage(), nil
	case ".cpp", ".hpp", ".cc", ".hh", ".cxx", ".hxx":
		return cpp.GetLanguage(), nil
	case ".cs", ".csx":
		return csharp.GetLanguage(), nil
	case ".css":
		return css.GetLanguage(), nil
	case ".java", ".bsh":
		return java.GetLanguage(), nil
	case ".js":
		return javascript.GetLanguage(), nil
	case ".lua":
		return lua.GetLanguage(), nil
	case ".py", ".py3", ".pyw", ".pyi", ".pyx", ".rpy":
		return python.GetLanguage(), nil
	case ".rb":
		return ruby.GetLanguage(), nil
	case ".rs":
		return rust.GetLanguage(), nil
	case ".scala", ".sbt":
		return scala.GetLanguage(), nil
	case ".ts":
		return typescript.GetLanguage(), nil
	default:
		return nil, errors.New("unsupported extension")
	}

	// fallback: haskell, less, perl, php, powershell, r, sass, swift
}

func getComments(source []byte, lang *sitter.Language) ([]Comment, error) {
	var comments []Comment

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree := parser.Parse(nil, source)
	n := tree.RootNode()

	q, err := sitter.NewQuery([]byte("(comment)+ @comment"), lang)
	if err != nil {
		return comments, err
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(q, n)

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range m.Captures {
			children := int(c.Node.ChildCount())
			for i := 0; i < children; i++ {
				child := c.Node.Child(i)
				fmt.Println(child)
			}
			text := c.Node.Content(source)

			scope := "text.comment.line"
			if strings.Count(text, "\n") > 0 {
				scope = "text.comment.block"
			}

			comments = append(comments, Comment{
				Line:   int(c.Node.StartPoint().Row) + 1,
				Offset: int(c.Node.StartPoint().Column),
				Scope:  scope,
				Text:   text,
			})
		}
	}

	return comments, nil
}

func getSubMatch(r *regexp.Regexp, s string) string {
	matches := r.FindStringSubmatch(s)
	for i, m := range matches {
		if i > 0 && m != "" {
			return m
		}
	}
	return ""
}

func padding(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func doMatch(p []*regexp.Regexp, line string) string {
	for _, r := range p {
		if m := getSubMatch(r, line); m != "" {
			return m
		}
	}
	return ""
}
