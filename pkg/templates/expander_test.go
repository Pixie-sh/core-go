package templates

import (
	"fmt"
	"net/url"
	"testing"
)

func TestExpandMap(t *testing.T) {

	fields := map[string]string{
		"deal_title":  "Title example #1",
		"deal_id":     fmt.Sprintf("%d", 12),
		"progress_id": fmt.Sprintf("%d", 34)}

	const htmlTemplate = `Creator approved for Deal '{{ .deal_title }}'`

	html, err := Expander{}.Expand(htmlTemplate, fields)
	if err != nil {
		fmt.Println("Error generating HTML:", err)
		return
	}

	fmt.Println(html)
}

func TestUnescapeHTML(t *testing.T) {

	escaped := "%7B%7B%23%3E%20en_PixieEmailTemplate%7D%7D%0A%0A%7B%7B%23*inline%20%22content%22%7D%7D%0A%3Ctr%3E%0A%20%20%20%20%3Ctd%20style%3D%22padding%3A%2020px%3B%20font-family%3A%20Helvetica%2C%20Arial%2C%20sans-serif%3B%20font-size%3A%2016px%3B%20color%3A%20%23333333%3B%22%3E%0A%20%20%20%20%20%20%20%20%3Cp%20style%3D%22margin%3A%200%200%2016px%3B%22%3EHi%20%7B%7Buser_first_name%7D%7D%2C%3C%2Fp%3E%0A%20%20%20%20%20%20%20%20%3Cp%20style%3D%22margin%3A%200%200%2016px%3B%22%3EWe%20received%20a%20request%20to%20reset%20your%20password.%20Click%20the%20button%20below%20to%20reset%20it%3A%3C%2Fp%3E%0A%20%20%20%20%3C%2Ftd%3E%0A%3C%2Ftr%3E%0A%3C!--%20BUTTON%20--%3E%0A%3Ctr%3E%0A%20%20%20%20%3Ctd%20align%3D%22center%22%20style%3D%22padding%3A%200%2020px%2030px%2020px%3B%22%3E%0A%20%20%20%20%20%20%20%20%3Ctable%20role%3D%22presentation%22%20border%3D%220%22%20cellspacing%3D%220%22%20cellpadding%3D%220%22%3E%0A%20%20%20%20%20%20%20%20%20%20%20%20%3Ctr%3E%0A%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%3Ctd%20align%3D%22center%22%20bgcolor%3D%22%230867ec%22%20style%3D%22border-radius%3A%205px%3B%20background-color%3A%20%230867ec%3B%22%3E%0A%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%3Ca%20href%3D%22%7B%7Breset_link%7D%7D%22%20target%3D%22_blank%22%20style%3D%22font-size%3A%2016px%3B%20font-family%3A%20Helvetica%2C%20Arial%2C%20sans-serif%3B%20color%3A%20%23ffffff%3B%20text-decoration%3A%20none%3B%20padding%3A%2012px%2024px%3B%20display%3A%20inline-block%3B%20border%3A%202px%20solid%20%230867ec%3B%20border-radius%3A%205px%3B%22%3EReset%20Password%3C%2Fa%3E%0A%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%3C%2Ftd%3E%0A%20%20%20%20%20%20%20%20%20%20%20%20%3C%2Ftr%3E%0A%20%20%20%20%20%20%20%20%3C%2Ftable%3E%0A%20%20%20%20%3C%2Ftd%3E%0A%3C%2Ftr%3E%0A%3C!--%20FALLBACK%20LINK%20--%3E%0A%3Ctr%3E%0A%20%20%20%20%3Ctd%20style%3D%22padding%3A%200%2020px%2020px%2020px%3B%20font-family%3A%20Helvetica%2C%20Arial%2C%20sans-serif%3B%20font-size%3A%2016px%3B%20color%3A%20%23333333%3B%22%3E%0A%20%20%20%20%20%20%20%20%3Cp%20style%3D%22margin%3A%200%200%2016px%3B%22%3EIf%20the%20button%20above%20doesn't%20work%2C%20copy%20and%20paste%20this%20link%20into%20your%20web%20browser%3A%3C%2Fp%3E%0A%20%20%20%20%20%20%20%20%3Cp%20style%3D%22margin%3A%200%3B%20word-break%3A%20break-all%3B%22%3E%3Ca%20href%3D%22%7B%7Breset_link%7D%7D%22%20style%3D%22color%3A%20%230867ec%3B%20text-decoration%3A%20underline%3B%22%3E%7B%7Breset_link%7D%7D%3C%2Fa%3E%3C%2Fp%3E%0A%20%20%20%20%3C%2Ftd%3E%0A%3C%2Ftr%3E%0A%7B%7B%2Finline%7D%7D%0A%0A%7B%7B%2Fen_PixieEmailTemplate%7D%7D"
	unescaped, err := url.QueryUnescape(escaped)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(unescaped)
	}

}
