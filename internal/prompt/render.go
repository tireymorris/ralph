package prompt

import (
	"bytes"
	"fmt"
)

func render(name string, data any) string {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		panic(fmt.Sprintf("prompt template %q: %v", name, err))
	}
	return buf.String()
}

func mustRender(name string, data any) string {
	return render(name, data)
}
