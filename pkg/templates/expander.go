package templates

import (
	"bytes"
	htmlTmpl "html/template"
	"regexp"

	"github.com/mailgun/raymond/v2"
	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/types/strings"
)

type Expander struct {
	layoutByName func(string) (string, error)
}

// NewExpander this is just a wrapper for our expander, raymond (handlebars) implementation is global: TBI soon.
func NewExpander(getLayoutByName ...func(string) (string, error)) Expander {
	var layoutByName func(string) (string, error) = nil
	if len(getLayoutByName) > 0 && getLayoutByName[0] != nil {
		layoutByName = getLayoutByName[0]
	}

	return Expander{
		layoutByName: layoutByName,
	}
}

// Expand is the original method for simple template expansion
// example:
//
//	Fields: map[string]string{
//				"deal_title":  getEvent.Title,
//				"deal_id":     fmt.Sprintf("%d", getEvent.ID),
//				"progress_id": fmt.Sprintf("%d", getEvent.ProgressID)},
//
// Template: "Creator approved for Deal '{{ .deal_title }}'",
func (e Expander) Expand(template string, values any) (string, error) {
	tmpl, err := htmlTmpl.New("tmpl").Parse(template)
	if err != nil {
		return "", err
	}

	var htmlOutput bytes.Buffer
	err = tmpl.Execute(&htmlOutput, values)
	if err != nil {
		return "", err
	}

	return htmlOutput.String(), nil
}

// ExpandLayout is a new method for handlebars like template expansion
// example: check TestReplaceWithExpander and TestSimpleReplaceWithLayout
func (e Expander) ExpandLayout(tmpl string, event any) (string, error) {
	unescapedTmpl, err := strings.Unescape(tmpl)
	if err != nil {
		return "", err
	}

	layoutName, inline, err := extractLayoutAndContent(unescapedTmpl)
	if err != nil {
		return raymond.Render(unescapedTmpl, event)
	}

	layoutTemplate := unescapedTmpl
	if len(layoutName) > 0 && e.layoutByName != nil {
		layoutTemplate, err = e.layoutByName(layoutName)
		if err != nil {
			return "", err
		}

		layoutTemplate, err = strings.Unescape(layoutTemplate)
		if err != nil {
			return "", err
		}

		inlineKey, inlineValue, err := extractInlineContent(inline)
		if err != nil {
			return "", err
		}

		SetPartial(inlineKey, inlineValue)
	}

	result, err := raymond.Render(layoutTemplate, event)
	if err != nil {
		return "", err
	}

	return result, nil
}

func extractLayoutAndContent(input string) (string, string, error) {
	layoutRegex := regexp.MustCompile(`(?s){{#>\s*(\w+)}}(.*?{{/inline}}.*?){{/\w+}}`)
	match := layoutRegex.FindStringSubmatch(input)

	if match == nil {
		return "", "", errors.New("no match found")
	}

	if len(match) < 3 {
		return "", "", errors.New("invalid match: expected 3 groups, got %d groups", len(match))
	}

	// Verify that the closing tag matches the opening tag
	closingTag := regexp.MustCompile(`{{/(\w+)}}\s*$`)
	closingMatch := closingTag.FindStringSubmatch(input)
	if len(closingMatch) < 2 || closingMatch[1] != match[1] {
		return "", "", errors.New("closing tag does not match opening tag")
	}

	return match[1], strings.TrimSpace(match[2]), nil
}

func extractInlineContent(input string) (string, string, error) {
	re := regexp.MustCompile(`{{#\*inline\s*"(\w+)"}}([\s\S]*?){{/inline}}`)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 3 {
		return "", "", errors.New("no match found or incorrect number of capture groups")
	}

	name := matches[1]
	content := strings.TrimSpace(matches[2])
	return name, content, nil
}
