package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/errata-ai/vale/v2/internal/glob"
)

// A Vocabulary represents a set of accepted and rejected tokens.
type Vocabulary struct {
	pattern        glob.Glob
	acceptedTokens map[string]struct{}
	rejectedTokens map[string]struct{}
}

func NewVocabulary(section string) (*Vocabulary, error) {
	compiled, err := glob.Compile(section)
	if err != nil {
		return nil, err
	}
	return &Vocabulary{
		pattern:        compiled,
		acceptedTokens: make(map[string]struct{}),
		rejectedTokens: make(map[string]struct{}),
	}, nil
}

func (c *Vocabulary) Matches(fp string) bool {
	return c.pattern.Match(fp)
}

// AddWordListFile adds vocab terms from a provided file.
func (c *Vocabulary) AddWordListFile(name string, accept bool) error {
	fd, err := os.Open(name)
	if err != nil {
		return err
	}
	defer fd.Close()
	return c.addWordList(fd, accept)
}

func (c *Vocabulary) addWordList(r io.Reader, accept bool) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if len(word) == 0 || strings.HasPrefix(word, "# ") { //nolint:gocritic
			continue
		} else if accept {
			if _, ok := c.acceptedTokens[word]; !ok {
				c.acceptedTokens[word] = struct{}{}
			}
		} else {
			if _, ok := c.rejectedTokens[word]; !ok {
				c.rejectedTokens[word] = struct{}{}
			}
		}
	}
	return scanner.Err()
}

func loadVocab(label string, names []string, cfg *Config) (*Vocabulary, error) {
	vocab, err := NewVocabulary(label)
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		target := ""

		for _, p := range cfg.Paths {
			opt := filepath.Join(p, VocabDir, name)
			if IsDir(opt) {
				target = opt
				break
			}
		}

		if target == "" {
			return nil, NewE100("vocab", fmt.Errorf(
				"'%s/%s' directory does not exist", VocabDir, name))
		}

		accepted := filepath.Join(target, "accept.txt")
		if FileExists(accepted) {
			if err := vocab.AddWordListFile(accepted, true); err != nil {
				return nil, NewE100("vocab", err)
			}
		}

		rejected := filepath.Join(target, "reject.txt")
		if FileExists(rejected) {
			if err := vocab.AddWordListFile(rejected, false); err != nil {
				return nil, NewE100("vocab", err)
			}
		}
	}

	return vocab, nil
}
