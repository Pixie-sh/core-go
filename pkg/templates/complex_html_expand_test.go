package templates

import (
	"fmt"
	"testing"
	"time"

	"github.com/mailgun/raymond/v2"
)

type inner struct {
	Name string    `json:"name"`
	Data time.Time `json:"data"`
}
type list struct {
	Name  string `json:"name"`
	Age   uint64 `json:"age"`
	Inner inner  `json:"inner"`
}

type nesting struct {
	List []list `json:"nest_list"`
}

const complexEventExample = "{\"id\":\"partner_pending_applications_reminder_event-127\",\"timestamp\":\"2025-05-02T18:25:53.982257167Z\",\"headers\":{\"admin.user_id\":\"2001:818:d9ff:3000:c01b:e653:b7c9:6460\",\"admin.user_ip\":\"2001:818:d9ff:3000:c01b:e653:b7c9:6460\"},\"from_sender_id\":\"admin-1\",\"payload_type\":\"partner_pending_applications_reminder_event\",\"payload\":{\"deal_applications\":[{\"creator_applications\":[{\"applications\":[{\"application_date\":\"15-06-2023 14:30 GMT\",\"application_id\":\"01JRQ96NXX03468TP466VBEXM0\"}],\"creator_id\":\"123\",\"creator_image\":\"https://example.com/creators/987654321/profile.jpg\",\"creator_name\":\"Fitness Influencer\",\"socials\":[{\"followers\":\"500K\",\"icon\":\"https://example.com/icons/instagram.png\",\"name\":\"Instagram\"},{\"followers\":\"1.2M\",\"icon\":\"https://example.com/icons/tiktok.png\",\"name\":\"TikTok\"}]},{\"applications\":[{\"application_date\":\"16-06-2023 09:15 GMT\",\"application_id\":\"01JRQ96NXX03468TP466VBEXM0\"}],\"creator_id\":\"123\",\"creator_image\":\"https://example.com/creators/543216789/profile.jpg\",\"creator_name\":\"Health Coach\",\"socials\":[{\"followers\":\"250K\",\"icon\":\"https://example.com/icons/youtube.png\",\"name\":\"YouTube\"}]}],\"deal_id\":\"01JRQ96NXX03468TP466VBEXM0\",\"deal_image\":\"https://example.com/deals/123456789/image.jpg\",\"deal_status\":\"live\",\"deal_title\":\"Summer Fitness Campaign\"},{\"creator_applications\":[{\"applications\":[{\"application_date\":\"10-06-2023 11:45 GMT\",\"application_id\":\"01JRQ96NXX03468TP466VBEXM0\"},{\"application_date\":\"12-06-2023 16:20 GMT\",\"application_id\":\"01JRQ96NXX03468TP466VBEXM0\"}],\"creator_id\":\"1234\",\"creator_image\":\"https://example.com/creators/123789456/profile.jpg\",\"creator_name\":\"Eco Influencer\",\"socials\":[{\"followers\":\"320K\",\"icon\":\"https://example.com/icons/instagram.png\",\"name\":\"Instagram\"},{\"followers\":\"150K\",\"icon\":\"https://example.com/icons/twitter.png\",\"name\":\"Twitter\"}]}],\"deal_id\":\"01JRQ96NXX03468TP466VBEXM0\",\"deal_image\":\"https://example.com/deals/987654321/image.jpg\",\"deal_status\":\"live\",\"deal_title\":\"Green Living Initiative\"}],\"partner_id\":\"partner-123\"}"

func TestReplaceLayoutWithEventMockData(t *testing.T) {
	templat := `{{#> LayoutEmailTemplate}}
{{#*inline "content"}}
		<p>Partner ID: {{partner_id}}</p>
       {{#each deal_applications}}
           <p>Available keys: {{#each this}}{{@key}}, {{/each}}</p>
           <p>Deal title should be: {{deal_title}}</p>
       {{/each}}
{{/inline}}
{{/LayoutEmailTemplate}}`

	layoutName, inline, _ := extractLayoutAndContent(templat)
	layoutTemplate, _ := getLayoutByName(layoutName)

	inlineKey, inlineValue, err := extractInlineContent(inline)
	if err != nil {
		t.Fatalf("Failed to extract inline: %v", err)
	}

	raymond.RegisterPartial(inlineKey, inlineValue)

	// Use mock data instead of message factory
	mockData := map[string]interface{}{
		"partner_id": "partner-123",
		"deal_applications": []map[string]interface{}{
			{
				"deal_title": "Summer Fitness Campaign",
				"deal_id":    "01JRQ96NXX03468TP466VBEXM0",
			},
		},
	}

	result, err := raymond.Render(layoutTemplate, mockData)
	if err != nil {
		t.Fatalf("Error rendering layout: %v", err)
	}

	fmt.Println(result)
}
func getLayoutByName(name string) (string, error) {
	fmt.Println("layout names: ", name)
	return `
<!doctype html>
<html>
<head>
<title>{{fnName title}}</title>
</head>
<body>
    {{> content}}
</body>
</html>`, nil
}

func TestReplaceWithExpander(t *testing.T) {
	expander := NewExpander(getLayoutByName)
	inputText :=
		`{{#> LayoutEmailTemplate}}
			{{#*inline "content"}}
				{{#each items}}
					<p>{{this}}</p>
				{{/each}}
			{{/inline}}
		{{/LayoutEmailTemplate}}`

	event := map[string]interface{}{
		"items": []string{"Item 1", "Item 2", "Item 3"},
		"title": uint64(6969),
	}

	RegisterHelpers(GlobalRegisterHelper{FunctionName: "fnName", UsageExample: "<title>{{fnName title}}</title>", AnonymousFunction: func(coiso uint64) string {
		return fmt.Sprintf("%d%d", coiso, coiso)
	}})
	result, err := expander.ExpandLayout(inputText, event)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(result)
}
